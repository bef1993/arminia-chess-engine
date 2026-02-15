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
