package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard(t *testing.T) {
	board := NewBoard()
	assert.NotNil(t, board, "NewBoard returned nil")
}

func TestBoardInitializeStartingPosition(t *testing.T) {
	board := NewBoard()

	// Check white pawns on Rank 2
	for col := 0; col < 8; col++ {
		piece := board.GetPiece(col, Rank2)
		assert.NotEqual(t, NoPiece, piece)
		assert.Equal(t, Pawn, piece.Type())
		assert.Equal(t, White, piece.Color())
	}

	// Check black pawns on Rank 7
	for col := 0; col < 8; col++ {
		piece := board.GetPiece(col, Rank7)
		assert.NotEqual(t, NoPiece, piece)
		assert.Equal(t, Pawn, piece.Type())
		assert.Equal(t, Black, piece.Color())
	}

	// Check white back row
	expectedWhiteBack := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for col, expectedType := range expectedWhiteBack {
		piece := board.GetPiece(col, Rank1)
		assert.NotEqual(t, NoPiece, piece)
		assert.Equal(t, expectedType, piece.Type())
		assert.Equal(t, White, piece.Color())
	}

	// Check black back row
	expectedBlackBack := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for col, expectedType := range expectedBlackBack {
		piece := board.GetPiece(col, Rank8)
		assert.NotEqual(t, NoPiece, piece)
		assert.Equal(t, expectedType, piece.Type())
		assert.Equal(t, Black, piece.Color())
	}

	// Check middle is empty
	for row := Rank6; row <= Rank3; row++ {
		for col := 0; col < 8; col++ {
			piece := board.GetPiece(col, row)
			assert.Equal(t, NoPiece, piece, "Expected empty square at (col=%d, row=%d)", col, row)
		}
	}
}

func TestGetPiece(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		col       int
		row       int
		wantPiece bool
		wantType  PieceType
		wantColor Color
	}{
		{FileA, Rank8, true, Rook, Black},
		{FileE, Rank8, true, King, Black},
		{FileA, Rank1, true, Rook, White},
		{FileE, Rank1, true, King, White},
		{FileD, Rank4, false, 0, 0}, // Empty square in the middle
	}

	for _, tt := range tests {
		piece := board.GetPiece(tt.col, tt.row)

		if tt.wantPiece {
			assert.NotEqual(t, NoPiece, piece, "GetPiece(%d, %d) = NoPiece, want piece", tt.col, tt.row)
			assert.Equal(t, tt.wantType, piece.Type(), "Type mismatch at (%d, %d)", tt.col, tt.row)
			assert.Equal(t, tt.wantColor, piece.Color(), "Color mismatch at (%d, %d)", tt.col, tt.row)
		} else {
			assert.Equal(t, NoPiece, piece, "GetPiece(%d, %d) = %v, want NoPiece", tt.col, tt.row, piece)
		}
	}
}

func TestGetPieceOutOfBounds(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		col int
		row int
	}{
		{-1, 0},
		{8, 0},
		{0, -1},
		{0, 8},
		{10, 10},
	}

	for _, tt := range tests {
		piece := board.GetPiece(tt.col, tt.row)
		assert.Equal(t, NoPiece, piece, "GetPiece(%d, %d) = %v, want NoPiece for out of bounds", tt.col, tt.row, piece)
	}
}

func TestSetPiece(t *testing.T) {
	board := NewBoard()
	piece := WhitePawn

	board.SetPieceAt("d4", piece)

	retrieved := board.GetPieceAt("d4")
	assert.NotEqual(t, NoPiece, retrieved)
	assert.Equal(t, Pawn, retrieved.Type())
	assert.Equal(t, White, retrieved.Color())
}

func TestSetPieceOutOfBounds(t *testing.T) {
	board := NewBoard()
	piece := WhitePawn

	// Should not panic or crash
	assert.NotPanics(t, func() {
		board.SetPiece(-1, 0, piece)
		board.SetPiece(8, 0, piece)
		board.SetPiece(0, -1, piece)
		board.SetPiece(0, 8, piece)
	})
}

func TestMovePiece(t *testing.T) {
	board := NewBoard()

	// Move a white pawn from (4, 6) to (4, 4)
	success := board.MovePiece(FileE, Rank2, FileE, Rank4)

	assert.True(t, success, "MovePiece returned false, expected true")

	// Check source is now empty
	assert.True(t, board.IsEmpty(FileE, Rank2), "Source square should be empty after move")

	// Check destination has the piece
	piece := board.GetPieceAt("e4")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Pawn, piece.Type())
	assert.Equal(t, White, piece.Color())
}

func TestMovePieceFromEmpty(t *testing.T) {
	board := NewBoard()

	// Try to move from an empty square
	success := board.MovePiece(3, 4, 3, 5)

	assert.False(t, success, "MovePiece returned true, expected false for empty source")
}

func TestIsEmpty(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		col   int
		row   int
		empty bool
	}{
		{0, 0, false}, // Black rook
		{3, 4, true},  // Empty square
		{4, 6, false}, // White pawn
		{5, 5, true},  // Empty square
	}

	for _, tt := range tests {
		isEmpty := board.IsEmpty(tt.col, tt.row)
		assert.Equal(t, tt.empty, isEmpty, "IsEmpty(%d, %d)", tt.col, tt.row)
	}
}

func TestIsOccupiedByColor(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		col      int
		row      int
		color    Color
		occupied bool
	}{
		{0, 0, Black, true},  // Black rook
		{0, 0, White, false}, // Black rook, checking for white
		{4, 6, White, true},  // White pawn
		{4, 6, Black, false}, // White pawn, checking for black
		{3, 4, White, false}, // Empty square
		{3, 4, Black, false}, // Empty square
	}

	for _, tt := range tests {
		occupied := board.IsOccupiedByColor(tt.col, tt.row, tt.color)
		assert.Equal(t, tt.occupied, occupied, "IsOccupiedByColor(%d, %d, %v)", tt.col, tt.row, tt.color)
	}
}

func TestFindKing(t *testing.T) {
	tests := []struct {
		name      string
		color     Color
		setupFn   func(*Board)
		wantCol   int
		wantRow   int
		wantFound bool
	}{
		{
			name:      "find white king at starting position",
			color:     White,
			setupFn:   func(b *Board) {},
			wantCol:   4,
			wantRow:   7,
			wantFound: true,
		},
		{
			name:      "find black king at starting position",
			color:     Black,
			setupFn:   func(b *Board) {},
			wantCol:   4,
			wantRow:   0,
			wantFound: true,
		},
		{
			name:      "king moved to different square",
			color:     White,
			setupFn:   func(b *Board) { b.MovePiece(4, 7, 4, 5) },
			wantCol:   4,
			wantRow:   5,
			wantFound: true,
		},
		{
			name:      "king not found (empty board)",
			color:     White,
			setupFn:   func(b *Board) { b.Clear() },
			wantCol:   -1,
			wantRow:   -1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setupFn(board)

			col, row := board.FindKing(tt.color)

			assert.Equal(t, tt.wantCol, col, "FindKing col mismatch")
			assert.Equal(t, tt.wantRow, row, "FindKing row mismatch")
		})
	}
}

func TestIsSquareAttackedByColor(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func(*Board)
		col      int
		row      int
		attacker Color
		want     bool
	}{
		{
			name: "pawn attacks with capture piece present",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("d4", WhitePawn)
				b.SetPieceAt("c5", BlackPawn)
			},
			col:      2,
			row:      3,
			attacker: White,
			want:     true,
		},
		{
			name:     "pawn controls square in front",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhitePawn) },
			col:      3,
			row:      3,
			attacker: White,
			want:     true,
		},
		{
			name:     "rook attacks along rank",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteRook) },
			col:      6,
			row:      4,
			attacker: White,
			want:     true,
		},
		{
			name:     "rook attacks along file",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteRook) },
			col:      3,
			row:      1,
			attacker: White,
			want:     true,
		},
		{
			name:     "knight attacks L-shape",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteKnight) },
			col:      5,
			row:      5,
			attacker: White,
			want:     true,
		},
		{
			name:     "knight does not attack same square",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteKnight) },
			col:      3,
			row:      4,
			attacker: White,
			want:     false,
		},
		{
			name:     "no piece attacks empty square",
			setupFn:  func(b *Board) { b.Clear() },
			col:      3,
			row:      4,
			attacker: White,
			want:     false,
		},
		{
			name:     "bishop attacks diagonal",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteBishop) },
			col:      6,
			row:      1,
			attacker: White,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setupFn(board)

			got := board.IsSquareAttackedByColor(tt.col, tt.row, tt.attacker)

			assert.Equal(t, tt.want, got, "IsSquareAttackedByColor(%d, %d, %v)", tt.col, tt.row, tt.attacker)
		})
	}
}

func TestIsKingInCheck(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*Board)
		color   Color
		want    bool
	}{
		{
			name:    "king not in check at start",
			setupFn: func(b *Board) {},
			color:   White,
			want:    false,
		},
		{
			name: "king in check from rook",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("e4", WhiteKing)
				b.SetPieceAt("e7", BlackRook)
			},
			color: White,
			want:  true,
		},
		{
			name: "king not in check when rook is blocked",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("e4", WhiteKing)
				b.SetPieceAt("e7", BlackRook)
				b.SetPieceAt("e6", BlackPawn)
			},
			color: White,
			want:  false,
		},
		{
			name: "king in check from pawn",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("e4", WhiteKing)
				b.SetPieceAt("d5", BlackPawn)
			},
			color: White,
			want:  true,
		},
		{
			name: "king in check from knight",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("e4", WhiteKing)
				b.SetPieceAt("g3", BlackKnight)
			},
			color: White,
			want:  true,
		},
		{
			name: "black king in check",
			setupFn: func(b *Board) {
				b.Clear()
				b.SetPieceAt("e8", BlackKing)
				b.SetPieceAt("e3", WhiteRook)
			},
			color: Black,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setupFn(board)

			got := board.IsKingInCheck(tt.color)

			assert.Equal(t, tt.want, got, "IsKingInCheck(%v)", tt.color)
		})
	}
}
