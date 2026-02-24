package polyglot

import (
	"arminia-chess-engine/internal/engine"
	"encoding/binary"
	"math/rand"
	"os"
	"time"
)

// --- Opening OpeningBook State ---
var (
	BookEnabled  bool = false
	BookMaxDepth int  = 40 // Max ply (20 full moves) to use the book, from config.yml.
	OpeningBook  *Book
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// PolyglotEntry represents a single entry in the .bin file.
// 16 bytes total.
const PolyglotEntrySizeBytes = 16

type PolyglotEntry struct {
	Key    uint64 // 8 bytes
	Move   uint16 // 2 bytes
	Weight uint16 // 2 bytes
	Learn  uint32 // 4 bytes
}

// Book represents an opening book.
type Book struct {
	file *os.File
}

// OpenBook initializes the global opening book from the given path.
// It closes any previously opened book.
func OpenBook(path string) error {
	if OpeningBook != nil {
		OpeningBook.Close()
		OpeningBook = nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	OpeningBook = &Book{file: f}
	return nil
}

// Close closes the book file.
func (b *Book) Close() error {
	return b.file.Close()
}

// Probe returns a move from the book for the current position.
// It returns a zero move if no move is found or an error occurs.
// It uses a weighted random selection if multiple moves are available.
func (b *Book) Probe(game *engine.Game) engine.Move {
	hash := ComputePolyglotHash(game)

	// The entries are sorted by key. We can use binary search.
	// However, since the file is on disk, we can't easily slice it.
	// For simplicity and performance with large books, we usually map the file or seek.
	// Here we implement a simple binary search using Seek.

	stat, err := b.file.Stat()
	if err != nil {
		return engine.Move{}
	}

	fileSize := stat.Size()
	numEntries := fileSize / PolyglotEntrySizeBytes

	low := int64(0)
	high := numEntries - 1
	var firstMatch int64 = -1

	// Binary search to find the *first* entry with the matching key
	for low <= high {
		mid := low + (high-low)/2
		b.file.Seek(mid*PolyglotEntrySizeBytes, 0)

		var entry PolyglotEntry
		if err := binary.Read(b.file, binary.BigEndian, &entry); err != nil {
			return engine.Move{}
		}

		if entry.Key < hash {
			low = mid + 1
		} else if entry.Key > hash {
			high = mid - 1
		} else {
			// Found a match, but it might not be the first one.
			firstMatch = mid
			high = mid - 1 // Continue searching in the lower half
		}
	}

	if firstMatch == -1 {
		return engine.Move{}
	}

	// Read all entries with this key
	var entries []PolyglotEntry
	b.file.Seek(firstMatch*PolyglotEntrySizeBytes, 0)

	for {
		var entry PolyglotEntry
		if err := binary.Read(b.file, binary.BigEndian, &entry); err != nil {
			break
		}
		if entry.Key != hash {
			break
		}
		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return engine.Move{}
	}

	// Weighted random selection
	totalWeight := 0
	for _, e := range entries {
		totalWeight += int(e.Weight)
	}

	// If totalWeight is 0 (e.g., all moves have weight 0), just return the first one to avoid panic.
	if totalWeight == 0 {
		return polyglotMoveToEngineMove(entries[0].Move, game)
	}

	r := rand.Intn(totalWeight)
	currentWeight := 0
	for _, e := range entries {
		currentWeight += int(e.Weight)
		if r < currentWeight {
			return polyglotMoveToEngineMove(e.Move, game)
		}
	}

	return polyglotMoveToEngineMove(entries[0].Move, game)
}
