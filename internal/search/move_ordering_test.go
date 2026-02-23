package search

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderMoves(t *testing.T) {
	game, _ := engine.NewGameFromFEN("k2q4/4P3/8/1pP5/3Q1b2/3p4/8/K7 w - b6 0 1")
	noisyMoves := game.GetNoisyMoves()
	ttMove, _ := engine.ParseMove("d4d3", game)

	orderMoves(NewSearchContext(game), noisyMoves, ttMove, 0)

	// Expected Move Order based on current scoring:
	// 1. TT Move (d4d3)
	// 2. Capture-Promotions (e.g., e7xd8=Q) - Highest value due to combining capture and promotion scores.
	// 3. Good Captures (e.g., d4xd8, QxQ) - Based on MVV-LVA.
	// 4. Simple Promotions (e.g., e7-e8=Q)
	//
	// The test verifies this hierarchy.

	orderedMoves := ""
	for _, m := range noisyMoves {
		orderedMoves += m.String() + " "
	}

	// This is the correct order based on the current scoring constants.
	// Captures (Base 1,000,000) are scored higher than simple promotions (Base 900,000).
	// Capture-promotions are scored highest as they combine both scores.
	expectedOrder := "d4d3 e7d8q e7d8r e7d8b e7d8n d4d8 d4f4 c5b6 e7e8q e7e8r e7e8b e7e8n "
	assert.Equal(t, expectedOrder, orderedMoves, "Moves should be ordered correctly")
}
