package search

import (
	"arminia-chess-engine/internal/engine"
)

// TTFlag represents the type of score stored in the transposition table
type TTFlag uint8

const (
	FlagExact      TTFlag = iota // Exact score
	FlagLowerBound               // Alpha (Fail Low) - True score is >= stored score
	FlagUpperBound               // Beta (Fail High) - True score is <= stored score
)

// DefaultTTSizeMB is the default size of the transposition table in megabytes.
const DefaultTTSizeMB = 64

// TTEntry represents a single entry in the transposition table
type TTEntry struct {
	Hash     uint64
	Score    int
	Depth    int
	Flag     TTFlag
	BestMove engine.Move
}

// TranspositionTable represents the hash table
type TranspositionTable struct {
	entries []TTEntry
	size    uint64
}

// GlobalTT is the global transposition table instance
var GlobalTT *TranspositionTable

func init() {
	// Initialize with a default size
	GlobalTT = NewTranspositionTable(DefaultTTSizeMB)
}

// NewTranspositionTable creates a new TT with the given size in MB
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	// Estimate entry size.
	// Hash(8) + Score(8) + Depth(8) + Flag(1) + Move(~40) + Padding ~ 72 bytes
	// We'll be conservative and assume ~80 bytes per entry
	entrySize := 80
	count := (sizeMB * 1024 * 1024) / entrySize

	if count <= 0 {
		count = 1024 // Minimum fallback
	}

	return &TranspositionTable{
		entries: make([]TTEntry, count),
		size:    uint64(count),
	}
}

// Probe retrieves an entry from the TT
func (tt *TranspositionTable) Probe(hash uint64) (TTEntry, bool) {
	if tt.size == 0 {
		return TTEntry{}, false
	}
	index := hash % tt.size
	entry := tt.entries[index]

	if entry.Hash == hash {
		return entry, true
	}
	return TTEntry{}, false
}

// Store saves an entry to the TT
func (tt *TranspositionTable) Store(hash uint64, depth, score int, flag TTFlag, bestMove engine.Move) {
	if tt.size == 0 {
		return
	}
	index := hash % tt.size

	entry := &tt.entries[index]

	// If we have an entry for the same position, be careful about overwriting.
	if entry.Hash == hash {
		// 1. Don't overwrite deep results with shallow ones (e.g. QS overwriting Main Search)
		if depth < entry.Depth {
			return
		}
		// 2. Preserve existing best move if the new one is empty (e.g. fail-low)
		if bestMove == (engine.Move{}) {
			bestMove = entry.BestMove
		}
	}

	// Replacement strategy: Always replace (if not same hash & deeper)
	*entry = TTEntry{
		Hash:     hash,
		Score:    score,
		Depth:    depth,
		Flag:     flag,
		BestMove: bestMove,
	}
}

// Resize resizes the TT (clears existing entries)
func (tt *TranspositionTable) Resize(sizeMB int) {
	*tt = *NewTranspositionTable(sizeMB)
}

// Clear clears the transposition table
func (tt *TranspositionTable) Clear() {
	tt.entries = make([]TTEntry, len(tt.entries))
}
