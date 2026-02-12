package engine

// Board represents the 8x8 chess board
type Board struct {
	Squares [8][8]Piece
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
	Rank8 = iota
	Rank7
	Rank6
	Rank5
	Rank4
	Rank3
	Rank2
	Rank1
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
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			b.Squares[row][col] = NoPiece
		}
	}
}

// InitializeStartingPosition sets up the board with the standard chess starting position
func (b *Board) InitializeStartingPosition() {
	// Clear the board first
	b.Clear()

	// Place pawns
	for col := 0; col < 8; col++ {
		b.Squares[Rank7][col] = BlackPawn
		b.Squares[Rank2][col] = WhitePawn
	}

	// Place back row pieces
	b.Squares[Rank8][FileA] = BlackRook
	b.Squares[Rank8][FileB] = BlackKnight
	b.Squares[Rank8][FileC] = BlackBishop
	b.Squares[Rank8][FileD] = BlackQueen
	b.Squares[Rank8][FileE] = BlackKing
	b.Squares[Rank8][FileF] = BlackBishop
	b.Squares[Rank8][FileG] = BlackKnight
	b.Squares[Rank8][FileH] = BlackRook

	b.Squares[Rank1][FileA] = WhiteRook
	b.Squares[Rank1][FileB] = WhiteKnight
	b.Squares[Rank1][FileC] = WhiteBishop
	b.Squares[Rank1][FileD] = WhiteQueen
	b.Squares[Rank1][FileE] = WhiteKing
	b.Squares[Rank1][FileF] = WhiteBishop
	b.Squares[Rank1][FileG] = WhiteKnight
	b.Squares[Rank1][FileH] = WhiteRook
}

// GetPiece returns the piece at the given position
func (b *Board) GetPiece(col, row int) Piece {
	if row < 0 || row >= 8 || col < 0 || col >= 8 {
		return NoPiece
	}
	return b.Squares[row][col]
}

// SetPiece places a piece at the given position
func (b *Board) SetPiece(col, row int, piece Piece) {
	if col >= 0 && col < 8 && row >= 0 && row < 8 {
		b.Squares[row][col] = piece
	}
}

// MovePiece moves a piece from source to destination
func (b *Board) MovePiece(fromCol, fromRow, toCol, toRow int) bool {
	piece := b.GetPiece(fromCol, fromRow)
	if piece == NoPiece {
		return false
	}

	b.SetPiece(toCol, toRow, piece)
	b.SetPiece(fromCol, fromRow, NoPiece)
	return true
}

// IsEmpty checks if a square is empty
func (b *Board) IsEmpty(col, row int) bool {
	return b.GetPiece(col, row) == NoPiece
}

// IsOccupiedByColor checks if a square is occupied by a specific color
func (b *Board) IsOccupiedByColor(col, row int, color Color) bool {
	piece := b.GetPiece(col, row)
	return piece != NoPiece && piece.Color() == color
}

// FindKing locates the king of the given color and returns its position
// Returns (-1, -1) if king is not found
func (b *Board) FindKing(color Color) (int, int) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.GetPiece(col, row)
			if piece != NoPiece && piece.Type() == King && piece.Color() == color {
				return col, row
			}
		}
	}
	return -1, -1
}

// IsSquareAttackedByColor checks if a square can be attacked by any piece of the attacker color
func (b *Board) IsSquareAttackedByColor(col, row int, attacker Color) bool {
	mg := NewMoveGenerator(b)

	for attackerRow := 0; attackerRow < 8; attackerRow++ {
		for attackerCol := 0; attackerCol < 8; attackerCol++ {
			piece := b.GetPiece(attackerCol, attackerRow)
			if piece != NoPiece && piece.Color() == attacker {
				// Generate all moves for this attacking piece
				moves := mg.GenerateMovesForPiece(attackerCol, attackerRow)

				// Check if any of these moves can reach the target square
				for _, move := range moves {
					if move.ToCol == col && move.ToRow == row {
						return true
					}
				}
			}
		}
	}

	return false
}

// IsKingInCheck checks if the king of the given color is in check
func (b *Board) IsKingInCheck(color Color) bool {
	kingCol, kingRow := b.FindKing(color)
	if kingCol == -1 || kingRow == -1 {
		// King not found (shouldn't happen in valid game)
		return false
	}

	// Determine opponent color
	opponent := Black
	if color == Black {
		opponent = White
	}

	// Check if the king's square is attacked by any opponent piece
	return b.IsSquareAttackedByColor(kingCol, kingRow, opponent)
}

// Sq converts algebraic notation (e.g., "e4") to col, row coordinates.
// Returns -1, -1 if the string is invalid.
func Sq(s string) (int, int) {
	if len(s) != 2 {
		return -1, -1
	}
	col := int(s[0] - 'a')
	row := 8 - int(s[1]-'0')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		return -1, -1
	}
	return col, row
}

// SetPieceAt places a piece using algebraic notation (e.g., "e4")
func (b *Board) SetPieceAt(sq string, piece Piece) {
	col, row := Sq(sq)
	if col != -1 && row != -1 {
		b.SetPiece(col, row, piece)
	}
}

// RemovePieceAt removes a piece from the board using algebraic notation (e.g., "e4")
func (b *Board) RemovePieceAt(sq string) {
	col, row := Sq(sq)
	b.SetPiece(col, row, NoPiece)
}


// GetPieceAt retrieves a piece using algebraic notation (e.g., "e4")
func (b *Board) GetPieceAt(sq string) Piece {
	col, row := Sq(sq)
	return b.GetPiece(col, row)
}
