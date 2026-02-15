package engine

import (
	"fmt"
)

// Move represents a chess move (pure data)
type Move struct {
	From           int
	To             int
	PromotionPiece Piece // 0 if no promotion
}

// NewMove creates a move without promotion
func NewMove(from, to int) Move {
	return Move{
		From:           from,
		To:             to,
		PromotionPiece: NoPiece, // No promotion (Pawn constant)
	}
}

// NewPromotionMove creates a pawn promotion move
func NewPromotionMove(from, to int, promotionPiece Piece) Move {
	return Move{
		From:           from,
		To:             to,
		PromotionPiece: promotionPiece,
	}
}

// String returns the UCI representation of the move.
func (m Move) String() string {
	if m == (Move{}) {
		return "0000"
	}
	moveStr := fmt.Sprintf("%c%d%c%d",
		rune('a'+GetFile(m.From)),
		GetRank(m.From)+1,
		rune('a'+GetFile(m.To)),
		GetRank(m.To)+1)

	if m.PromotionPiece != NoPiece {
		switch m.PromotionPiece.Type() {
		case Queen:
			moveStr += "q"
		case Rook:
			moveStr += "r"
		case Bishop:
			moveStr += "b"
		case Knight:
			moveStr += "n"
		default:
		}
	}
	return moveStr
}

// ParseMove parses a UCI move string (e.g., "e2e4", "a7a8q") into a Move struct.
// It validates the move against the current game state (legal moves).
func ParseMove(moveStr string, game *Game) (Move, error) { // TODO remove Validation from ParseMove
	if len(moveStr) < 4 {
		return Move{}, fmt.Errorf("invalid move format: %s", moveStr)
	}

	from := Sq(moveStr[0:2])
	to := Sq(moveStr[2:4])

	if from == -1 || to == -1 {
		return Move{}, fmt.Errorf("square out of bounds: %s", moveStr)
	}

	move := NewMove(from, to)

	if len(moveStr) == 5 {
		switch moveStr[4] {
		case 'q':
			move.PromotionPiece = Queen.FromColor(game.CurrentTurn)
		case 'r':
			move.PromotionPiece = Rook.FromColor(game.CurrentTurn)
		case 'b':
			move.PromotionPiece = Bishop.FromColor(game.CurrentTurn)
		case 'n':
			move.PromotionPiece = Knight.FromColor(game.CurrentTurn)
		default:
			return Move{}, fmt.Errorf("invalid promotion piece: %c", moveStr[4])
		}
	}

	err := game.ValidateMove(move)
	if err != nil {
		return Move{}, err
	}

	return move, nil
}
