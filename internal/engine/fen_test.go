package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFEN(t *testing.T) {
	game := NewGame()

	// Test starting position FEN
	startFen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	err := game.LoadFEN(startFen)
	assert.NoError(t, err, "Failed to load start FEN")

	assert.Equal(t, White, game.CurrentTurn, "Expected White turn")
	assert.Equal(t, AllCastling, game.CastlingRights, "Expected AllCastling")

	// Test a custom position
	// White King on e1, Black King on e8, White Rook on a1
	customFen := "4k3/8/8/8/8/8/8/R3K3 w Q - 0 1"
	err = game.LoadFEN(customFen)
	assert.NoError(t, err, "Failed to load custom FEN")

	assert.Equal(t, WhiteKing, game.Board.GetPiece(FileE, Rank1), "Expected WhiteKing at e1")
	assert.Equal(t, BlackKing, game.Board.GetPiece(FileE, Rank8), "Expected BlackKing at e8")
	assert.Equal(t, WhiteRook, game.Board.GetPiece(FileA, Rank1), "Expected WhiteRook at a1")
	assert.Equal(t, WhiteQueenside, game.CastlingRights, "Expected WhiteQueenside castling")
}
