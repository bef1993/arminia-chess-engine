package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZobristInitialConsistency(t *testing.T) {
	game := NewGame()
	computed := game.ComputeZobristHash()
	assert.Equal(t, computed, game.ZobristHash, "Initial hash should match computed hash")
}

func TestZobristIncrementalUpdate(t *testing.T) {
	game := NewGame()

	// A sequence of moves covering various mechanics:
	// Pawn moves, Knight moves, Bishop moves, Castling, Captures
	moves := []string{
		"e2e4", // Pawn move
		"e7e5",
		"g1f3", // Knight move
		"b8c6",
		"f1c4", // Bishop move
		"g8f6",
		"e1g1", // Castling (White Kingside)
		"f8e7",
		"d2d4", // Pawn push
		"e5d4", // Capture
		"f3d4", // Capture
		"d7d5", // Pawn push
		"e4d5", // Capture
		"c6d4", // Capture
		"d1d4", // Queen capture
	}

	for _, moveStr := range moves {
		move, err := ParseMove(moveStr, game)
		assert.NoError(t, err)

		success := game.ExecuteMove(move)
		assert.True(t, success)

		computed := game.ComputeZobristHash()
		assert.Equal(t, computed, game.ZobristHash, "Hash mismatch after move %s", moveStr)
	}
}

func TestZobristUndo(t *testing.T) {
	game := NewGame()
	initialHash := game.ZobristHash

	move := NewMove(Sq("e2"), Sq("e4"))
	game.ExecuteMove(move)

	assert.NotEqual(t, initialHash, game.ZobristHash, "Hash should change after move")

	game.UnmakeMove()
	assert.Equal(t, initialHash, game.ZobristHash, "Hash should be restored after undo")
}

func TestZobristTransposition(t *testing.T) {
	// Position: d3 d6, Nf3 Nc6
	// We use single pawn pushes to avoid En Passant target differences affecting the hash.
	// Path 1: 1. d3 d6 2. Nf3 Nc6
	// Path 2: 1. Nf3 Nc6 2. d3 d6

	// Path 1
	g1 := NewGame()
	moves1 := []string{"d2d3", "d7d6", "g1f3", "b8c6"}
	for _, m := range moves1 {
		move, _ := ParseMove(m, g1)
		g1.ExecuteMove(move)
	}
	hash1 := g1.ZobristHash

	// Path 2
	g2 := NewGame()
	moves2 := []string{"g1f3", "b8c6", "d2d3", "d7d6"}
	for _, m := range moves2 {
		move, _ := ParseMove(m, g2)
		g2.ExecuteMove(move)
	}
	hash2 := g2.ZobristHash

	assert.Equal(t, hash1, hash2, "Transpositions should have identical hashes")
}

func TestZobristSideToMove(t *testing.T) {
	g1 := NewGame()
	// White to move
	h1 := g1.ZobristHash

	// Force change side to move without changing pieces
	g1.CurrentTurn = Black
	h2 := g1.ComputeZobristHash()

	assert.NotEqual(t, h1, h2, "Hash should differ by side to move")
}

func TestZobristCastlingRights(t *testing.T) {
	g1 := NewGame()
	h1 := g1.ZobristHash

	g1.CastlingRights = NoCastling
	h2 := g1.ComputeZobristHash()

	assert.NotEqual(t, h1, h2, "Hash should differ by castling rights")
}

func TestZobristEnPassantCapture(t *testing.T) {
	// Set up a position where EP capture is possible
	// White pawn on e5, Black pawn on d5 (just moved d7-d5)
	game := NewEmptyGame()
	game.Board.SetPieceAt("e5", WhitePawn)
	game.Board.SetPieceAt("d5", BlackPawn)
	game.CurrentTurn = White
	game.EnPassantTarget = Sq("d6") // Target square for EP capture

	// Initial hash
	game.ZobristHash = game.ComputeZobristHash()

	// Execute EP capture: e5xd6
	move := NewMove(Sq("e5"), Sq("d6"))
	game.ExecuteMove(move)

	computed := game.ComputeZobristHash()
	assert.Equal(t, computed, game.ZobristHash, "Hash mismatch after En Passant capture")

	// Undo
	game.UnmakeMove()
	computedUndo := game.ComputeZobristHash()
	assert.Equal(t, computedUndo, game.ZobristHash, "Hash mismatch after Undo EP capture")
}

func TestZobristPromotion(t *testing.T) {
	game := NewGame()
	game.Board.Clear()
	game.Board.SetPieceAt("e7", WhitePawn)
	game.CurrentTurn = White

	game.ZobristHash = game.ComputeZobristHash()

	// e7e8q
	move := NewPromotionMove(Sq("e7"), Sq("e8"), WhiteQueen)
	game.ExecuteMove(move)

	computed := game.ComputeZobristHash()
	assert.Equal(t, computed, game.ZobristHash, "Hash mismatch after promotion")
}

func TestZobristHashConsistency(t *testing.T) {
	game := NewGame()

	// Store hashes: initial hash + hash after each move
	var hashes []uint64
	hashes = append(hashes, game.ZobristHash)

	// The Opera Game (Morphy vs. Duke of Brunswick/Count Isouard, 1858)
	moves := []string{
		"e2e4", "e7e5",
		"g1f3", "d7d6",
		"d2d4", "c8g4",
		"d4e5", "g4f3",
		"d1f3", "d6e5",
		"f1c4", "g8f6",
		"f3b3", "d8e7",
		"b1c3", "c7c6",
		"c1g5", "b7b5",
		"c3b5", "c6b5",
		"c4b5", "b8d7",
		"e1c1", "a8d8",
		"d1d7", "d8d7",
		"h1d1", "e7e6",
		"b5d7", "f6d7",
		"b3b8", "d7b8",
		"d1d8", // Checkmate
	}

	// Execute all moves and store hashes
	for _, moveStr := range moves {
		move, err := ParseMove(moveStr, game)
		assert.NoError(t, err, "Failed to parse move %s", moveStr)
		success := game.ExecuteMove(move)
		assert.True(t, success, "Failed to execute move %s", moveStr)

		hashes = append(hashes, game.ZobristHash)
	}

	// Undo all moves and verify hashes match history
	// We start undoing the last move.
	// After undoing the last move (index len-1), the hash should match hashes[len-1].
	// We iterate backwards through the moves.
	for i := len(moves) - 1; i >= 0; i-- {
		game.UnmakeMove()

		expectedHash := hashes[i] // Hash BEFORE the move at index i was made
		assert.Equal(t, expectedHash, game.ZobristHash, "Hash mismatch after undoing move %s (index %d)", moves[i], i)

		// Also verify against computed hash to ensure state is consistent
		computed := game.ComputeZobristHash()
		assert.Equal(t, computed, game.ZobristHash, "Computed hash mismatch after undoing move %s", moves[i])
	}
}
