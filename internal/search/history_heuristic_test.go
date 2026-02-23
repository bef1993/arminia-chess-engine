package search

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoryTable_AddAndGet(t *testing.T) {
	h := &HistoryTable{}
	move := engine.NewMove(engine.Sq("e2"), engine.Sq("e4"))
	depth := 5
	player := engine.White

	// Initial score should be 0
	assert.Equal(t, 0, h.Get(move, player))

	// Add score
	h.Add(move, depth, player)

	// Expected bonus is depth * depth = 25
	expected := depth * depth
	assert.Equal(t, expected, h.Get(move, player))

	// Add again
	h.Add(move, depth, player)
	assert.Equal(t, expected*2, h.Get(move, player))
}

func TestHistoryTable_ScaleDown(t *testing.T) {
	h := &HistoryTable{}
	move := engine.NewMove(engine.Sq("e2"), engine.Sq("e4"))
	player := engine.White

	// Set a high score
	// We need to manually set it or add enough times to reach near max
	// Since Add() calls ScaleDown() automatically if it exceeds max,
	// we can test ScaleDown() directly by setting a value and calling it.

	// Manually inject a large value to test explicit ScaleDown
	h[player][move.From][move.To] = 1000

	h.ScaleDown()

	assert.Equal(t, 500, h.Get(move, player), "Score should be halved after ScaleDown")
}

func TestHistoryTable_AutomaticScaling(t *testing.T) {
	h := &HistoryTable{}
	move := engine.NewMove(engine.Sq("a1"), engine.Sq("h8"))
	player := engine.Black

	// Force a value just below the limit
	h[player][move.From][move.To] = HistoryMax

	// Add a small depth to push it over
	h.Add(move, 1, player)

	// The value should have been added (HistoryMax + 1) then scaled down ((HistoryMax + 1) / 2)
	expected := (HistoryMax + 1) / 2
	assert.Equal(t, expected, h.Get(move, player), "Score should be scaled down automatically when exceeding max")
}
