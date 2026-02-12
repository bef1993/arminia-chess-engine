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
	// It's White's turn, so a positive score would mean White is better.
	assert.Equal(t, 0, score, "Initial position should have a score of 0")
}

func TestEvaluateMaterialAdvantage(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()

	// White has a Rook, Black has nothing.
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("a1", engine.WhiteRook)
	game.Board.SetPieceAt("e8", engine.BlackKing)

	// Test from White's perspective
	game.CurrentTurn = engine.White
	scoreWhite := Evaluate(game)
	assert.Equal(t, engine.RookValue, scoreWhite, "White should have a +%d advantage", engine.RookValue)

	// Test from Black's perspective
	// The material difference is the same, but since it's Black's turn,
	// the score should be negative (disadvantage).
	game.CurrentTurn = engine.Black
	scoreBlack := Evaluate(game)
	assert.Equal(t, -engine.RookValue, scoreBlack, "Black should have a -%d disadvantage", engine.RookValue)
}
