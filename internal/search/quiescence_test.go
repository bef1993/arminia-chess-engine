package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuiescence_AvoidsBadCapture(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()

	// Setup:
	// White: King e1, Queen d1, Pawn h2 (to provide a quiet move)
	// Black: King e8, Rook d8 (protected by Knight), Knight c6
	// White to move.
	// Bad capture: Qxd8+ (Exchange Q(9) for R(5). Net -4).
	// Quiet move: h2h3 (Score ~ +1 due to material advantage Q vs R+N).

	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("d1", engine.WhiteQueen)
	game.Board.SetPieceAt("h2", engine.WhitePawn)

	game.Board.SetPieceAt("e8", engine.BlackKing)
	game.Board.SetPieceAt("d8", engine.BlackRook)
	game.Board.SetPieceAt("c6", engine.BlackKnight)

	game.CurrentTurn = engine.White

	// Search at depth 1.
	// Without QS, negamax sees Qxd8 -> +6 (Q vs N) because it doesn't see the recapture.
	// With QS, negamax sees Qxd8 -> -3 (N vs nothing) because it sees Nxd8.
	move, eval, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 1}, nil)

	assert.NotEqual(t, "d1d8", move.String(), "Quiescence search should avoid bad capture d1d8")
	assert.Greater(t, eval, 0, "eval should be greater than 0")
}

func TestQuiescence_IncludesEnPassant(t *testing.T) {
	game := engine.NewGame()
	game.Board.Clear()

	// Setup: White Pawn e5, Black Pawn d5 (just moved). EP target d6.
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("e5", engine.WhitePawn)
	game.Board.SetPieceAt("e8", engine.BlackKing)
	game.Board.SetPieceAt("d5", engine.BlackPawn)

	game.CurrentTurn = engine.White
	game.EnPassantTargetCol = engine.FileD
	game.EnPassantTargetRow = engine.Rank6

	// Evaluate at root should be 0 (equal material).
	// Quiescence should find exd6 e.p. which wins a pawn.
	nodes := 0
	var selDepth int
	score, _ := quiescence(context.Background(), game, -EvalInfinity, EvalInfinity, 0, &nodes, &selDepth)

	// Score should reflect winning a pawn (~100)
	// We use 50 as a safe lower bound for a pawn advantage
	assert.Greater(t, score, 50, "Quiescence search should find en passant capture winning a pawn")
}