package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
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
	Score    int
	Nodes    int
	BestMove engine.Move
	PV       []engine.Move
}

// SearchOptions holds configuration for the search
type SearchOptions struct {
	MaxDepth int
}

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
// Returns the best move, the score, and the number of nodes visited.
// infoCh is a channel to report search progress (can be nil).
func Search(ctx context.Context, game *engine.Game, options SearchOptions, infoCh chan<- SearchInfo) (engine.Move, int, int) {
	var bestMove engine.Move
	var score int
	totalNodes := 0

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

	// Iterative Deepening
	for depth := 1; depth <= options.MaxDepth; depth++ {
		var selDepth int
		// We use a new node counter for each iteration to track work done
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
				Nodes:    totalNodes,
				BestMove: bestMove,
				PV:       getPV(game, bestMove, depth),
			}
		}

		if ctx.Err() != nil {
			break
		}
	}
	return bestMove, score, totalNodes
}

// getPV extracts the Principal Variation (best line of play) from the TT
func getPV(game *engine.Game, firstMove engine.Move, depth int) []engine.Move {
	pv := []engine.Move{firstMove}
	movesMade := 0

	// We must execute moves to update the hash and probe the TT for the next move
	// We will undo them all at the end to restore game state
	if !game.ExecuteMove(firstMove) {
		return pv
	}
	movesMade++

	for i := 0; i < depth-1; i++ {
		entry, ok := GlobalTT.Probe(game.ZobristHash)
		if !ok || entry.BestMove == (engine.Move{}) {
			break
		}
		// In a real engine, we should verify legality here
		if !game.ExecuteMove(entry.BestMove) {
			break
		}
		pv = append(pv, entry.BestMove)
		movesMade++
	}

	for i := 0; i < movesMade; i++ {
		game.UnmakeMove()
	}
	return pv
}
