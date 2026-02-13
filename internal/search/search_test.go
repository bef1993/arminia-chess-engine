package search

import (
	"arminia-chess-engine/internal/engine"
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

func TestSearchPlaceholder(t *testing.T) {
	game := engine.NewGame()

	// The placeholder search just returns the first legal move.
	// Let's just ensure it returns a valid move.
	move, _ := Search(game, 1)

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
	move, score := Search(game, 2)

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

	move, score := Search(game, 2)

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

	move, score := Search(game, 4)

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

	move, score := Search(game, 6)

	// Expected move: e6c8
	assert.Equal(t, "e6c8", move.String(), "Should find mate in 3 with en passant")
	assertMateScore(t, score)
}