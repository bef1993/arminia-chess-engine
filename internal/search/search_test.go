package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func assertMateScore(t *testing.T, score int) {
	t.Helper()
	// Mate scores are close to EvalMate (29000).
	// The score is reduced by the number of plies to mate, so we check if it's
	// within a large margin of the base mate score to confirm it's a mate.
	assert.Greater(t, score, EvalMate-1000, "Score %d should indicate a forced mate", score)
}

func TestSearchPlaceholder(t *testing.T) {
	game := engine.NewGame()

	// The placeholder search just returns the first legal move.
	// Let's just ensure it returns a valid move.
	move, _, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 1}, nil)

	// A zero-value move would have FromCol=0, FromRow=0, etc.
	assert.NotEqual(t, engine.Move{}, move, "Search should return a non-zero move from the starting position")
}

func TestSearchFindsMateInOne(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 1 position
	// White Queen on d6, Black King on e8. Move Qd6-e6#
	fen := "3rkr2/8/3Q1p2/4p3/8/8/8/K7 w - - 0 1"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	// Search should find the mate
	// Depth 2 is required because depth 1 only evaluates the position (material),
	// while depth 2 checks if the opponent has any legal moves left.
	move, score, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 2}, nil)

	// Expected move: Qd6-e6#
	assert.Equal(t, "d6e6", move.String(), "Should find mate d6e6")
	assertMateScore(t, score)
}

func TestSearchFindsMateInOneBlack(t *testing.T) {
	game := engine.NewGame()
	// Mate in 1 for Black
	// White King at a8, Black King at c7, Black Rook at b1. Move Rb1-a1#
	fen := "K7/2k5/8/8/8/8/8/1r6 b - - 0 1"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	move, score, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 2}, nil)

	// Expected move: Rb1-a1#
	assert.Equal(t, "b1a1", move.String(), "Should find mate b1a1")
	assertMateScore(t, score)
}

func TestSearchFindsMateInTwo(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 2 moves
	fen := "rn2kb2/ppp2p1Q/6pn/3p4/4q1b1/3P4/PPPK1PPP/RNB2BNR b q - 2 8"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	move, score, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 4}, nil)

	// Expected move: Qf4+, followed by Qxc1#
	assert.Equal(t, "e4f4", move.String(), "Should find mate in 2")
	assertMateScore(t, score)
}

func TestSearchFindsMateInThreeWithEnPassant(t *testing.T) {
	game := engine.NewGame()
	// Unique mate in 3 moves involving en passant
	fen := "rn3k1r/pp2p2p/3pQ1pn/1BpP2N1/5P2/3K4/P1PB2qP/8 w - - 2 17"
	err := game.LoadFEN(fen)
	assert.NoError(t, err)

	move, score, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 6}, nil)

	// Expected move: e6c8
	assert.Equal(t, "e6c8", move.String(), "Should find mate in 3 with en passant")
	assertMateScore(t, score)
}

func TestTTIntegration_ReducesNodeCount(t *testing.T) {
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
	move1, score1, nodes1 := Search(context.Background(), game, SearchOptions{MaxDepth: depth}, nil)

	// 2. Second Search (Warm TT)
	// We expect the search to find the entry in the TT and return immediately or prune heavily
	move2, score2, nodes2 := Search(context.Background(), game, SearchOptions{MaxDepth: depth}, nil)

	// Assertions
	assert.Equal(t, move1, move2, "Best move should be consistent")
	assert.Equal(t, score1, score2, "Score should be consistent")

	// The second search should visit significantly fewer nodes
	// In a pure Negamax with TT, if the exact position is found at sufficient depth, nodes2 might be 1.
	assert.Less(t, nodes2, nodes1, "Second search should visit fewer nodes due to TT hit")
}

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
	move, _, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 1}, nil)

	assert.NotEqual(t, "d1d8", move.String(), "Quiescence search should avoid bad capture d1d8")
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
	score, _ := quiescence(context.Background(), game, -EvalInfinity, EvalInfinity, 0, &nodes)

	// Score should reflect winning a pawn (~100)
	// We use 50 as a safe lower bound for a pawn advantage
	assert.Greater(t, score, 50, "Quiescence search should find en passant capture winning a pawn")
}

func TestIterativeDeepening_Callback(t *testing.T) {
	game := engine.NewGame()
	maxDepth := 3
	reportedDepths := []int{}

	infoCh := make(chan SearchInfo, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for info := range infoCh {
			reportedDepths = append(reportedDepths, info.Depth)
			assert.Greater(t, info.Nodes, 0, "Nodes should be > 0")
			assert.NotEqual(t, engine.Move{}, info.BestMove, "Best move should be valid")
		}
	}()

	move, _, nodes := Search(context.Background(), game, SearchOptions{MaxDepth: maxDepth}, infoCh)
	close(infoCh)
	<-done

	assert.NotEqual(t, engine.Move{}, move)
	assert.Greater(t, nodes, 0)

	expectedDepths := []int{1, 2, 3}
	assert.Equal(t, expectedDepths, reportedDepths, "Callback should be called for depths 1..maxDepth")
}

func TestIterativeDeepening_NodeAccumulation(t *testing.T) {
	game := engine.NewGame()
	maxDepth := 3

	var lastNodes int
	infoCh := make(chan SearchInfo, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for info := range infoCh {
			if info.Depth > 1 {
				assert.Greater(t, info.Nodes, lastNodes, "Total nodes should increase with depth")
			}
			lastNodes = info.Nodes
		}
	}()

	_, _, totalNodes := Search(context.Background(), game, SearchOptions{MaxDepth: maxDepth}, infoCh)
	close(infoCh)
	<-done

	assert.Equal(t, lastNodes, totalNodes, "Final returned nodes should match last callback nodes")
}

func TestSearch_RespectsTimeout(t *testing.T) {
	game := engine.NewGame()
	// Use a complex position to ensure it doesn't finish depth 20 instantly
	game.LoadFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

	// Set a short timeout
	duration := 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	start := time.Now()
	// Request a deep search that definitely takes longer than 50ms
	options := SearchOptions{MaxDepth: 20}

	move, _, _ := Search(ctx, game, options, nil)
	elapsed := time.Since(start)

	assert.NotEqual(t, engine.Move{}, move, "Should return a valid move even on timeout")
	// Allow some overhead for node checking interval
	assert.Less(t, elapsed, 500*time.Millisecond, "Search should stop near the timeout")
}

func TestSearch_RespectsCancellation(t *testing.T) {
	game := engine.NewGame()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately to test quick return
	cancel()

	start := time.Now()
	options := SearchOptions{MaxDepth: 20}
	move, _, _ := Search(ctx, game, options, nil)
	elapsed := time.Since(start)

	assert.NotEqual(t, engine.Move{}, move, "Should return a valid move")
	assert.Less(t, elapsed, 50*time.Millisecond, "Search should stop immediately after cancellation")
}
