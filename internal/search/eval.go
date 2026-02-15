package search

import "arminia-chess-engine/internal/engine"

const (
	// Infinity is a score that is higher than any possible evaluation.
	// Used for alpha-beta pruning.
	EvalInfinity = 30000

	// Mate is a score indicating a checkmate. It's slightly less than infinity
	// to allow for distinguishing between mates at different depths.
	EvalMate = 29000

	// MateBound is the threshold for considering a score a mate score
	MateBound = EvalMate - 1000
)

// Evaluate calculates the score of the current board position from the perspective
// of the current player. A positive score means the current player has an advantage.
// TODO : add more evaluation factors (piece-square tables, mobility, king safety, etc.)
func Evaluate(game *engine.Game) int {

	score := countMaterial(game.Board)

	// Convert the absolute score to a perspective-based score.
	if game.CurrentTurn == engine.White {
		return score
	}
	return -score
}

// Counts the material on the board and returns a score.
// Positive values indicate an advantage for White,
// negative values indicate an advantage for Black.
func countMaterial(board *engine.Board) int {
	score := 0
	score += (board.Pieces[engine.White][engine.Pawn].Count() - board.Pieces[engine.Black][engine.Pawn].Count()) * engine.PawnValue
	score += (board.Pieces[engine.White][engine.Knight].Count() - board.Pieces[engine.Black][engine.Knight].Count()) * engine.KnightValue
	score += (board.Pieces[engine.White][engine.Bishop].Count() - board.Pieces[engine.Black][engine.Bishop].Count()) * engine.BishopValue
	score += (board.Pieces[engine.White][engine.Rook].Count() - board.Pieces[engine.Black][engine.Rook].Count()) * engine.RookValue
	score += (board.Pieces[engine.White][engine.Queen].Count() - board.Pieces[engine.Black][engine.Queen].Count()) * engine.QueenValue
	score += (board.Pieces[engine.White][engine.King].Count() - board.Pieces[engine.Black][engine.King].Count()) * engine.KingValue
	return score
}
