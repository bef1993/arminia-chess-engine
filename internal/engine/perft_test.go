package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// perft is a helper function for Performance Test (Perft).
// It counts the number of leaf nodes in the game tree at a given depth.
// This is used to verify the correctness of the move generator.
func perft(game *Game, depth int) int {
	if depth == 0 {
		return 1
	}

	moves := game.GenerateLegalMoves()

	if depth == 1 {
		return len(moves)
	}

	nodes := 0
	for _, move := range moves {
		game.ExecuteMove(move)
		nodes += perft(game, depth-1)
		game.UnmakeMove()
	}
	return nodes
}

func TestPerft_StartPos_Depth2(t *testing.T) {
	game := NewGame()
	// Depth 1: 20 moves
	// Depth 2: 400 moves (20 * 20)
	nodes := perft(game, 2)
	assert.Equal(t, 400, nodes, "Expected 400 positions after 2 moves")
}

func TestPerft_StartPos_Depth3(t *testing.T) {
	game := NewGame()
	nodes := perft(game, 3)
	assert.Equal(t, 8902, nodes, "Expected 8,902 positions after 3 moves")
}

func TestPerft_StartPos_Depth4(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping depth 4 perft in short mode")
	}
	game := NewGame()
	nodes := perft(game, 4)
	assert.Equal(t, 197281, nodes, "Expected 197,281 positions after 4 moves")
}

func TestPerft_KiwiPete_Depth3(t *testing.T) {
	game := NewGame()
	err := game.LoadFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	assert.NoError(t, err)
	nodes := perft(game, 3)
	assert.Equal(t, 97862, nodes, "Expected 97,862 positions for KiwiPete at depth 3")
}

// 5268653 nodes/sec
func BenchmarkPerft_StartPos_Depth5(b *testing.B) {
	game := NewGame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		perft(game, 5)
	}
	// Depth 5 start pos has 4,865,609 nodes
	b.ReportMetric(float64(b.N)*4865609/b.Elapsed().Seconds(), "nodes/sec")
}

// 9543168 nodes/sec
func BenchmarkPerft_KiwiPete_Depth3(b *testing.B) {
	game := NewGame()
	game.LoadFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		perft(game, 3)
	}
	// Depth 3 KiwiPete has 97,862 nodes
	b.ReportMetric(float64(b.N)*97862/b.Elapsed().Seconds(), "nodes/sec")
}
