package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"sync"
	"sync/atomic"
)

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

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
// Returns the best move, the score, and the number of nodes visited.
// infoCh is a channel to report search progress (can be nil).
func Search(ctx context.Context, game *engine.Game, options SearchOptions, infoCh chan<- SearchInfo) (engine.Move, int, int64) {
	var bestMove engine.Move
	var score int
	var totalNodes atomic.Int64

	bestMove = InitializeBestMoveFallback(game)

	// Configure threads
	numThreads := 1 //TODO use options.Threads
	if numThreads < 1 {
		numThreads = 1
	}

	var wg sync.WaitGroup

	// Spawn helper threads (Lazy SMP)
	// Helpers search the same position to populate the TT.
	for i := 1; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each thread needs its own game instance to avoid race conditions on board state
			gameClone := game.Clone()
			var helperSelDepth int
			// Helpers search to the max depth; the context handles termination.
			negamax(ctx, gameClone, options.MaxDepth, -EvalInfinity, EvalInfinity, 0, &totalNodes, &helperSelDepth)
		}()
	}

	// Iterative Deepening
	for depth := 1; depth <= options.MaxDepth; depth++ {
		var selDepth int
		// The main thread performs the iterative deepening search.
		// It shares the totalNodes counter and the TT with the helper threads.
		eval, move, interrupted := negamax(ctx, game, depth, -EvalInfinity, EvalInfinity, 0, &totalNodes, &selDepth)

		if interrupted {
			break
		}

		bestMove = move
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

	// Wait for all helper threads to finish.
	// This is important because the context might be cancelled (e.g., by timeout),
	// and we need to ensure they have exited before this function returns.
	wg.Wait()
	return bestMove, score, totalNodes.Load()
}

func InitializeBestMoveFallback(game *engine.Game) (bestMove engine.Move) {
	legalMoves := game.GenerateLegalMoves()
	// Initialize bestMove with a fallback in case the search is interrupted before depth 1 completes.
	// 1. Try to retrieve a move from the Transposition Table (e.g. from previous search)
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok && entry.BestMove != (engine.Move{}) {
		// Validate TT move to ensure it's legal (guards against hash collisions)
		for _, m := range legalMoves {
			if m == entry.BestMove {
				bestMove = entry.BestMove
				break
			}
		}
	}

	// 2. If no TT move, use the first legal move
	// in case the search is interrupted before depth 1 completes.

	if bestMove == (engine.Move{}) {
		if len(legalMoves) > 0 {
			bestMove = legalMoves[0]
		}
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
