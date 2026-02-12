package search

import "arminia-chess-engine/internal/engine"

// Evaluate calculates the score of the current board position from the perspective
// of the current player. A positive score means the current player has an advantage.
func Evaluate(game *engine.Game) int {
	// This is an absolute score from White's perspective.
	// Positive means White is winning, negative means Black is winning.
	score := 0

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := game.Board.GetPiece(col, row)
			if piece == engine.NoPiece {
				continue
			}

			if piece.Color() == engine.White {
				score += piece.Value()
			} else {
				score -= piece.Value()
			}
		}
	}

	// Convert the absolute score to a perspective-based score.
	if game.CurrentTurn == engine.Black {
		return -score
	}
	return score
}
