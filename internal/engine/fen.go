package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func NewGameFromFEN(fen string) (*Game, error) {
	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return nil, errors.New("invalid FEN: too few fields")
	}

	g := NewEmptyGame()

	// 1. Piece placement
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, errors.New("invalid FEN: wrong number of ranks")
	}

	for r, rankStr := range ranks {
		rank := 7 - r // FEN starts from rank 8 (row 7) to rank 1 (row 0)
		file := 0
		for _, char := range rankStr {
			if unicode.IsDigit(char) {
				emptySquares, _ := strconv.Atoi(string(char))
				file += emptySquares
			} else {
				piece := NewPieceFromChar(char)
				if piece == NoPiece {
					return nil, fmt.Errorf("invalid piece char: %c", char)
				}
				g.Board.SetPiece(GetSq(file, rank), piece)
				file++
			}
		}
		if file != 8 {
			return nil, fmt.Errorf("invalid FEN: rank %d has wrong width", 8-r)
		}
	}

	// 2. Active color
	switch parts[1] {
	case "w":
		g.CurrentTurn = White
	case "b":
		g.CurrentTurn = Black
	default:
		return nil, errors.New("invalid active color")
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
	g.FullMoveCounter = 1
	if len(parts) > 5 {
		if val, err := strconv.Atoi(parts[5]); err == nil {
			g.FullMoveCounter = val
		}
	}

	// Reset history as we are starting from a position
	g.MoveHistory = []Move{}
	g.PreviousState = nil

	g.ZobristHash = g.ComputeZobristHash()
	return g, nil
}

// GenerateFEN returns the FEN string representing the current game state.
func (g *Game) GenerateFEN() string {
	var sb strings.Builder

	// 1. Piece placement
	for rank := 7; rank >= 0; rank-- {
		emptyCount := 0
		for file := 0; file < 8; file++ {
			piece := g.Board.Squares[GetSq(file, rank)]
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
		if rank > 0 {
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
		col := rune('a' + GetFile(g.EnPassantTarget))
		row := GetRank(g.EnPassantTarget) + 1
		sb.WriteString(fmt.Sprintf("%c%d", col, row))
	} else {
		sb.WriteString("-")
	}

	sb.WriteString(" ")

	// 5. Halfmove clock
	sb.WriteString(strconv.Itoa(g.HalfMoveClock))

	sb.WriteString(" ")

	// 6. Fullmove number
	sb.WriteString(strconv.Itoa(g.FullMoveCounter))

	return sb.String()
}
