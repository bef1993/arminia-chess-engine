package engine

import (
	"math/rand"
)

// Zobrist keys
var (
	zobristPiece     [2][8][64]uint64 // [Color][PieceType][Square] - Size 8 to accommodate PieceType values safely
	zobristCastling  [16]uint64       // [CastlingRights]
	zobristEnPassant [9]uint64        // [File] (8 for none)
	zobristBlackTurn uint64           // XORed if it's Black's turn
)

func init() {
	// Use a fixed seed for deterministic behavior across runs (useful for debugging)
	rng := rand.New(rand.NewSource(1070372))

	// Initialize piece keys
	for color := 0; color < 2; color++ {
		for piece := 0; piece < 8; piece++ {
			for sq := 0; sq < 64; sq++ {
				zobristPiece[color][piece][sq] = rng.Uint64()
			}
		}
	}

	// Initialize castling keys
	for i := 0; i < 16; i++ {
		zobristCastling[i] = rng.Uint64()
	}

	// Initialize en passant keys
	for i := 0; i < 9; i++ {
		zobristEnPassant[i] = rng.Uint64()
	}

	// Initialize turn key
	zobristBlackTurn = rng.Uint64()
}

// ComputeZobristHash calculates the Zobrist hash for the current game state from scratch.
func (g *Game) ComputeZobristHash() uint64 {
	var hash uint64

	// 1. Pieces
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != NoPiece {
				sq := row*8 + col
				// Assuming PieceType constants map to 0-7 range
				// Assuming Color constants: White=0, Black=1
				hash ^= zobristPiece[piece.Color()][piece.Type()][sq]
			}
		}
	}

	// 2. Castling Rights
	hash ^= zobristCastling[g.CastlingRights]

	// 3. En Passant
	if g.EnPassantTargetCol != -1 {
		hash ^= zobristEnPassant[g.EnPassantTargetCol]
	} else {
		hash ^= zobristEnPassant[8] // No EP target
	}

	// 4. Side to move
	if g.CurrentTurn == Black {
		hash ^= zobristBlackTurn
	}

	return hash
}
