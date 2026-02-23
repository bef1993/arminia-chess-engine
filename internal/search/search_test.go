package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchPlaceholder(t *testing.T) {
	game := engine.NewGame()

	// The placeholder search just returns the first legal move.
	// Let's just ensure it returns a valid move.
	move, _, _ := Search(context.Background(), game, SearchOptions{MaxDepth: 1}, nil)

	// A zero-value move would have FromCol=0, FromRow=0, etc.
	assert.NotEqual(t, engine.Move{}, move, "Search should return a non-zero move from the starting position")
}

func TestIterativeDeepening_ReportsInfo(t *testing.T) {
	game := engine.NewGame()
	maxDepth := 3
	reportedDepths := []int{}

	infoCh := make(chan SearchInfo, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for info := range infoCh {
			reportedDepths = append(reportedDepths, info.Depth)
			assert.Greater(t, info.Nodes, int64(0), "Nodes should be > 0")
			assert.NotEqual(t, engine.Move{}, info.BestMove, "Best move should be valid")
		}
	}()

	move, _, nodes := Search(context.Background(), game, SearchOptions{MaxDepth: maxDepth}, infoCh)
	close(infoCh)
	<-done

	assert.NotEqual(t, engine.Move{}, move, "Search should return a valid move")
	assert.Greater(t, nodes, int64(0), "Search should visit some nodes")

	expectedDepths := []int{1, 2, 3}
	assert.Equal(t, expectedDepths, reportedDepths, "Info channel should receive updates for depths 1..maxDepth")
}

func TestIterativeDeepening_NodeAccumulation(t *testing.T) {
	game := engine.NewGame()
	maxDepth := 3

	var lastNodes int64
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

	_, _, totalNodes := Search(context.Background(), game, SearchOptions{MaxDepth: maxDepth, Threads: 1}, infoCh)
	close(infoCh)
	<-done

	assert.Equal(t, lastNodes, totalNodes, "Final returned nodes should match last callback nodes")
}

func TestSearch_RespectsTimeout(t *testing.T) {
	// Use a complex position to ensure it doesn't finish depth 20 instantly
	game, _ := engine.NewGameFromFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

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

func TestSearch_AvoidsRepetitionWhenWinning(t *testing.T) {
	// Setup a winning position for White
	// White King e1, Rook a1. Black King e8.
	// White is winning (Rook up).
	game := engine.NewGame()
	game.Board.Clear()
	game.Board.SetPieceAt("e1", engine.WhiteKing)
	game.Board.SetPieceAt("a1", engine.WhiteRook)
	game.Board.SetPieceAt("e8", engine.BlackKing)
	game.CurrentTurn = engine.White
	game.CastlingRights = engine.NoCastling

	// Create a history of repetition
	// We want the move Ra1-b1 to lead to "Pos B" for the 3rd time.
	moves := []string{
		"a1b1", "e8d8", // 1
		"b1a1", "d8e8", // 2
		"a1b1", "e8d8", // 3
		"b1a1", "d8e8", // 4
	}

	for _, m := range moves {
		move, _ := engine.ParseMove(m, game)
		game.ExecuteMove(move)
	}

	// Verify Ra1-b1 is a repetition
	clone := game.Clone()
	move, _ := engine.ParseMove("a1b1", clone)
	clone.ExecuteMove(move)
	assert.True(t, clone.CanClaimDrawByThreefoldRepetition(), "a1b1 should lead to 3rd repetition")
}
