package engine

import "math/bits"

// Board holds bitboard representations of piece placements and occupancy for efficient move generation and evaluation.
// It also contains a "mailbox" representation of the board for fast piece lookups.
type Board struct {
	Pieces    [2][6]Bitboard // [color][pieceType] bitboards for each piece type and color
	Occupancy [3]Bitboard    // bitboard for occupied squares: [0] white, [1] black, [2] all
	Squares   [64]Piece      // "Mailbox" representation for O(1) piece lookup
}

// Board coordinates constants for readability
const (
	FileA = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	Rank1 = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

// NewBoard creates a new chess board with standard starting position
func NewBoard() Board {
	board := Board{}
	board.InitializeStartingPosition()
	return board
}

// NewEmptyBoard creates a new empty chess board with no pieces
func NewEmptyBoard() Board {
	board := Board{}
	board.Clear()
	return board
}

// Clear removes all pieces from the board
func (b *Board) Clear() {
	for i := 0; i < 6; i++ {
		b.Pieces[White][i] = 0
		b.Pieces[Black][i] = 0
	}
	for i := 0; i < 64; i++ {
		b.Squares[i] = NoPiece
	}
	b.Occupancy[White] = 0
	b.Occupancy[Black] = 0
	b.Occupancy[AnyColor] = 0
}

// InitializeStartingPosition sets up the board with the standard chess starting position
func (b *Board) InitializeStartingPosition() {
	b.Clear()

	// White Pieces
	for f := FileA; f <= FileH; f++ {
		b.SetPiece(GetSq(f, Rank2), Pawn.White())
	}
	b.SetPiece(GetSq(FileA, Rank1), Rook.White())
	b.SetPiece(GetSq(FileH, Rank1), Rook.White())
	b.SetPiece(GetSq(FileB, Rank1), Knight.White())
	b.SetPiece(GetSq(FileG, Rank1), Knight.White())
	b.SetPiece(GetSq(FileC, Rank1), Bishop.White())
	b.SetPiece(GetSq(FileF, Rank1), Bishop.White())
	b.SetPiece(GetSq(FileD, Rank1), Queen.White())
	b.SetPiece(GetSq(FileE, Rank1), King.White())

	// Black Pieces
	for f := FileA; f <= FileH; f++ {
		b.SetPiece(GetSq(f, Rank7), Pawn.Black())
	}
	b.SetPiece(GetSq(FileA, Rank8), Rook.Black())
	b.SetPiece(GetSq(FileH, Rank8), Rook.Black())
	b.SetPiece(GetSq(FileB, Rank8), Knight.Black())
	b.SetPiece(GetSq(FileG, Rank8), Knight.Black())
	b.SetPiece(GetSq(FileC, Rank8), Bishop.Black())
	b.SetPiece(GetSq(FileF, Rank8), Bishop.Black())
	b.SetPiece(GetSq(FileD, Rank8), Queen.Black())
	b.SetPiece(GetSq(FileE, Rank8), King.Black())
}

// IsOnBoard checks if a square index is valid (0-63)
func IsOnBoard(sq int) bool {
	return sq >= 0 && sq < 64
}

// IsOnBoard2D checks if file and rank are valid (0-7)
func IsOnBoard2D(file, rank int) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

// GetRank returns the rank (0-7) of a square index
func GetRank(sq int) int {
	return sq / 8
}

// GetFile returns the file (0-7) of a square index
func GetFile(sq int) int {
	return sq % 8
}

// GetSq returns the square index (0-63) for a given file and rank
func GetSq(file, rank int) int {
	return rank*8 + file
}

// SetPiece places a piece at the given position, keeping all board representations in sync.
func (b *Board) SetPiece(sq int, piece Piece) {
	if !IsOnBoard(sq) {
		return
	}

	// Get the piece currently on the square
	oldPiece := b.Squares[sq]

	// If there was a piece, clear it from the bitboards
	if oldPiece != NoPiece {
		b.Pieces[oldPiece.Color()][oldPiece.Type()].Clear(sq)
		b.Occupancy[oldPiece.Color()].Clear(sq)
	}

	// Update the Squares array with the new piece
	b.Squares[sq] = piece

	// If the new piece is not NoPiece, add it to the bitboards
	if piece != NoPiece {
		b.Pieces[piece.Color()][piece.Type()].Set(sq)
		b.Occupancy[piece.Color()].Set(sq)
	}

	// Update the combined occupancy bitboard
	b.Occupancy[AnyColor] = b.Occupancy[White] | b.Occupancy[Black]
}

// MovePiece moves a piece from source to destination
func (b *Board) MovePiece(from, to int) bool {
	piece := b.Squares[from]
	if piece == NoPiece {
		return false
	}

	b.SetPiece(to, piece)
	b.SetPiece(from, NoPiece)
	return true
}

// IsEmpty checks if a square is empty
func (b *Board) IsEmpty(sq int) bool {
	return !b.Occupancy[AnyColor].IsSet(sq)
}

// IsOccupiedByColor checks if a square is occupied by a specific color
func (b *Board) IsOccupiedByColor(sq int, color Color) bool {
	return b.Occupancy[color].IsSet(sq)
}

// HasNonPawnMaterial checks if the given color has any non-pawn pieces (Knight, Bishop, Rook, Queen).
func (b *Board) HasNonPawnMaterial(color Color) bool {
	return (b.Pieces[color][Knight] | b.Pieces[color][Bishop] | b.Pieces[color][Rook] | b.Pieces[color][Queen]) != 0
}

// FindKing locates the king of the given color and returns its position
// Returns -1 if king is not found
func (b *Board) FindKing(color Color) int {
	kingBB := b.Pieces[color][King]
	if kingBB == 0 {
		return -1
	}
	return bits.TrailingZeros64(uint64(kingBB))
}

// IsSquareAttackedByColor checks if a square can be attacked by any piece of the attacker color
func (b *Board) IsSquareAttackedByColor(sq int, attacker Color) bool {
	// Check for pawn attacks
	// We check if the square is attacked by a pawn of 'attacker' color.
	// This is equivalent to checking if 'PawnAttacks[Opponent][sq]' overlaps with 'Pieces[Attacker][Pawn]'.
	opponent := Black
	if attacker == White {
		opponent = Black
	} else {
		opponent = White
	}

	if (PawnAttacks[opponent][sq] & b.Pieces[attacker][Pawn]) != 0 {
		return true
	}

	// Check for knight attacks
	// Use pre-calculated KnightAttacks
	if (KnightAttacks[sq] & b.Pieces[attacker][Knight]) != 0 {
		return true
	}

	// Check for sliding pieces (Rook, Bishop, Queen)
	occ := b.Occupancy[AnyColor]

	// Bishop/Queen (Diagonal)
	bishopAttacks := GetBishopAttacks(sq, occ)
	if (bishopAttacks & (b.Pieces[attacker][Bishop] | b.Pieces[attacker][Queen])) != 0 {
		return true
	}

	// Rook/Queen (Straight)
	rookAttacks := GetRookAttacks(sq, occ)
	if (rookAttacks & (b.Pieces[attacker][Rook] | b.Pieces[attacker][Queen])) != 0 {
		return true
	}

	// Check for King attacks (symmetric)
	if (KingAttacks[sq] & b.Pieces[attacker][King]) != 0 {
		return true
	}

	return false
}

// GetAttackedSquares returns a bitboard of all squares attacked by the given color.
// This is useful for evaluation (mobility, king safety) but typically too slow for move generation checks.
// TODO create test and use it in evaluation (e.g., mobility, king safety)
func (b *Board) GetAttackedSquares(attacker Color) Bitboard {
	var attacks Bitboard
	occ := b.Occupancy[AnyColor]

	// Pawns
	pawns := b.Pieces[attacker][Pawn]
	if attacker == White {
		// White captures Up-Left (<<7) and Up-Right (<<9)
		// We must mask out files to prevent wrapping attacks
		attacks |= (pawns & ^FileA_BB) << 7
		attacks |= (pawns & ^FileH_BB) << 9
	} else {
		// Black captures Down-Right (>>7) and Down-Left (>>9)
		attacks |= (pawns & ^FileH_BB) >> 7
		attacks |= (pawns & ^FileA_BB) >> 9
	}

	// Knights
	knights := b.Pieces[attacker][Knight]
	for knights != 0 {
		sq := knights.PopLSB()
		attacks |= KnightAttacks[sq]
	}

	// King
	king := b.Pieces[attacker][King]
	if king != 0 {
		sq := king.PopLSB()
		attacks |= KingAttacks[sq]
	}

	// Sliding Pieces
	// We can combine Bishop+Queen and Rook+Queen to reduce loop overhead
	bishopsQueens := b.Pieces[attacker][Bishop] | b.Pieces[attacker][Queen]
	for bishopsQueens != 0 {
		sq := bishopsQueens.PopLSB()
		attacks |= GetBishopAttacks(sq, occ)
	}

	rooksQueens := b.Pieces[attacker][Rook] | b.Pieces[attacker][Queen]
	for rooksQueens != 0 {
		sq := rooksQueens.PopLSB()
		attacks |= GetRookAttacks(sq, occ)
	}

	return attacks
}

// IsKingInCheck checks if the king of the given color is in check
func (b *Board) IsKingInCheck(color Color) bool {
	kingSq := b.FindKing(color)
	if kingSq == -1 {
		// King not found (shouldn't happen in valid game)
		return false
	}

	// Determine opponent color
	opponent := Black
	if color == Black {
		opponent = White
	}

	// Check if the king's square is attacked by any opponent piece
	return b.IsSquareAttackedByColor(kingSq, opponent)
}

// Sq converts algebraic notation (e.g., "e4") to file, rank coordinates.
// Returns -1 if the string is invalid.
func Sq(s string) int {
	if len(s) != 2 {
		return -1
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if !IsOnBoard2D(file, rank) {
		return -1
	}
	return GetSq(file, rank)
}

// SetPieceAt places a piece using algebraic notation (e.g., "e4")
func (b *Board) SetPieceAt(sq string, piece Piece) {
	idx := Sq(sq)
	if idx != -1 {
		b.SetPiece(idx, piece)
	}
}

// RemovePieceAt removes a piece from the board using algebraic notation (e.g., "e4")
func (b *Board) RemovePieceAt(sq string) {
	b.SetPieceAt(sq, NoPiece)
}

// GetPieceAt retrieves a piece using algebraic notation (e.g., "e4")
func (b *Board) GetPieceAt(sq string) Piece {
	idx := Sq(sq)
	if idx == -1 {
		return NoPiece
	}
	return b.Squares[idx]
}

// IsInsufficientMaterial checks if there are enough pieces to force a checkmate
func (b *Board) IsInsufficientMaterial() bool {
	// Fast check: If there are any Pawns, Rooks, or Queens, it's not insufficient material.
	// We can check all at once using bitwise OR.
	heavyPieces := b.Pieces[White][Pawn] | b.Pieces[White][Rook] | b.Pieces[White][Queen] |
		b.Pieces[Black][Pawn] | b.Pieces[Black][Rook] | b.Pieces[Black][Queen]

	if heavyPieces != 0 {
		return false
	}

	// Count minor pieces using Population Count (very fast)
	whiteKnights := b.Pieces[White][Knight].Count()
	blackKnights := b.Pieces[Black][Knight].Count()
	whiteBishops := b.Pieces[White][Bishop].Count()
	blackBishops := b.Pieces[Black][Bishop].Count()

	whiteMinors := whiteKnights + whiteBishops
	blackMinors := blackKnights + blackBishops
	totalMinors := whiteMinors + blackMinors

	// King vs King (No minors)
	if totalMinors == 0 {
		return true
	}

	// King + Knight vs King OR King + Bishop vs King
	// (Exactly one minor piece on the board)
	if totalMinors == 1 {
		return true
	}

	// King + Bishop vs King + Bishop (same color squares)
	if whiteBishops == 1 && blackBishops == 1 && whiteKnights == 0 && blackKnights == 0 {
		// Get the square of the white bishop
		wbBB := b.Pieces[White][Bishop]
		wbSq := wbBB.PopLSB() // Safe because we know count is 1

		// Get the square of the black bishop
		bbBB := b.Pieces[Black][Bishop]
		bbSq := bbBB.PopLSB()

		wbColor := (GetRank(wbSq) + GetFile(wbSq)) % 2
		bbColor := (GetRank(bbSq) + GetFile(bbSq)) % 2

		return wbColor == bbColor
	}

	return false
}
