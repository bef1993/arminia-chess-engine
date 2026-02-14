package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnightAttacks(t *testing.T) {
	// Test Knight at A1 (Corner)
	// Should attack B3 and C2
	// A Knight on the rim is dim :)
	attacks := KnightAttacks[A1]
	assert.Equal(t, 2, attacks.Count(), "Knight at A1 should have 2 attacks")
	assert.True(t, attacks.IsSet(B3), "Knight at A1 should attack B3")
	assert.True(t, attacks.IsSet(C2), "Knight at A1 should attack C2")

	// Test Knight at E4 (Center)
	// Should have 8 attacks
	attacks = KnightAttacks[E4]
	assert.Equal(t, 8, attacks.Count(), "Knight at E4 should have 8 attacks")
	
	// Verify specific targets for E4
	// E4 is file 4, rank 3 (0-indexed). 
	// Targets: (2,2)=C3, (2,4)=C5, (3,1)=D2, (3,5)=D6, (5,1)=F2, (5,5)=F6, (6,2)=G3, (6,4)=G5
	expectedTargets := []int{C3, C5, D2, D6, F2, F6, G3, G5}
	for _, target := range expectedTargets {
		assert.True(t, attacks.IsSet(target), "Knight at E4 should attack square index %d", target)
	}
}

func TestKingAttacks(t *testing.T) {
	// Test King at A1 (Corner)
	// Should attack A2, B1, B2
	attacks := KingAttacks[A1]
	assert.Equal(t, 3, attacks.Count(), "King at A1 should have 3 attacks")
	assert.True(t, attacks.IsSet(A2), "King at A1 should attack A2")
	assert.True(t, attacks.IsSet(B1), "King at A1 should attack B1")
	assert.True(t, attacks.IsSet(B2), "King at A1 should attack B2")

	// Test King at E1 (Edge)
	// Should attack D1, F1, D2, E2, F2
	attacks = KingAttacks[E1]
	assert.Equal(t, 5, attacks.Count(), "King at E1 should have 5 attacks")
	assert.True(t, attacks.IsSet(D1))
	assert.True(t, attacks.IsSet(F1))
	assert.True(t, attacks.IsSet(D2))
	assert.True(t, attacks.IsSet(E2))
	assert.True(t, attacks.IsSet(F2))

	// Test King at E4 (Center)
	// Should have 8 attacks
	attacks = KingAttacks[E4]
	assert.Equal(t, 8, attacks.Count(), "King at E4 should have 8 attacks")
	// Verify specific targets for E4
	expectedTargets := []int{D3, D4, D5, E3, E5, F3, F4, F5}
	for _, target := range expectedTargets {
		assert.True(t, attacks.IsSet(target), "King at E4 should attack square index %d", target)
	}
}
