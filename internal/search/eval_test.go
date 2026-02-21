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

func TestEvaluateThreats(t *testing.T) {
	// 1. Hanging Piece (White Queen)
	// White Queen at e4. Black Rook at e8.
	game := engine.NewEmptyGame()
	game.Board.SetPieceAt("e4", engine.WhiteQueen)
	game.Board.SetPieceAt("e6", engine.BlackRook)
	game.Board.SetPieceAt("d7", engine.BlackPawn)

	// Expected penalty: HangingPiecePenalty (100) + QueenValue (900) / 4 = 325
	// Since it's white piece hanging, score should be -325
	score := evaluateThreats(&game.Board)
	assert.Equal(t, -325, score, "Hanging White Queen should be penalized")

	// 2. Hanging Piece (Black Rook)
	// Black Rook at a8. White Bishop at h1.
	game.Board.Clear()
	game.Board.SetPieceAt("a8", engine.BlackRook)
	game.Board.SetPieceAt("h1", engine.WhiteBishop)

	// Expected penalty: HangingPiecePenalty (100) + RookValue (500) / 4 = 225
	// Since it's black piece hanging, score should be +225
	score = evaluateThreats(&game.Board)
	assert.Equal(t, 225, score, "Hanging Black Rook should be penalized (positive score)")

	// 3. Defended Piece (White Knight attacked by Black Pawn)
	// White Knight at e4. Black Pawn at d5. White Pawn at f3.
	game.Board.Clear()
	game.Board.SetPieceAt("e4", engine.WhiteKnight)
	game.Board.SetPieceAt("d5", engine.BlackPawn)
	game.Board.SetPieceAt("f3", engine.WhitePawn)

	// Attacker: Pawn (100). Victim: Knight (320).
	// Penalty: (320 - 100) / 4 = 55
	// White piece threatened -> -55
	score = evaluateThreats(&game.Board)
	assert.Equal(t, -55, score, "White Knight attacked by Pawn should be penalized")

	// 4. Defended Piece (Equal value or attacker more valuable)
	// White Pawn at e4. Black Knight at f6. White Pawn at d3.
	// White Pawn attacked by Knight. Defended by Pawn.
	// Attacker (Knight 320) > Victim (Pawn 100).
	// Penalty should be 0 because attacker is more valuable (not a tactical weakness in this simple model).
	game.Board.Clear()
	game.Board.SetPieceAt("e4", engine.WhitePawn)
	game.Board.SetPieceAt("f6", engine.BlackKnight)
	game.Board.SetPieceAt("d3", engine.WhitePawn)

	score = evaluateThreats(&game.Board)
	assert.Equal(t, 0, score, "Pawn attacked by Knight (defended) should not be penalized")

	// 5. Defended Piece (Attacker less valuable but still a threat)
	// White Rook at e4. Black Knight at f6. White Pawn at d3.
	// White Rook attacked by Knight. Defended by Pawn.
	// Attacker (Knight 320) < Victim (Rook 500).
	// Penalty: (500 - 320) / 4 = 45
	// White piece threatened -> -45
	game.Board.Clear()
	game.Board.SetPieceAt("e4", engine.WhiteRook)
	game.Board.SetPieceAt("f6", engine.BlackKnight)
	game.Board.SetPieceAt("d3", engine.WhitePawn)

	score = evaluateThreats(&game.Board)
	assert.Equal(t, -45, score, "White Rook attacked by Knight (defended) should be penalized")
}
