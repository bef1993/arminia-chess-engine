package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFEN(t *testing.T) {
	// Test starting position FEN
	startFen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	game, err := NewGameFromFEN(startFen)
	assert.NoError(t, err, "Failed to load start FEN")

	assert.Equal(t, White, game.CurrentTurn, "Expected White turn")
	assert.Equal(t, AllCastling, game.CastlingRights, "Expected AllCastling")

	// Test a custom position
	// White King on e1, Black King on e8, White Rook on a1
	customFen := "4k3/8/8/8/8/8/8/R3K3 w Q - 0 1"
	game, err = NewGameFromFEN(customFen)
	assert.NoError(t, err, "Failed to load custom FEN")

	assert.Equal(t, WhiteKing, game.Board.GetPiece(Sq("e1")), "Expected WhiteKing at e1")
	assert.Equal(t, BlackKing, game.Board.GetPiece(Sq("e8")), "Expected BlackKing at e8")
	assert.Equal(t, WhiteRook, game.Board.GetPiece(Sq("a1")), "Expected WhiteRook at a1")
	assert.Equal(t, WhiteQueenside, game.CastlingRights, "Expected WhiteQueenside castling")
}

func TestGenerateFEN(t *testing.T) {
	fen := "r3k2r/ppp2ppb/2p4p/4P3/3N2Pb/4B2P/PPP2P2/R2R2K1 b - - 0 16"
	game, err := NewGameFromFEN(fen)
	assert.NoError(t, err, "Failed to load FEN")

	generated := game.GenerateFEN()
	assert.Equal(t, fen, generated, "Generated FEN should match loaded FEN")

	move, _ := ParseMove("a7a5", game)
	game.ExecuteMove(move)
	game.UnmakeMove()
	assert.Equal(t, fen, generated, "Generated FEN should match loaded FEN")
}
