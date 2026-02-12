package search

import "arminia-chess-engine/internal/engine"

// Evaluate calculates the score of the current board position from the perspective
// of the current player. A positive score means the current player has an advantage.
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
func countMaterial(board *engine.Board) (score int) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := board.GetPiece(col, row)
			if piece.Color() == engine.White {
				score += piece.Value()
			} else {
				score -= piece.Value()
			}

		}
	}
	return score
}
