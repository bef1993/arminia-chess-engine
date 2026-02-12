package search

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchPlaceholder(t *testing.T) {
	game := engine.NewGame()

	// The placeholder search just returns the first legal move.
	// Let's just ensure it returns a valid move.
	move := Search(game, 1)

	// A zero-value move would have FromCol=0, FromRow=0, etc.
	assert.NotEqual(t, engine.Move{}, move, "Search should return a non-zero move from the starting position")
}