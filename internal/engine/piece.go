package engine

// Color represents the color of a piece
type Color int

const (
	White Color = iota
	Black
)

// PieceType represents the kind of piece
type PieceType int

const (
	NoType PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Piece values for standard material counting and move ordering
const (
	PawnValue   = 100
	KnightValue = 320
	BishopValue = 330
	RookValue   = 500
	QueenValue  = 900
	KingValue   = 20000 // Represents "Infinity" (Checkmate)
)

func (p PieceType) White() Piece {
	return Piece(p)
}

func (p PieceType) Black() Piece {
	return Piece(p + King)
}

func (p PieceType) FromColor(color Color) Piece {
	if color == White {
		return p.White()
	}
	return p.Black()
}

// Piece represents the type of chess piece
type Piece uint8

const (
	NoPiece Piece = iota
	WhitePawn
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

// CastlingRights represents which castling moves are available (bitmask)
type CastlingRights int

const (
	NoCastling     CastlingRights = 0
	WhiteKingside  CastlingRights = 1 << 0 // 0001
	WhiteQueenside CastlingRights = 1 << 1 // 0010
	BlackKingside  CastlingRights = 1 << 2 // 0100
	BlackQueenside CastlingRights = 1 << 3 // 1000
	AllCastling    CastlingRights = WhiteKingside | WhiteQueenside | BlackKingside | BlackQueenside
)

// Type returns the type of the piece
func (p Piece) Type() PieceType {
	switch p {
	case WhitePawn, BlackPawn:
		return Pawn
	case WhiteKnight, BlackKnight:
		return Knight
	case WhiteBishop, BlackBishop:
		return Bishop
	case WhiteRook, BlackRook:
		return Rook
	case WhiteQueen, BlackQueen:
		return Queen
	case WhiteKing, BlackKing:
		return King
	}
	return NoType
}

// Color returns the color of the piece
func (p Piece) Color() Color {
	if p >= BlackPawn && p <= BlackKing {
		return Black
	}
	return White
}

// GetSymbol returns the symbol representing this piece
func (p Piece) GetSymbol() string {
	switch p {
	case WhitePawn:
		return "♙"
	case WhiteKnight:
		return "♘"
	case WhiteBishop:
		return "♗"
	case WhiteRook:
		return "♖"
	case WhiteQueen:
		return "♕"
	case WhiteKing:
		return "♔"
	case BlackPawn:
		return "♟"
	case BlackKnight:
		return "♞"
	case BlackBishop:
		return "♝"
	case BlackRook:
		return "♜"
	case BlackQueen:
		return "♛"
	case BlackKing:
		return "♚"
	}
	return " "
}

// Value returns the standard material value of the piece
func (p Piece) Value() int {
	switch p.Type() {
	case Pawn:
		return PawnValue
	case Knight:
		return KnightValue
	case Bishop:
		return BishopValue
	case Rook:
		return RookValue
	case Queen:
		return QueenValue
	case King:
		return KingValue
	default:
		return 0
	}
}
