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
func ParseMove(moveStr string, game *Game) (Move, error) {
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

// MoveGenerator generates legal moves for the current position
type MoveGenerator struct {
	Board           *Board
	EnPassantTarget int // Square of en passant target (-1 if none)
	CastlingRights  CastlingRights
}

// NewMoveGenerator creates a new move generator
func NewMoveGenerator(board *Board) *MoveGenerator {
	return &MoveGenerator{
		Board:           board,
		EnPassantTarget: -1,
		CastlingRights:  NoCastling,
	}
}

// NewMoveGeneratorWithEnPassant creates a move generator with en passant target
func NewMoveGeneratorWithEnPassant(board *Board, epSq int) *MoveGenerator {
	return &MoveGenerator{
		Board:           board,
		EnPassantTarget: epSq,
		CastlingRights:  NoCastling,
	}
}

// NewMoveGeneratorFull creates a move generator with all state
func NewMoveGeneratorFull(board *Board, epSq int, castlingRights CastlingRights) *MoveGenerator {
	return &MoveGenerator{
		Board:           board,
		EnPassantTarget: epSq,
		CastlingRights:  castlingRights,
	}
}

// GenerateMovesForPiece generates all possible moves for a piece at the given position
func (mg *MoveGenerator) GenerateMovesForPiece(sq int) []Move {
	piece := mg.Board.GetPiece(sq)
	if piece == NoPiece {
		return []Move{}
	}

	var moves []Move

	switch piece.Type() {
	case Pawn:
		moves = mg.generatePawnMoves(sq, piece.Color())
	case Knight:
		moves = mg.generateKnightMoves(sq, piece.Color())
	case Bishop:
		moves = mg.generateBishopMoves(sq, piece.Color())
	case Rook:
		moves = mg.generateRookMoves(sq, piece.Color())
	case Queen:
		moves = mg.generateQueenMoves(sq, piece.Color())
	case King:
		moves = mg.generateKingMoves(sq, piece.Color())
	case NoType:
	}

	return moves
}

func (mg *MoveGenerator) generatePawnMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)

	direction := 1
	startRank := int(Rank2)
	promotionRank := int(Rank8)

	if color == Black {
		direction = -1
		startRank = int(Rank7)
		promotionRank = int(Rank1)
	}

	// Helper function to add promotion moves
	addPromotionMoves := func(toSq int) {
		// Pawn can promote to Queen, Rook, Bishop, or Knight
		for _, piece := range []PieceType{Queen, Rook, Bishop, Knight} {
			moves = append(moves, NewPromotionMove(sq, toSq, piece.FromColor(color)))
		}
	}

	// Move forward one square
	newRank := rank + direction
	if IsOnBoard2D(file, newRank) {
		toSq := GetSq(file, newRank)
		if mg.Board.IsEmpty(toSq) {
			if newRank == promotionRank {
				addPromotionMoves(toSq)
			} else {
				moves = append(moves, NewMove(sq, toSq))
			}

			// Move forward two squares from starting position
			if rank == startRank {
				newRank2 := rank + 2*direction
				toSq2 := GetSq(file, newRank2)
				if mg.Board.IsEmpty(toSq2) {
					moves = append(moves, NewMove(sq, toSq2))
				}
			}
		}
	}

	// Capture diagonally
	for dfile := -1; dfile <= 1; dfile += 2 {
		newFile := file + dfile
		newRank := rank + direction
		if IsOnBoard2D(newFile, newRank) {
			toSq := GetSq(newFile, newRank)
			// Regular capture
			if !mg.Board.IsEmpty(toSq) && !mg.Board.IsOccupiedByColor(toSq, color) {
				if newRank == promotionRank {
					addPromotionMoves(toSq)
				} else {
					moves = append(moves, NewMove(sq, toSq))
				}
			}

			// En passant capture (never a promotion)
			if toSq == mg.EnPassantTarget {
				moves = append(moves, NewMove(sq, toSq))
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) addCastlingMoves(sq int, color Color, moves *[]Move) {
	if mg.CastlingRights == NoCastling {
		return
	}

	// Determine opponent color for attack checks
	opponent := Black
	if color == Black {
		opponent = White
	}

	// Cannot castle if King is currently in check
	if mg.Board.IsSquareAttackedByColor(sq, opponent) {
		return
	}

	// White Castling
	if color == White && sq == E1 {
		// Kingside (e1 -> g1)
		if mg.CastlingRights&WhiteKingside != 0 {
			// Check path is empty (f1, g1)
			if mg.Board.IsEmpty(F1) && mg.Board.IsEmpty(G1) {
				// Check path is not attacked (f1, g1)
				// Note: We already checked e1 (current pos) above
				if !mg.Board.IsSquareAttackedByColor(F1, opponent) &&
					!mg.Board.IsSquareAttackedByColor(G1, opponent) {
					*moves = append(*moves, NewMove(E1, G1))
				}
			}
		}
		// Queenside (e1 -> c1)
		if mg.CastlingRights&WhiteQueenside != 0 {
			// Check path is empty (d1, c1, b1)
			if mg.Board.IsEmpty(D1) && mg.Board.IsEmpty(C1) && mg.Board.IsEmpty(B1) {
				// Check path is not attacked (d1, c1)
				// Note: b1 does not need to be safe, only empty
				if !mg.Board.IsSquareAttackedByColor(D1, opponent) &&
					!mg.Board.IsSquareAttackedByColor(C1, opponent) {
					*moves = append(*moves, NewMove(E1, C1))
				}
			}
		}
	}

	// Black Castling
	if color == Black && sq == E8 {
		// Kingside (e8 -> g8)
		if mg.CastlingRights&BlackKingside != 0 {
			if mg.Board.IsEmpty(F8) && mg.Board.IsEmpty(G8) {
				if !mg.Board.IsSquareAttackedByColor(F8, opponent) &&
					!mg.Board.IsSquareAttackedByColor(G8, opponent) {
					*moves = append(*moves, NewMove(E8, G8))
				}
			}
		}
		// Queenside (e8 -> c8)
		if mg.CastlingRights&BlackQueenside != 0 {
			if mg.Board.IsEmpty(D8) && mg.Board.IsEmpty(C8) && mg.Board.IsEmpty(B8) {
				if !mg.Board.IsSquareAttackedByColor(D8, opponent) &&
					!mg.Board.IsSquareAttackedByColor(C8, opponent) {
					*moves = append(*moves, NewMove(E8, C8))
				}
			}
		}
	}
}

func (mg *MoveGenerator) generateKnightMoves(sq int, color Color) []Move {
	var moves []Move

	// Use pre-calculated attacks
	attacks := KnightAttacks[sq]

	// Mask out own pieces
	validMoves := attacks & ^mg.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	return moves
}

func (mg *MoveGenerator) generateBishopMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if mg.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !mg.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateRookMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if mg.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !mg.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateQueenMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if mg.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !mg.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateKingMoves(sq int, color Color) []Move {
	var moves []Move

	// Use pre-calculated attacks
	attacks := KingAttacks[sq]

	// Mask out own pieces
	validMoves := attacks & ^mg.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	// Add castling moves
	mg.addCastlingMoves(sq, color, &moves)

	return moves
}
