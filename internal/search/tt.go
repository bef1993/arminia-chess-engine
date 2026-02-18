package search

import (
	"arminia-chess-engine/internal/engine"
	"sync/atomic"
)

// TTFlag represents the type of score stored in the transposition table
type TTFlag uint8

const (
	FlagExact      TTFlag = iota // Exact score
	FlagLowerBound               // Alpha (Fail Low) - True score is >= stored score
	FlagUpperBound               // Beta (Fail High) - True score is <= stored score
)

// DefaultTTSizeMB is the default size of the transposition table in megabytes.
const DefaultTTSizeMB = 512

// TTEntry represents the unpacked data
type TTEntry struct {
	Hash     uint64
	Score    int
	Depth    int
	Flag     TTFlag
	BestMove engine.Move
}

// packedEntry represents the compact storage
type packedEntry struct {
	key  uint64 // hash ^ data
	data uint64 // packed data
}

// TranspositionTable represents the hash table
type TranspositionTable struct {
	entries []packedEntry
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
	// Entry size is 16 bytes (2 uint64s)
	entrySize := 16
	count := (sizeMB * 1024 * 1024) / entrySize

	if count <= 0 {
		count = 1024 // Minimum fallback
	}

	return &TranspositionTable{
		entries: make([]packedEntry, count),
		size:    uint64(count),
	}
}

// Probe retrieves an entry from the TT
func (tt *TranspositionTable) Probe(hash uint64) (TTEntry, bool) {
	if tt.size == 0 {
		return TTEntry{}, false
	}
	index := hash % tt.size
	entry := &tt.entries[index]

	// Lockless read
	data := atomic.LoadUint64(&entry.data)
	key := atomic.LoadUint64(&entry.key)

	if (key ^ data) == hash {
		return unpack(hash, data), true
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

	// Lockless replacement strategy
	oldData := atomic.LoadUint64(&entry.data)
	oldKey := atomic.LoadUint64(&entry.key)

	replace := false
	if (oldKey ^ oldData) != hash {
		// Collision or empty: Always replace
		replace = true
	} else {
		// Same position: Replace if new depth is sufficient or better quality
		oldDepth := int(uint8((oldData >> 16) & 0xFF))
		oldFlag := TTFlag((oldData >> 24) & 0x3)

		if depth > oldDepth {
			replace = true
		} else if depth == oldDepth {
			// If depths are equal, replace unless we are downgrading from Exact to Bound.
			// 1. If new is Exact, always replace (Upgrade or Refresh).
			// 2. If old is NOT Exact, always replace (Upgrade or Refresh).
			if flag == FlagExact || oldFlag != FlagExact {
				replace = true
			}
		}
	}

	if replace {
		newData := pack(depth, score, flag, bestMove)

		// Optimization: Preserve the existing move if the new search didn't find one (e.g. fail-low).
		// This ensures we don't lose the Hash Move for the next search.
		if bestMove == (engine.Move{}) && (oldKey^oldData) == hash {
			newData |= oldData & (0xFFFFF << 26) // Copy move bits (20 bits starting at 26)
		}

		newKey := hash ^ newData

		atomic.StoreUint64(&entry.data, newData)
		atomic.StoreUint64(&entry.key, newKey)
	}
}

// Resize resizes the TT (clears existing entries)
func (tt *TranspositionTable) Resize(sizeMB int) {
	*tt = *NewTranspositionTable(sizeMB)
}

// Clear clears the transposition table
func (tt *TranspositionTable) Clear() {
	tt.entries = make([]packedEntry, len(tt.entries))
}

func pack(depth, score int, flag TTFlag, move engine.Move) uint64 {
	// Clamp score to int16 range
	if score > 32000 {
		score = 32000
	} else if score < -32000 {
		score = -32000
	}

	// Clamp depth to uint8
	if depth > 255 {
		depth = 255
	}

	d := uint64(uint16(score))      // Bits 0-15
	d |= uint64(uint8(depth)) << 16 // Bits 16-23
	d |= uint64(flag) << 24         // Bits 24-25

	// Move packing: From(6) + To(6) + Promo(8)
	mv := uint64(move.From) & 0x3F
	mv |= (uint64(move.To) & 0x3F) << 6
	mv |= (uint64(move.PromotionPiece) & 0xFF) << 12

	d |= mv << 26 // Bits 26-45

	return d
}

func unpack(hash, data uint64) TTEntry {
	score := int(int16(data & 0xFFFF))
	depth := int(uint8((data >> 16) & 0xFF))
	flag := TTFlag((data >> 24) & 0x3)

	mvData := data >> 26
	from := int(mvData & 0x3F)
	to := int((mvData >> 6) & 0x3F)
	promoRaw := (mvData >> 12) & 0xFF
	var promo engine.Piece
	if promoRaw == 0xFF { // TODO this is necessary because NoPiece currently is -1, which becomes 0xFF when cast to uint64. We should consider changing NoPiece to 0 for cleaner packing.
		promo = engine.NoPiece
	} else {
		promo = engine.Piece(promoRaw)
	}

	move := engine.Move{From: from, To: to, PromotionPiece: promo}

	return TTEntry{
		Hash:     hash,
		Score:    score,
		Depth:    depth,
		Flag:     flag,
		BestMove: move,
	}
}
