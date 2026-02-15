package search

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateInitialPosition(t *testing.T) {
	game := engine.NewGame()
	score := Evaluate(game)

	// In the initial position, the material is equal, so the score should be 0.
	assert.Equal(t, 0, score, "Initial position should have a score of 0")
}

func TestEvaluateMaterialAdvantage(t *testing.T) {
	game := engine.NewEmptyGame()

	// White has a Rook, Black has nothing.
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("a1", engine.WhiteRook)
	game.Board.SetPieceAt("e8", engine.BlackKing)

	// Test from White's perspective
	game.CurrentTurn = engine.White
	scoreWhite := Evaluate(game)
	// We use Greater instead of Equal to allow for small PST adjustments
	assert.Greater(t, scoreWhite, 400, "White should have a rook advantage")

	// Test from Black's perspective
	game.CurrentTurn = engine.Black
	scoreBlack := Evaluate(game)
	assert.Less(t, scoreBlack, -400, "Black should have a rook disadvantage")
}

func TestEvaluatePositionalAdvantage(t *testing.T) {
	game := engine.NewEmptyGame()
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("e8", engine.BlackKing)

	// Equal material: 1 Knight each
	// White Knight on e5 (Strong central outpost)
	game.Board.SetPieceAt("e5", engine.WhiteKnight)

	// Black Knight on h8 (Corner, terrible square)
	game.Board.SetPieceAt("h8", engine.BlackKnight)

	game.CurrentTurn = engine.White
	score := Evaluate(game)

	// White should be winning despite equal material due to PST
	assert.Greater(t, score, 10, "White should have advantage due to better piece placement")
}

func TestEvaluateSymmetry(t *testing.T) {
	game := engine.NewEmptyGame()

	// Perfectly symmetric position
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("e8", engine.BlackKing)

	game.Board.SetPieceAt("d4", engine.WhitePawn)
	game.Board.SetPieceAt("d5", engine.BlackPawn)

	game.Board.SetPieceAt("c3", engine.WhiteKnight)
	game.Board.SetPieceAt("c6", engine.BlackKnight)

	game.CurrentTurn = engine.White
	score := Evaluate(game)
	assert.Equal(t, 0, score, "Symmetric position should have 0 score")
}

func TestEvaluateCenterControl(t *testing.T) {
	game := engine.NewEmptyGame()
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("e8", engine.BlackKing)

	// White pawn in center (e4)
	game.Board.SetPieceAt("e4", engine.WhitePawn)

	// Black pawn on side (h5)
	game.Board.SetPieceAt("h5", engine.BlackPawn)

	game.CurrentTurn = engine.White
	score := Evaluate(game)

	assert.Greater(t, score, 0, "Center pawn should be evaluated higher than side pawn")
}
