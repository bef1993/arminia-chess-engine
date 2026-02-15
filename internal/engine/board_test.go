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
	for file := 0; file < 8; file++ {
		sq := GetSq(file, Rank2)
		piece := board.GetPiece(sq)
		assert.Equal(t, WhitePawn, piece, "Expected WhitePawn at %d", sq)
	}

	// Check black pawns on Rank 7
	for file := 0; file < 8; file++ {
		sq := GetSq(file, Rank7)
		piece := board.GetPiece(sq)
		assert.Equal(t, BlackPawn, piece, "Expected BlackPawn at %d", sq)
	}

	// Check white back row
	expectedWhiteBack := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for file, expectedType := range expectedWhiteBack {
		sq := GetSq(file, Rank1)
		piece := board.GetPiece(sq)
		assert.Equal(t, expectedType, piece.Type(), "Expected %v at %d", expectedType, sq)
		assert.Equal(t, White, piece.Color())
	}

	// Check black back row
	expectedBlackBack := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for file, expectedType := range expectedBlackBack {
		sq := GetSq(file, Rank8)
		piece := board.GetPiece(sq)
		assert.Equal(t, expectedType, piece.Type(), "Expected %v at %d", expectedType, sq)
		assert.Equal(t, Black, piece.Color())
	}

	// Check middle is empty
	for rank := Rank3; rank <= Rank6; rank++ {
		for file := 0; file < 8; file++ {
			sq := GetSq(file, rank)
			piece := board.GetPiece(sq)
			assert.Equal(t, NoPiece, piece, "Expected empty square at (file=%d, rank=%d)", file, rank)
		}
	}
}

func TestGetPieceOutOfBounds(t *testing.T) {
	board := NewBoard()

	tests := []int{
		-1,
		64,
		100,
	}

	for _, sq := range tests {
		piece := board.GetPiece(sq)
		assert.Equal(t, NoPiece, piece, "GetPiece(%d) = %v, want NoPiece for out of bounds", sq, piece)
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
		board.SetPiece(-1, piece)
		board.SetPiece(64, piece)
	})
}

func TestMovePiece(t *testing.T) {
	board := NewBoard()

	success := board.MovePiece(Sq("e2"), Sq("e4"))

	assert.True(t, success, "MovePiece returned false, expected true")

	// Check source is now empty
	assert.True(t, board.IsEmpty(Sq("e2")), "Source square should be empty after move")
	assert.Equal(t, WhitePawn, board.GetPieceAt("e4"), "Destination square should have the moved piece")
}

func TestMovePieceFromEmpty(t *testing.T) {
	board := NewBoard()

	// Try to move from an empty square
	success := board.MovePiece(Sq("e3"), Sq("e4"))

	assert.False(t, success, "MovePiece returned true, expected false for empty source")
}

func TestIsOccupiedByColor(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		sq       int
		color    Color
		occupied bool
	}{
		{Sq("a8"), Black, true},  // Black rook
		{Sq("a1"), White, true},  // White rook
		{Sq("e2"), White, true},  // White pawn
		{Sq("e7"), Black, true},  // Black Pawn
		{Sq("d3"), White, false}, // Empty square
		{Sq("d3"), Black, false}, // Empty square
		{Sq("a1"), Black, false}, // Occupied by white, not black
	}

	for _, tt := range tests {
		occupied := board.IsOccupiedByColor(tt.sq, tt.color)
		assert.Equal(t, tt.occupied, occupied, "IsOccupiedByColor(%d, %v)", tt.sq, tt.color)
	}
}

func TestFindKing(t *testing.T) {
	tests := []struct {
		name      string
		color     Color
		setupFn   func(*Board)
		wantSq    int
		wantFound bool
	}{
		{
			name:      "find white king at starting position",
			color:     White,
			setupFn:   func(b *Board) {},
			wantSq:    Sq("e1"),
			wantFound: true,
		},
		{
			name:      "find black king at starting position",
			color:     Black,
			setupFn:   func(b *Board) {},
			wantSq:    Sq("e8"),
			wantFound: true,
		},
		{
			name:      "king moved to different square",
			color:     White,
			setupFn:   func(b *Board) { b.MovePiece(Sq("e1"), Sq("e5")) },
			wantSq:    Sq("e5"),
			wantFound: true,
		},
		{
			name:      "king not found (empty board)",
			color:     White,
			setupFn:   func(b *Board) { b.Clear() },
			wantSq:    -1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setupFn(board)

			sq := board.FindKing(tt.color)

			if tt.wantFound {
				assert.NotEqual(t, -1, sq, "FindKing should have found king")
				assert.Equal(t, tt.wantSq, sq, "FindKing sq mismatch")
			} else {
				assert.Equal(t, -1, sq, "FindKing should not have found king")
			}
		})
	}
}

func TestIsSquareAttackedByColor(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func(*Board)
		sq       string
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
			sq:       "c5",
			attacker: White,
			want:     true,
		},
		{
			name:     "pawn does not attack square in front",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhitePawn) },
			sq:       "d5",
			attacker: White,
			want:     false,
		},
		{
			name:     "rook attacks along rank",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteRook) },
			sq:       "g4",
			attacker: White,
			want:     true,
		},
		{
			name:     "rook attacks along file",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteRook) },
			sq:       "d2",
			attacker: White,
			want:     true,
		},
		{
			name:     "knight attacks L-shape",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteKnight) },
			sq:       "f5",
			attacker: White,
			want:     true,
		},
		{
			name:     "knight does not attack same square",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteKnight) },
			sq:       "d5",
			attacker: White,
			want:     false,
		},
		{
			name:     "no piece attacks empty square",
			setupFn:  func(b *Board) { b.Clear() },
			sq:       "d5",
			attacker: White,
			want:     false,
		},
		{
			name:     "bishop attacks diagonal",
			setupFn:  func(b *Board) { b.Clear(); b.SetPieceAt("d4", WhiteBishop) },
			sq:       "f2",
			attacker: White,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setupFn(board)

			got := board.IsSquareAttackedByColor(Sq(tt.sq), tt.attacker)

			assert.Equal(t, tt.want, got, "IsSquareAttackedByColor(%s, %v)", tt.sq, tt.attacker)
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
