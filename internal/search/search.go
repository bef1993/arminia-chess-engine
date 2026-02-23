package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// MaxPly is the maximum search depth we support for arrays indexed by ply
const MaxPly = 64

// scoreToTT converts a search score to a TT score (independent of ply)
func scoreToTT(score, ply int) int {
	if score > MateBound {
		return score + ply
	}
	if score < -MateBound {
		return score - ply
	}
	return score
}

// scoreFromTT converts a TT score to a search score (relative to ply)
func scoreFromTT(score, ply int) int {
	if score > MateBound {
		return score - ply
	}
	if score < -MateBound {
		return score + ply
	}
	return score
}

// SearchInfo holds progress information sent via channel
type SearchInfo struct {
	Depth    int
	SelDepth int // Maximum depth reached (including quiescence extensions)
	Score    int // Score in centipawns, relative to the side to move
	Nodes    int64
	BestMove engine.Move
	PV       []engine.Move
}

// SearchOptions holds configuration for the search
type SearchOptions struct {
	MaxDepth int
	Threads  int
}

// SearchContext holds thread-local and shared state for the search
type SearchContext struct {
	Ctx         context.Context
	Game        *engine.Game
	Nodes       *atomic.Int64           // Shared node counter
	SelDepth    *int                    // Thread-local max depth
	Rand        *rand.Rand              // Thread-local random number generator (nil for main thread)
	KillerMoves *[MaxPly][2]engine.Move // Thread-local killer moves table
	History     *HistoryTable           // Shared history heuristic table
}

func NewSearchContext(game *engine.Game) *SearchContext {
	return &SearchContext{
		Ctx:         context.Background(),
		Game:        game,
		Nodes:       &atomic.Int64{},
		SelDepth:    new(int),
		Rand:        nil,
		KillerMoves: &[MaxPly][2]engine.Move{},
		History:     &HistoryTable{},
	}
}

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
// Returns the best move, the score, and the number of nodes visited.
// infoCh is a channel to report search progress (can be nil).
func Search(ctx context.Context, game *engine.Game, options SearchOptions, infoCh chan<- SearchInfo) (engine.Move, int, int64) {
	totalNodes := atomic.Int64{}
	var wg sync.WaitGroup

	// Spawn helper threads for Lazy SMP to populate the TT
	runLazySMPHelpers(ctx, &wg, game, options, &totalNodes)

	// Run the main iterative deepening search
	bestMove, score := runIterativeDeepening(ctx, game, options, &totalNodes, infoCh)

	// Wait for helper threads to finish
	wg.Wait()

	return bestMove, score, totalNodes.Load()
}

// runLazySMPHelpers spawns helper threads to populate the TT.
func runLazySMPHelpers(ctx context.Context, wg *sync.WaitGroup, game *engine.Game, options SearchOptions, totalNodes *atomic.Int64) {
	numThreads := options.Threads
	if numThreads < 1 {
		numThreads = 1
	}

	for i := 1; i < numThreads; i++ {
		threadId := i
		gameClone := game.Clone()
		wg.Go(func() {
			var helperSelDepth int
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadId)))
			killerMoves := [MaxPly][2]engine.Move{}
			history := HistoryTable{}
			sc := &SearchContext{
				Ctx:         ctx,
				Game:        gameClone,
				Nodes:       totalNodes,
				SelDepth:    &helperSelDepth,
				Rand:        rng,
				KillerMoves: &killerMoves,
				History:     &history,
			}

			// Helpers run Iterative Deepening to populate TT.
			for d := 1; d <= options.MaxDepth; d++ {
				_, _, interrupted := negamax(sc, d, -EvalInfinity, EvalInfinity, 0, 0, true)
				if interrupted || ctx.Err() != nil {
					break
				}
			}
		})
	}
}

// runIterativeDeepening performs the main search loop.
func runIterativeDeepening(ctx context.Context, game *engine.Game, options SearchOptions, totalNodes *atomic.Int64, infoCh chan<- SearchInfo) (engine.Move, int) {
	var bestMove engine.Move
	var score int

	bestMove = InitializeBestMoveFallback(game)

	// Killer moves are maintained across iterative deepening depths to improve move ordering.
	// They are not reset after each depth iteration.
	killerMoves := [MaxPly][2]engine.Move{}
	history := HistoryTable{}
	for depth := 1; depth <= options.MaxDepth; depth++ {
		var selDepth int
		sc := &SearchContext{
			Ctx:         ctx,
			Game:        game,
			Nodes:       totalNodes,
			SelDepth:    &selDepth,
			Rand:        nil, // Main thread is deterministic
			KillerMoves: &killerMoves,
			History:     &history,
		}

		eval, move, interrupted := negamax(sc, depth, -EvalInfinity, EvalInfinity, 0, 0, true)

		if interrupted {
			break
		}

		if move != (engine.Move{}) {
			bestMove = move
		}
		score = eval

		if infoCh != nil {
			infoCh <- SearchInfo{
				Depth:    depth,
				SelDepth: selDepth,
				Score:    score,
				Nodes:    totalNodes.Load(),
				BestMove: bestMove,
				PV:       getPV(game, bestMove),
			}
		}

		if ctx.Err() != nil {
			break
		}
	}
	return bestMove, score
}

func InitializeBestMoveFallback(game *engine.Game) (bestMove engine.Move) {
	// In case the search is interrupted before depth 1 completes, we need a fallback move.
	// We simply take the first legal move available. This ensures we always return a valid
	// move, preventing a forfeit in time-critical situations.
	legalMoves := game.GenerateLegalMoves()
	if len(legalMoves) > 0 {
		bestMove = legalMoves[0]
	}
	return bestMove
}

// getPV extracts the Principal Variation (best line of play) from the TT
func getPV(game *engine.Game, firstMove engine.Move) []engine.Move {
	pv := []engine.Move{firstMove}

	// Use a clone to avoid modifying the search state.
	// This prevents any risk of corrupting the main search loop's board state.
	simGame := game.Clone()

	if !simGame.ExecuteMove(firstMove) {
		return pv
	}

	// Follow PV up to a reasonable limit (e.g. 64) to include QS moves or extensions
	for i := 0; i < 64; i++ {
		entry, ok := GlobalTT.Probe(simGame.ZobristHash)
		if !ok || entry.BestMove == (engine.Move{}) {
			break
		}

		// Verify legality of the TT move
		// This handles hash collisions where the move might be valid for a different position
		// but illegal (or nonsense) for the current one.
		legalMoves := simGame.GenerateLegalMoves()
		isLegal := false
		for _, m := range legalMoves {
			if m == entry.BestMove {
				isLegal = true
				break
			}
		}
		if !isLegal {
			break
		}

		if !simGame.ExecuteMove(entry.BestMove) {
			break
		}
		pv = append(pv, entry.BestMove)
	}

	return pv
}
