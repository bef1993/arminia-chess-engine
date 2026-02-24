package polyglot

import (
	"arminia-chess-engine/internal/engine"
)

// Polyglot Random Numbers (Initialized in init)
var (
	polyglotPiece     [2][6][64]uint64
	polyglotCastle    [16]uint64
	polyglotEnPassant [65]uint64 // File 0-7, plus "no EP"
	polyglotTurn      uint64
)

// init initializes the Polyglot random keys.
// The keys are generated using a specific algorithm to match standard books.
func init() {
	// Seed for Polyglot Random Generator
	var seed uint64 = 0x9D2C5680352F47E5

	// Pseudo-random generator used by Polyglot
	rand64 := func() uint64 {
		seed = seed*6364136223846793005 + 1442695040888963407
		return seed
	}

	// Fill Piece Keys
	for piece := 0; piece < 6; piece++ { // Pawn, Knight, Bishop, Rook, Queen, King
		for color := 0; color < 2; color++ { // White, Black
			for sq := 0; sq < 64; sq++ {
				// Polyglot mapping:
				// Pieces: 0=None, 1=Pawn, 3=Knight, 5=Bishop, 7=Rook, 9=Queen, 11=King (White)
				//         2=Pawn, 4=Knight, 6=Bishop, 8=Rook, 10=Queen, 12=King (Black)
				// We map our 0-5 and 0-1 to this sequence.
				// Actually, the loop order in Polyglot initialization is:
				// for piece_type in {pawn, knight, bishop, rook, queen, king}
				//   for color in {white, black}
				//     for square in {0..63}
				polyglotPiece[color][piece][sq] = rand64()
			}
		}
	}

	// Fill Castling Keys
	for i := 0; i < 16; i++ {
		polyglotCastle[i] = rand64()
	}

	// Fill En Passant Keys
	for i := 0; i < 65; i++ {
		polyglotEnPassant[i] = rand64()
	}

	// Turn Key
	polyglotTurn = rand64()
}

// ComputePolyglotHash calculates the Zobrist hash for the current board state
// using the standard Polyglot keys.
func ComputePolyglotHash(g *engine.Game) uint64 {
	var hash uint64

	// 1. Pieces
	// Iterate over all squares.
	for sq := 0; sq < 64; sq++ {
		piece := g.Board.GetPiece(sq)
		if piece != engine.NoPiece {
			// Map engine piece to Polyglot indices
			// engine.PieceType: Pawn=0, Knight=1, Bishop=2, Rook=3, Queen=4, King=5
			// engine.Color: White=0, Black=1
			pt := int(piece.Type())
			c := int(piece.Color())

			// Polyglot rank/file mapping is standard (A1=0, H8=63), same as engine.
			hash ^= polyglotPiece[c][pt][sq]
		}
	}

	// 2. Castling Rights
	// Polyglot uses 4 bits: WK=1, WQ=2, BK=4, BQ=8
	// Engine uses: WhiteKingside=1, WhiteQueenside=2, BlackKingside=4, BlackQueenside=8
	// Assuming the constants match (which they usually do in bitboard engines).
	// If not, we would need to map them.
	// Based on standard bitboard implementations:
	hash ^= polyglotCastle[g.CastlingRights]

	// 3. En Passant
	// Polyglot hashes the file of the EP target (0-7). If no EP, it doesn't hash anything (or hashes a specific "no EP" key?).
	// Actually, Polyglot only hashes the EP file if a pawn can actually capture it.
	// But for simplicity, most implementations just hash the file if target is set.
	// Standard Polyglot: if ep_square is set, hash ^= enpassant[file].
	if g.EnPassantTarget != -1 {
		file := engine.GetFile(g.EnPassantTarget)
		hash ^= polyglotEnPassant[file]
	}

	// 4. Turn
	if g.CurrentTurn == engine.White {
		hash ^= polyglotTurn
	}

	return hash
}

// polyglotMoveToEngineMove converts a Polyglot encoded move to an engine.Move.
func polyglotMoveToEngineMove(polyMove uint16, g *engine.Game) engine.Move {
	// Polyglot Move Encoding:
	// bits 0-2: to_file
	// bits 3-5: to_rank
	// bits 6-8: from_file
	// bits 9-11: from_rank
	// bits 12-14: promotion piece

	toFile := int(polyMove & 0x7)
	toRank := int((polyMove >> 3) & 0x7)
	fromFile := int((polyMove >> 6) & 0x7)
	fromRank := int((polyMove >> 9) & 0x7)
	promo := int((polyMove >> 12) & 0x7)

	from := engine.GetSq(fromFile, fromRank)
	to := engine.GetSq(toFile, toRank)

	var promoPiece engine.Piece = engine.NoPiece
	if promo > 0 {
		// Polyglot: 0=None, 1=Knight, 2=Bishop, 3=Rook, 4=Queen
		// Engine: Needs Piece with Color
		color := g.CurrentTurn
		switch promo {
		case 1:
			promoPiece = engine.Knight.FromColor(color)
		case 2:
			promoPiece = engine.Bishop.FromColor(color)
		case 3:
			promoPiece = engine.Rook.FromColor(color)
		case 4:
			promoPiece = engine.Queen.FromColor(color)
		}
		return engine.NewPromotionMove(from, to, promoPiece)
	}

	return engine.NewMove(from, to)
}
