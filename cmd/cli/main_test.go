package main

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMove_NormalMove(t *testing.T) {
	game := engine.NewGame()
	// e2e4
	err := handleMove(game, "e2e4")
	assert.NoError(t, err)

	// Check piece moved
	piece := game.Board.GetPieceAt("e4")
	assert.Equal(t, engine.WhitePawn, piece)
	assert.Equal(t, engine.NoPiece, game.Board.GetPieceAt("e2"))
}

func TestHandleMove_InvalidFormat(t *testing.T) {
	game := engine.NewGame()
	err := handleMove(game, "e2")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid move format")
	}
}

func TestHandleMove_OutOfBounds(t *testing.T) {
	game := engine.NewGame()
	err := handleMove(game, "i2e4") // 'i' is col 8 (out of bounds 0-7)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "square out of bounds")
	}
}

func TestHandleMove_IllegalMove(t *testing.T) {
	game := engine.NewGame()
	// Pawn e2 to e5 is illegal (can only move to e3 or e4)
	err := handleMove(game, "e2e5")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "illegal move")
	}
}

func TestHandleMove_Promotion(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()
	// White pawn at e7
	game.Board.SetPieceAt("e7", engine.WhitePawn)

	// Promote to Queen
	err := handleMove(game, "e7e8q")
	assert.NoError(t, err)

	piece := game.Board.GetPieceAt("e8")
	assert.Equal(t, engine.WhiteQueen, piece)
}

func TestHandleMove_Promotion_MissingPiece(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()
	game.Board.SetPieceAt("e7", engine.WhitePawn)

	// Missing promotion piece
	err := handleMove(game, "e7e8")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "promotion piece required")
	}
}

func TestHandleMove_Promotion_InvalidPiece(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()
	game.Board.SetPieceAt("e7", engine.WhitePawn)

	// Invalid promotion piece 'k' (King)
	err := handleMove(game, "e7e8k")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid promotion piece")
	}
}

func TestHandleMove_Castling(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("h1", engine.WhiteRook)
	game.CastlingRights = engine.WhiteKingside

	// Execute castling move (e1 to g1)
	err := handleMove(game, "e1g1")
	assert.NoError(t, err)

	// Check King position
	king := game.Board.GetPieceAt("g1")
	assert.Equal(t, engine.WhiteKing, king)

	// Check Rook position (should be at f1)
	rook := game.Board.GetPieceAt("f1")
	assert.Equal(t, engine.WhiteRook, rook)
}
