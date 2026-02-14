package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// LoadFEN loads a game state from a FEN string
func (g *Game) LoadFEN(fen string) error {
	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return errors.New("invalid FEN: too few fields")
	}

	// 1. Piece placement
	g.Board.Clear()
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return errors.New("invalid FEN: wrong number of ranks")
	}

	for r, rankStr := range ranks {
		row := 7 - r // FEN starts from rank 8 (row 7) to rank 1 (row 0)
		col := 0
		for _, char := range rankStr {
			if unicode.IsDigit(char) {
				emptySquares, _ := strconv.Atoi(string(char))
				col += emptySquares
			} else {
				piece := NewPieceFromChar(char)
				if piece == NoPiece {
					return fmt.Errorf("invalid piece char: %c", char)
				}
				g.Board.SetPiece(row*8+col, piece)
				col++
			}
		}
		if col != 8 {
			return fmt.Errorf("invalid FEN: rank %d has wrong width", 8-r)
		}
	}

	// 2. Active color
	switch parts[1] {
	case "w":
		g.CurrentTurn = White
	case "b":
		g.CurrentTurn = Black
	default:
		return errors.New("invalid active color")
	}

	// 3. Castling availability
	g.CastlingRights = NoCastling
	if parts[2] != "-" {
		for _, char := range parts[2] {
			switch char {
			case 'K':
				g.CastlingRights |= WhiteKingside
			case 'Q':
				g.CastlingRights |= WhiteQueenside
			case 'k':
				g.CastlingRights |= BlackKingside
			case 'q':
				g.CastlingRights |= BlackQueenside
			default:
				// Ignore invalid chars or handle error
			}
		}
	}

	// 4. En passant target square
	g.EnPassantTarget = -1
	if parts[3] != "-" {
		sq := Sq(parts[3])
		if sq != -1 {
			g.EnPassantTarget = sq
		}
	}

	// 5. Halfmove clock (optional)
	g.HalfMoveClock = 0
	if len(parts) > 4 {
		if val, err := strconv.Atoi(parts[4]); err == nil {
			g.HalfMoveClock = val
		}
	}

	// 6. Fullmove number (optional)
	g.FullMoveNumber = 1
	if len(parts) > 5 {
		if val, err := strconv.Atoi(parts[5]); err == nil {
			g.FullMoveNumber = val
		}
	}

	// Reset history as we are starting from a position
	g.MoveHistory = []Move{}
	g.LastMove = nil

	// Reset position history
	// Compute initial hash
	g.ZobristHash = g.ComputeZobristHash()

	g.PositionHistory = []uint64{g.ZobristHash}

	return nil
}

// GenerateFEN returns the FEN string representing the current game state.
func (g *Game) GenerateFEN() string {
	var sb strings.Builder

	// 1. Piece placement
	for row := 7; row >= 0; row-- {
		emptyCount := 0
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(row*8 + col)
			if piece == NoPiece {
				emptyCount++
			} else {
				if emptyCount > 0 {
					sb.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				sb.WriteString(piece.GetChar())
			}
		}
		if emptyCount > 0 {
			sb.WriteString(strconv.Itoa(emptyCount))
		}
		if row > 0 {
			sb.WriteString("/")
		}
	}

	sb.WriteString(" ")

	// 2. Active color
	if g.CurrentTurn == White {
		sb.WriteString("w")
	} else {
		sb.WriteString("b")
	}

	sb.WriteString(" ")

	// 3. Castling availability
	if g.CastlingRights == NoCastling {
		sb.WriteString("-")
	} else {
		if g.CastlingRights&WhiteKingside != 0 {
			sb.WriteString("K")
		}
		if g.CastlingRights&WhiteQueenside != 0 {
			sb.WriteString("Q")
		}
		if g.CastlingRights&BlackKingside != 0 {
			sb.WriteString("k")
		}
		if g.CastlingRights&BlackQueenside != 0 {
			sb.WriteString("q")
		}
	}

	sb.WriteString(" ")

	// 4. En passant target square
	if g.EnPassantTarget != -1 {
		col := rune('a' + (g.EnPassantTarget % 8))
		row := 8 - (g.EnPassantTarget / 8)
		sb.WriteString(fmt.Sprintf("%c%d", col, row))
	} else {
		sb.WriteString("-")
	}

	sb.WriteString(" ")

	// 5. Halfmove clock
	sb.WriteString(strconv.Itoa(g.HalfMoveClock))

	sb.WriteString(" ")

	// 6. Fullmove number
	sb.WriteString(strconv.Itoa(g.FullMoveNumber))

	return sb.String()
}
