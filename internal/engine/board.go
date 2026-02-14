package engine

import "math/bits"

// Board holds bitboard representations of piece placements and occupancy for efficient move generation and evaluation.
type Board struct {
	Pieces    [2][6]Bitboard // [color][pieceType] bitboards for each piece type and color
	Occupancy [3]Bitboard    // bitboard for occupied squares: [0] white, [1] black, [2] all
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
func NewBoard() *Board {
	board := &Board{}
	board.InitializeStartingPosition()
	return board
}

// NewEmptyBoard creates a new empty chess board with no pieces
func NewEmptyBoard() *Board {
	board := &Board{}
	board.Clear()
	return board
}

// Clear removes all pieces from the board
func (b *Board) Clear() {
	for i := 0; i < 6; i++ {
		b.Pieces[White][i] = 0
		b.Pieces[Black][i] = 0
	}
	b.Occupancy[White] = 0
	b.Occupancy[Black] = 0
	b.Occupancy[AnyColor] = 0
}

// InitializeStartingPosition sets up the board with the standard chess starting position
func (b *Board) InitializeStartingPosition() {
	b.Clear()

	// White Pieces
	b.Pieces[White][Pawn] = 0x000000000000FF00 // Rank 2
	b.Pieces[White][Rook].Set(A1)
	b.Pieces[White][Rook].Set(H1)
	b.Pieces[White][Knight].Set(B1)
	b.Pieces[White][Knight].Set(G1)
	b.Pieces[White][Bishop].Set(C1)
	b.Pieces[White][Bishop].Set(F1)
	b.Pieces[White][Queen].Set(D1)
	b.Pieces[White][King].Set(E1)

	// Black Pieces
	b.Pieces[Black][Pawn] = 0x00FF000000000000 // Rank 7
	b.Pieces[Black][Rook].Set(A8)
	b.Pieces[Black][Rook].Set(H8)
	b.Pieces[Black][Knight].Set(B8)
	b.Pieces[Black][Knight].Set(G8)
	b.Pieces[Black][Bishop].Set(C8)
	b.Pieces[Black][Bishop].Set(F8)
	b.Pieces[Black][Queen].Set(D8)
	b.Pieces[Black][King].Set(E8)

	// Update Occupancy
	for i := 0; i < 6; i++ {
		b.Occupancy[White] |= b.Pieces[White][i]
		b.Occupancy[Black] |= b.Pieces[Black][i]
	}

	b.Occupancy[AnyColor] = b.Occupancy[White] | b.Occupancy[Black]
}

// GetPiece returns the piece at the given position
func (b *Board) GetPiece(sq int) Piece {
	if sq < 0 || sq >= 64 {
		// TODO maybe panic?
		return NoPiece
	}
	if !b.Occupancy[AnyColor].IsSet(sq) {
		return NoPiece
	}

	// Check White
	if b.Occupancy[White].IsSet(sq) {
		for i := 0; i < 6; i++ {
			if b.Pieces[White][i].IsSet(sq) {
				return PieceType(i).White()
			}
		}
	} else {
		for i := 0; i < 6; i++ {
			if b.Pieces[Black][i].IsSet(sq) {
				return PieceType(i).Black()
			}
		}
	}
	// TODO return panic?
	return NoPiece
}

// SetPiece places a piece at the given position
func (b *Board) SetPiece(sq int, piece Piece) {
	if sq < 0 || sq >= 64 {
		return // TODO maybe panic
	}

	// Clear existing piece
	if b.Occupancy[AnyColor].IsSet(sq) {
		b.Occupancy[White].Clear(sq)
		b.Occupancy[Black].Clear(sq)
		b.Occupancy[AnyColor].Clear(sq)

		for i := 0; i < 6; i++ {
			b.Pieces[White][i].Clear(sq)
			b.Pieces[Black][i].Clear(sq)
		}
	}

	if piece == NoPiece {
		return
	}

	// Set new piece
	color := piece.Color()
	pieceType := piece.Type()

	b.Pieces[color][pieceType].Set(sq)
	b.Occupancy[color].Set(sq)
	b.Occupancy[AnyColor].Set(sq)
}

// MovePiece moves a piece from source to destination
func (b *Board) MovePiece(from, to int) bool {
	piece := b.GetPiece(from)
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
	// Temporary: Convert back to col/row for sliding logic until Magic Bitboards are implemented
	col, row := sq%8, sq/8

	// Check for pawn attacks
	pawnCheckRow := row + 1 // If attacker is Black, check for pawns on row+1 (attacking down)
	if attacker == White {
		pawnCheckRow = row - 1 // If attacker is White, check for pawns on row-1 (attacking up)
	}
	if pawnCheckRow >= 0 && pawnCheckRow < 8 {
		if col > 0 {
			p := b.GetPiece(pawnCheckRow*8 + col - 1)
			if p.Type() == Pawn && p.Color() == attacker {
				return true
			}
		}
		if col < 7 {
			p := b.GetPiece(pawnCheckRow*8 + col + 1)
			if p.Type() == Pawn && p.Color() == attacker {
				return true
			}
		}
	}

	// Check for knight attacks
	// Use pre-calculated KnightAttacks
	if (KnightAttacks[sq] & b.Pieces[attacker][Knight]) != 0 {
		return true
	}

	// Check for sliding pieces (Rook, Bishop, Queen) and King attacks
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}, {-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for i, dir := range directions {
		for dist := 1; dist < 8; dist++ {
			checkRow := row + dist*dir[0]
			checkCol := col + dist*dir[1]
			if checkRow < 0 || checkRow >= 8 || checkCol < 0 || checkCol >= 8 {
				break // Off board
			}
			p := b.GetPiece(checkRow*8 + checkCol)
			if p != NoPiece {
				if p.Color() == attacker {
					pt := p.Type()
					isDiagonal := i < 4
					isStraight := i >= 4
					if pt == Queen || (pt == Bishop && isDiagonal) || (pt == Rook && isStraight) {
						return true
					}
				}
				break // Blocked by a piece
			}
		}
	}

	// Check for King attacks (symmetric)
	if (KingAttacks[sq] & b.Pieces[attacker][King]) != 0 {
		return true
	}

	return false
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

// Sq converts algebraic notation (e.g., "e4") to col, row coordinates.
// Returns -1 if the string is invalid.
func Sq(s string) int {
	if len(s) != 2 {
		return -1
	}
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		return -1
	}
	return row*8 + col
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
	return b.GetPiece(Sq(sq))
}
