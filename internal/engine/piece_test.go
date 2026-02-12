package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPiece(t *testing.T) {
	tests := []struct {
		name      string
		pieceType PieceType
		color     Color
	}{
		{"White Pawn", Pawn, White},
		{"Black King", King, Black},
		{"White Queen", Queen, White},
		{"Black Rook", Rook, Black},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			piece := tt.pieceType.FromColor(tt.color)

			assert.Equal(t, tt.pieceType, piece.Type())
			assert.Equal(t, tt.color, piece.Color())
		})
	}
}

func TestPieceGetSymbol(t *testing.T) {
	tests := []struct {
		name     string
		piece    Piece
		expected string
	}{
		{"White Pawn", WhitePawn, "♙"},
		{"Black Pawn", BlackPawn, "♟"},
		{"White Knight", WhiteKnight, "♘"},
		{"Black Knight", BlackKnight, "♞"},
		{"White Bishop", WhiteBishop, "♗"},
		{"Black Bishop", BlackBishop, "♝"},
		{"White Rook", WhiteRook, "♖"},
		{"Black Rook", BlackRook, "♜"},
		{"White Queen", WhiteQueen, "♕"},
		{"Black Queen", BlackQueen, "♛"},
		{"White King", WhiteKing, "♔"},
		{"Black King", BlackKing, "♚"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol := tt.piece.GetSymbol()

			assert.Equal(t, tt.expected, symbol)
		})
	}
}
