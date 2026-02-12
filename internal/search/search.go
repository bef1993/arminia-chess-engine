package search

import "arminia-chess-engine/internal/engine"

const (
	// Infinity is a score that is higher than any possible evaluation.
	// Used for alpha-beta pruning.
	Infinity = 30000

	// Mate is a score indicating a checkmate. It's slightly less than infinity
	// to allow for distinguishing between mates at different depths.
	Mate = 29000
)

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
func Search(game *engine.Game, depth int) engine.Move {
	// TODO: Implement Negamax with alpha-beta pruning.

	// For now, as a placeholder, we just return the first legal move.
	legalMoves := game.GetLegalMoves()
	if len(legalMoves) == 0 {
		return engine.Move{} // No legal moves, return a zero-value move.
	}

	return legalMoves[0]
}
