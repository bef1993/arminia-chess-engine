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

// Piece-Square Tables (PST)
// The PST are defined in human readable format (rank 1 at the bottom) and are indexed by square.
// This means for White we mirror the square (sq ^ 56) to flip the rank.

var pawnPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightPST = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopPST = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenPST = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingPST = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

// Evaluate calculates the score of the current board position from the perspective
// of the current player. A positive score means the current player has an advantage.
func Evaluate(game *engine.Game) int {
	score := evaluatePosition(game.Board)

	// Convert the absolute score to a perspective-based score.
	if game.CurrentTurn == engine.White {
		return score
	}
	return -score
}

// evaluatePosition calculates the score based on material and piece-square tables.
// Positive values indicate an advantage for White,
// negative values indicate an advantage for Black.
func evaluatePosition(board engine.Board) int {
	score := 0
	score += evaluatePiece(board, engine.Pawn, &pawnPST, engine.PawnValue)
	score += evaluatePiece(board, engine.Knight, &knightPST, engine.KnightValue)
	score += evaluatePiece(board, engine.Bishop, &bishopPST, engine.BishopValue)
	score += evaluatePiece(board, engine.Rook, &rookPST, engine.RookValue)
	score += evaluatePiece(board, engine.Queen, &queenPST, engine.QueenValue)
	score += evaluatePiece(board, engine.King, &kingPST, engine.KingValue)
	return score
}

func evaluatePiece(board engine.Board, pieceType engine.PieceType, table *[64]int, val int) int {
	score := 0
	// White
	bb := board.Pieces[engine.White][pieceType]
	for bb != 0 {
		sq := bb.PopLSB()
		// Mirror square for black (flip rank: sq ^ 56)
		score += val + table[sq^56]
	}
	// Black
	bb = board.Pieces[engine.Black][pieceType]
	for bb != 0 {
		sq := bb.PopLSB()
		score -= val + table[sq]
	}
	return score
}
