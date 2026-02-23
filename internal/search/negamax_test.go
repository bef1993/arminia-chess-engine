package search

import (
	"arminia-chess-engine/internal/engine"
	"sync/atomic"
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
	// Unique mate in 1 position
	// White Queen on d6, Black King on e8. Move Qd6-e6#
	fen := "3rkr2/8/3Q1p2/4p3/8/8/8/K7 w - - 0 1"
	game, err := engine.NewGameFromFEN(fen)
	assert.NoError(t, err)

	// Search should find the mate

	score, move, _ := negamax(NewSearchContext(game), 2, -EvalInfinity, EvalInfinity, 0, 0, true)

	// Expected move: Qd6-e6#
	assert.Equal(t, "d6e6", move.String(), "Should find mate d6e6")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInOneBlack(t *testing.T) {
	// Mate in 1 for Black
	// White King at a8, Black King at c7, Black Rook at b1. Move Rb1-a1#
	fen := "K7/2k5/8/8/8/8/8/1r6 b - - 0 1"
	game, err := engine.NewGameFromFEN(fen)
	assert.NoError(t, err)
	score, move, _ := negamax(NewSearchContext(game), 2, -EvalInfinity, EvalInfinity, 0, 0, true)

	// Expected move: Rb1-a1#
	assert.Equal(t, "b1a1", move.String(), "Should find mate b1a1")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInTwo(t *testing.T) {
	// Unique mate in 2 moves
	fen := "rn2kb2/ppp2p1Q/6pn/3p4/4q1b1/3P4/PPPK1PPP/RNB2BNR b q - 2 8"
	game, err := engine.NewGameFromFEN(fen)
	assert.NoError(t, err)
	score, move, _ := negamax(NewSearchContext(game), 4, -EvalInfinity, EvalInfinity, 0, 0, true)

	// Expected move: Qf4+, followed by Qxc1#
	assert.Equal(t, "e4f4", move.String(), "Should find mate in 2")
	assertMateScore(t, score)
}

func TestNegamax_FindsMateInThreeWithEnPassant(t *testing.T) {
	// Unique mate in 3 moves involving en passant
	fen := "rn3k1r/pp2p2p/3pQ1pn/1BpP2N1/5P2/3K4/P1PB2qP/8 w - - 2 17"
	game, err := engine.NewGameFromFEN(fen)
	assert.NoError(t, err)
	score, move, _ := negamax(NewSearchContext(game), 6, -EvalInfinity, EvalInfinity, 0, 0, true)

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
	sc := NewSearchContext(game)

	nodes1 := atomic.Int64{}
	sc.Nodes = &nodes1
	score1, move1, _ := negamax(sc, depth, -EvalInfinity, EvalInfinity, 0, 0, true)

	// 2. Second Search (Warm TT)
	// We expect the search to find the entry in the TT and return immediately or prune heavily
	nodes2 := atomic.Int64{}
	sc.Nodes = &nodes2
	score2, move2, _ := negamax(sc, depth, -EvalInfinity, EvalInfinity, 0, 0, true)

	// Assertions
	assert.Equal(t, move1, move2, "Best move should be consistent")
	assert.Equal(t, score1, score2, "Score should be consistent")

	// The second search should visit significantly fewer nodes
	// In a pure Negamax with TT, if the exact position is found at sufficient depth, nodes2 might be 1.
	assert.Less(t, nodes2.Load(), nodes1.Load(), "Second search should visit fewer nodes due to TT hit")
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

	// Call negamax with ply=1 to trigger draw detection
	score, _, _ := negamax(NewSearchContext(game), 2, -EvalInfinity, EvalInfinity, 1, 0, true)

	assert.Equal(t, 0, score, "Negamax should return a score of 0 for a draw by repetition, despite material advantage")
}

func TestNegamax_CheckExtension(t *testing.T) {
	game := engine.NewEmptyGame()
	// Setup: White King at e1, Black Rook at e8 (Check)
	// White has quiet moves (e.g., Kf1) but no captures.
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("e8", engine.BlackRook)
	game.CurrentTurn = engine.White

	assert.True(t, game.Board.IsKingInCheck(engine.White), "White King should be in check")

	selDepth := 0
	sc := NewSearchContext(game)
	sc.SelDepth = &selDepth

	// Search with depth 1.
	// If check extension works, it should search to depth 1+1 = 2.
	// Since there are no captures, Q-search won't extend further.
	// So selDepth should be 2.
	// If no extension, selDepth would be 1.
	negamax(sc, 1, -EvalInfinity, EvalInfinity, 0, 0, true)

	assert.GreaterOrEqual(t, selDepth, 2, "Search should extend depth when in check")
}
