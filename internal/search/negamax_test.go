package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertMateScore(t *testing.T, score int) {
	t.Helper()
	// Mate scores are close to EvalMate (29000).
	// The score is reduced by the number of plies to mate, so we check if it's
	// within a large margin of the base mate score to confirm it's a mate.
	assert.Greater(t, score, EvalMate-1000, "Score %d should indicate a forced mate", score)
}

func TestNegamax_FindsMateInOne(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 1 position
	// White Queen on d6, Black King on e8. Move Qd6-e6#
	fen := "3rkr2/8/3Q1p2/4p3/8/8/8/K7 w - - 0 1"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	// Search should find the mate
	// Depth 2 is required because depth 1 only evaluates the position (material),
	// while depth 2 checks if the opponent has any legal moves left.
	score, move, _ := negamax(context.Background(), game, 2, -EvalInfinity, EvalInfinity, 0, new(0), new(0))

	// Expected move: Qd6-e6#
	assert.Equal(t, "d6e6", move.String(), "Should find mate d6e6")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInOneBlack(t *testing.T) {
	game := engine.NewGame()
	// Mate in 1 for Black
	// White King at a8, Black King at c7, Black Rook at b1. Move Rb1-a1#
	fen := "K7/2k5/8/8/8/8/8/1r6 b - - 0 1"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	score, move, _ := negamax(context.Background(), game, 2, -EvalInfinity, EvalInfinity, 0, new(0), new(0))

	// Expected move: Rb1-a1#
	assert.Equal(t, "b1a1", move.String(), "Should find mate b1a1")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInTwo(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 2 moves
	fen := "rn2kb2/ppp2p1Q/6pn/3p4/4q1b1/3P4/PPPK1PPP/RNB2BNR b q - 2 8"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	score, move, _ := negamax(context.Background(), game, 4, -EvalInfinity, EvalInfinity, 0, new(0), new(0))

	// Expected move: Qf4+, followed by Qxc1#
	assert.Equal(t, "e4f4", move.String(), "Should find mate in 2")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInThreeWithEnPassant(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 3 moves involving en passant
	fen := "rn3k1r/pp2p2p/3pQ1pn/1BpP2N1/5P2/3K4/P1PB2qP/8 w - - 2 17"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	score, move, _ := negamax(context.Background(), game, 6, -EvalInfinity, EvalInfinity, 0, new(0), new(0))

	// Expected move: e6c8
	assert.Equal(t, "e6c8", move.String(), "Should find mate in 3 with en passant")
	assertMateScore(t, score)
}

func TestNegamax_TTIntegration_ReducesNodeCount(t *testing.T) {
	// Ensure TT is fresh and large enough
	GlobalTT.Resize(16)

	game := engine.NewGame()
	// Use a position that isn't the start pos to ensure some complexity
	// e.g. after 1. e4 e5
	m1, _ := engine.ParseMove("e2e4", game)
	game.ExecuteMove(m1)
	m2, _ := engine.ParseMove("e7e5", game)
	game.ExecuteMove(m2)

	// 1. First Search (Cold TT)
	depth := 4
	nodes1 := 0
	score1, move1, _ := negamax(context.Background(), game, depth, -EvalInfinity, EvalInfinity, 0, &nodes1, new(0))

	// 2. Second Search (Warm TT)
	// We expect the search to find the entry in the TT and return immediately or prune heavily
	nodes2 := 0
	score2, move2, _ := negamax(context.Background(), game, depth, -EvalInfinity, EvalInfinity, 0, &nodes2, new(0))

	// Assertions
	assert.Equal(t, move1, move2, "Best move should be consistent")
	assert.Equal(t, score1, score2, "Score should be consistent")

	// The second search should visit significantly fewer nodes
	// In a pure Negamax with TT, if the exact position is found at sufficient depth, nodes2 might be 1.
	assert.Less(t, nodes2, nodes1, "Second search should visit fewer nodes due to TT hit")
}

func TestNegamax_DetectsDrawByRepetition(t *testing.T) {
	game := engine.NewEmptyGame()
	// Setup a position where white is winning, but a repetition is possible
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("a1", engine.WhiteRook)
	game.Board.SetPieceAt("e8", engine.BlackKing)
	game.CurrentTurn = engine.White

	// The static evaluation should be > 0 for white
	initialEval := Evaluate(game)
	assert.Greater(t, initialEval, 400, "White should have a material advantage initially")

	// --- Create a 3-fold repetition ---
	// The board state is now the 1st occurrence.

	// Move sequence to repeat the position
	moves := []engine.Move{
		engine.NewMove(engine.Sq("a1"), engine.Sq("b1")), // W
		engine.NewMove(engine.Sq("e8"), engine.Sq("d8")), // B
		engine.NewMove(engine.Sq("b1"), engine.Sq("a1")), // W
		engine.NewMove(engine.Sq("d8"), engine.Sq("e8")), // B -> 2nd occurrence
		engine.NewMove(engine.Sq("a1"), engine.Sq("b1")), // W
		engine.NewMove(engine.Sq("e8"), engine.Sq("d8")), // B
		engine.NewMove(engine.Sq("b1"), engine.Sq("a1")), // W
		engine.NewMove(engine.Sq("d8"), engine.Sq("e8")), // B -> 3rd occurrence
	}

	for _, move := range moves {
		game.ExecuteMove(move)
	}

	assert.True(t, game.CanClaimDrawByThreefoldRepetition(), "Game state should recognize the 3-fold repetition")

	score, _, _ := negamax(context.Background(), game, 2, -EvalInfinity, EvalInfinity, 1, new(0), new(0))

	assert.Equal(t, 0, score, "Negamax should return a score of 0 for a draw by repetition, despite material advantage")
}
