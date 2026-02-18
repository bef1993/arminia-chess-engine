package search

import (
	"arminia-chess-engine/internal/engine"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTranspositionTable(t *testing.T) {
	// 1 MB
	tt := NewTranspositionTable(1)
	assert.NotNil(t, tt)
	assert.Greater(t, tt.size, uint64(0))
	assert.Equal(t, int(tt.size), len(tt.entries))
}

func TestStoreAndProbe(t *testing.T) {
	tt := NewTranspositionTable(1)
	hash := uint64(0x123456789ABC)
	move := engine.NewMove(engine.Sq("b2"), engine.Sq("b4"))
	depth := 5
	score := 150
	flag := FlagExact

	tt.Store(hash, depth, score, flag, move)

	entry, found := tt.Probe(hash)
	assert.True(t, found)
	assert.Equal(t, hash, entry.Hash)
	assert.Equal(t, depth, entry.Depth)
	assert.Equal(t, score, entry.Score)
	assert.Equal(t, flag, entry.Flag)
	assert.Equal(t, move, entry.BestMove)
}

func TestProbeMiss(t *testing.T) {
	tt := NewTranspositionTable(1)
	hash := uint64(0xDEADBEEF)

	_, found := tt.Probe(hash)
	assert.False(t, found)
}

func TestOverwrite(t *testing.T) {
	// Create a TT. NewTranspositionTable has a minimum count, so we calculate a collision.
	tt := NewTranspositionTable(1)

	hash1 := uint64(10)
	// Ensure hash2 maps to the same index as hash1
	// index = hash % size. If hash2 = hash1 + size, then (hash1 + size) % size == hash1 % size.
	hash2 := hash1 + tt.size

	move1 := engine.NewMove(0, 0)
	move2 := engine.NewMove(1, 1)

	// Store first
	tt.Store(hash1, 1, 100, FlagExact, move1)

	// Verify it's there
	entry, found := tt.Probe(hash1)
	assert.True(t, found)
	assert.Equal(t, move1, entry.BestMove)

	// Store second (collision)
	tt.Store(hash2, 2, 200, FlagLowerBound, move2)

	// Verify hash1 is gone (replaced because index collision)
	_, found = tt.Probe(hash1)
	assert.False(t, found, "Old entry should be overwritten due to index collision")

	// Verify hash2 is there
	entry, found = tt.Probe(hash2)
	assert.True(t, found)
	assert.Equal(t, move2, entry.BestMove)
}

func TestResize(t *testing.T) {
	tt := NewTranspositionTable(1)
	hash := uint64(0xCAFEBABE)
	tt.Store(hash, 1, 100, FlagExact, engine.Move{})

	// Verify stored
	_, found := tt.Probe(hash)
	assert.True(t, found)

	// Resize to 2MB
	tt.Resize(2)

	// Verify cleared
	_, found = tt.Probe(hash)
	assert.False(t, found, "Table should be cleared after resize")

	// Verify size changed (approx double)
	// 1MB -> ~13107 entries, 2MB -> ~26214 entries
	assert.Greater(t, tt.size, uint64(20000))
}

func TestStorePreservesDeeper(t *testing.T) {
	tt := NewTranspositionTable(1)
	hash := uint64(0x12345)
	deepMove := engine.NewMove(1, 1)
	shallowMove := engine.NewMove(2, 2)

	// 1. Store a deep search result (e.g. depth 5)
	tt.Store(hash, 5, 100, FlagExact, deepMove)

	// 2. Attempt to store a shallow result (e.g. depth 0 from Quiescence)
	tt.Store(hash, 0, 105, FlagExact, shallowMove)

	// 3. Verify the deep result was preserved
	entry, found := tt.Probe(hash)
	assert.True(t, found)
	assert.Equal(t, 5, entry.Depth, "Should preserve deeper depth")
	assert.Equal(t, 100, entry.Score, "Should preserve score from deeper search")
	assert.Equal(t, deepMove, entry.BestMove, "Should preserve best move from deeper search")

	// 4. Verify we CAN overwrite if depth is greater
	tt.Store(hash, 6, 200, FlagExact, shallowMove)
	entry, _ = tt.Probe(hash)
	assert.Equal(t, 6, entry.Depth, "Should overwrite if new depth is greater")
	assert.Equal(t, shallowMove, entry.BestMove)
}

func TestPackUnpack(t *testing.T) {
	tests := []struct {
		name  string
		depth int
		score int
		flag  TTFlag
		move  engine.Move
	}{
		{
			name:  "Standard Entry",
			depth: 5,
			score: 150,
			flag:  FlagExact,
			move:  engine.NewMove(engine.Sq("e2"), engine.Sq("e4")),
		},
		{
			name:  "Negative Score",
			depth: 10,
			score: -300,
			flag:  FlagLowerBound,
			move:  engine.NewMove(engine.Sq("a1"), engine.Sq("h8")),
		},
		{
			name:  "Promotion Move",
			depth: 3,
			score: 900,
			flag:  FlagUpperBound,
			move:  engine.NewPromotionMove(engine.Sq("a7"), engine.Sq("a8"), engine.WhiteQueen),
		},
		{
			name:  "Empty Move",
			depth: 1,
			score: 0,
			flag:  FlagExact,
			move:  engine.Move{},
		},
		{
			name:  "Null Move from NewMove",
			depth: 2,
			score: 10,
			flag:  FlagExact,
			move:  engine.NewMove(0, 0), // This is {From: 0, To: 0, PromotionPiece: -1}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := pack(tt.depth, tt.score, tt.flag, tt.move)
			// Use a dummy hash since it's not packed
			hash := uint64(0x1234567890ABCDEF)
			entry := unpack(hash, packed)

			assert.Equal(t, hash, entry.Hash)
			assert.Equal(t, tt.depth, entry.Depth)
			assert.Equal(t, tt.score, entry.Score)
			assert.Equal(t, tt.flag, entry.Flag)
			assert.Equal(t, tt.move, entry.BestMove)
		})
	}
}
