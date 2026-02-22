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

	// Make sure that TT move is #1
	// Pawn capture Queen and promote to Queen is #2
	// Pawn capture Queen and promote to Rook is #3
	// Pawn capture Queen and promote to Bishop is #4
	// Pawn capture Queen and promote to Knight is #5
	// Pawn promoting to a Queen is #6
	// Queen capturing enemy Queen is #7
	// Pawn promoting to a Rook is #8
	// Pawn promoting to a Bishop is #9
	// Pawn promoting to a Knight is #10
	// Queen capturing Bishop is #11
	// Pawn capturing en passant is #12

	orderedMoves := ""
	for _, m := range noisyMoves {
		orderedMoves += m.String() + " "
	}
	assert.Equal(t, "d4d3 e7d8q e7d8r e7d8b e7d8n e7e8q d4d8 e7e8r e7e8b e7e8n d4f4 c5b6 ", orderedMoves, "Moves should be ordered correctly")
}
