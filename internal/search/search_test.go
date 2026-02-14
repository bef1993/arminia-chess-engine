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
	assert.Equal(t, expectedDepths, reportedDepths, "Info channel should receive updates for depths 1..maxDepth")
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
