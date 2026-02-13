package engine

import (
	"fmt"
)

// Move represents a chess move (pure data)
type Move struct {
	FromCol        int
	FromRow        int
	ToCol          int
	ToRow          int
	PromotionPiece Piece // 0 if no promotion
}

// NewMove creates a move without promotion
func NewMove(fromCol, fromRow, toCol, toRow int) Move {
	return Move{
		FromCol:        fromCol,
		FromRow:        fromRow,
		ToCol:          toCol,
		ToRow:          toRow,
		PromotionPiece: 0, // No promotion (Pawn constant)
	}
}

// NewPromotionMove creates a pawn promotion move
func NewPromotionMove(fromCol, fromRow, toCol, toRow int, promotionPiece Piece) Move {
	return Move{
		FromCol:        fromCol,
		FromRow:        fromRow,
		ToCol:          toCol,
		ToRow:          toRow,
		PromotionPiece: promotionPiece,
	}
}

// String returns the UCI representation of the move.
func (m Move) String() string {
	moveStr := fmt.Sprintf("%c%d%c%d",
		rune('a'+m.FromCol),
		8-m.FromRow,
		rune('a'+m.ToCol),
		8-m.ToRow)

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

	// Parse move: format is e2e4 or e7e8q (with promotion)
	fromCol := int(moveStr[0] - 'a')
	fromRow := 8 - int(moveStr[1]-'0')
	toCol := int(moveStr[2] - 'a')
	toRow := 8 - int(moveStr[3]-'0')

	if fromCol < 0 || fromCol > 7 || fromRow < 0 || fromRow > 7 ||
		toCol < 0 || toCol > 7 || toRow < 0 || toRow > 7 {
		return Move{}, fmt.Errorf("square out of bounds")
	}

	var promotionPiece Piece
	if len(moveStr) == 5 {
		switch moveStr[4] {
		case 'q':
			promotionPiece = Queen.FromColor(game.CurrentTurn)
		case 'r':
			promotionPiece = Rook.FromColor(game.CurrentTurn)
		case 'b':
			promotionPiece = Bishop.FromColor(game.CurrentTurn)
		case 'n':
			promotionPiece = Knight.FromColor(game.CurrentTurn)
		default:
			return Move{}, fmt.Errorf("invalid promotion piece: %c", moveStr[4])
		}
	}

	// Check if move exists in legal moves
	moves := game.GetLegalMoves()
	for _, move := range moves {
		if move.FromCol == fromCol && move.FromRow == fromRow &&
			move.ToCol == toCol && move.ToRow == toRow &&
			move.PromotionPiece == promotionPiece {
			return move, nil
		}
	}

	// If we have a promotion move but user didn't specify promotion piece
	// Check if there are any promotion moves for these coordinates
	if promotionPiece == NoPiece {
		for _, move := range moves {
			if move.FromCol == fromCol && move.FromRow == fromRow &&
				move.ToCol == toCol && move.ToRow == toRow &&
				move.PromotionPiece != NoPiece {
				return Move{}, fmt.Errorf("promotion piece required (e.g., %sq)", moveStr)
			}
		}
	}

	return Move{}, fmt.Errorf("illegal move: %s", moveStr)
}

// MoveGenerator generates legal moves for the current position
type MoveGenerator struct {
	Board              *Board
	EnPassantTargetCol int // Column of en passant target (-1 if none)
	EnPassantTargetRow int // Row of en passant target (-1 if none)
	CastlingRights     CastlingRights
}

// NewMoveGenerator creates a new move generator
func NewMoveGenerator(board *Board) *MoveGenerator {
	return &MoveGenerator{
		Board:              board,
		EnPassantTargetCol: -1,
		EnPassantTargetRow: -1,
		CastlingRights:     NoCastling,
	}
}

// NewMoveGeneratorWithEnPassant creates a move generator with en passant target
func NewMoveGeneratorWithEnPassant(board *Board, epCol, epRow int) *MoveGenerator {
	return &MoveGenerator{
		Board:              board,
		EnPassantTargetCol: epCol,
		EnPassantTargetRow: epRow,
		CastlingRights:     NoCastling,
	}
}

// NewMoveGeneratorFull creates a move generator with all state
func NewMoveGeneratorFull(board *Board, epCol, epRow int, castlingRights CastlingRights) *MoveGenerator {
	return &MoveGenerator{
		Board:              board,
		EnPassantTargetCol: epCol,
		EnPassantTargetRow: epRow,
		CastlingRights:     castlingRights,
	}
}

// GenerateMovesForPiece generates all possible moves for a piece at the given position
func (mg *MoveGenerator) GenerateMovesForPiece(col, row int) []Move {
	piece := mg.Board.GetPiece(col, row)
	if piece == NoPiece {
		return []Move{}
	}

	var moves []Move

	switch piece.Type() {
	case Pawn:
		moves = mg.generatePawnMoves(col, row, piece.Color())
	case Knight:
		moves = mg.generateKnightMoves(col, row, piece.Color())
	case Bishop:
		moves = mg.generateBishopMoves(col, row, piece.Color())
	case Rook:
		moves = mg.generateRookMoves(col, row, piece.Color())
	case Queen:
		moves = mg.generateQueenMoves(col, row, piece.Color())
	case King:
		moves = mg.generateKingMoves(col, row, piece.Color())
	case NoType:
	}

	return moves
}

func (mg *MoveGenerator) generatePawnMoves(col, row int, color Color) []Move {
	var moves []Move
	direction := -1
	startRow := int(Rank2)
	promotionRank := int(Rank8)

	if color == Black {
		direction = 1
		startRow = int(Rank7)
		promotionRank = int(Rank1)
	}

	// Helper function to add promotion moves
	addPromotionMoves := func(toCol, toRow int) {
		// Pawn can promote to Queen, Rook, Bishop, or Knight
		for _, piece := range []PieceType{Queen, Rook, Bishop, Knight} {
			moves = append(moves, NewPromotionMove(col, row, toCol, toRow, piece.FromColor(color)))
		}
	}

	// Move forward one square
	newRow := row + direction
	if newRow >= 0 && newRow < 8 && mg.Board.IsEmpty(col, newRow) {
		if newRow == promotionRank {
			addPromotionMoves(col, newRow)
		} else {
			moves = append(moves, NewMove(col, row, col, newRow))
		}

		// Move forward two squares from starting position
		if row == startRow {
			newRow2 := row + 2*direction
			if mg.Board.IsEmpty(col, newRow2) {
				moves = append(moves, NewMove(col, row, col, newRow2))
			}
		}
	}

	// Capture diagonally
	for dcol := -1; dcol <= 1; dcol += 2 {
		newCol := col + dcol
		newRow := row + direction
		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			// Regular capture
			if !mg.Board.IsEmpty(newCol, newRow) && !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				if newRow == promotionRank {
					addPromotionMoves(newCol, newRow)
				} else {
					moves = append(moves, NewMove(col, row, newCol, newRow))
				}
			}

			// En passant capture (never a promotion)
			if newRow == mg.EnPassantTargetRow && newCol == mg.EnPassantTargetCol {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) addCastlingMoves(col, row int, color Color, moves *[]Move) {
	if mg.CastlingRights == NoCastling {
		return
	}

	// Determine opponent color for attack checks
	opponent := Black
	if color == Black {
		opponent = White
	}

	// Cannot castle if King is currently in check
	if mg.Board.IsSquareAttackedByColor(col, row, opponent) {
		return
	}

	// White Castling
	if color == White && col == FileE && row == Rank1 {
		// Kingside (e1 -> g1)
		if mg.CastlingRights&WhiteKingside != 0 {
			// Check path is empty (f1, g1)
			if mg.Board.IsEmpty(FileF, Rank1) && mg.Board.IsEmpty(FileG, Rank1) {
				// Check path is not attacked (f1, g1)
				// Note: We already checked e1 (current pos) above
				if !mg.Board.IsSquareAttackedByColor(FileF, Rank1, opponent) &&
					!mg.Board.IsSquareAttackedByColor(FileG, Rank1, opponent) {
					*moves = append(*moves, NewMove(FileE, Rank1, FileG, Rank1))
				}
			}
		}
		// Queenside (e1 -> c1)
		if mg.CastlingRights&WhiteQueenside != 0 {
			// Check path is empty (d1, c1, b1)
			if mg.Board.IsEmpty(FileD, Rank1) && mg.Board.IsEmpty(FileC, Rank1) && mg.Board.IsEmpty(FileB, Rank1) {
				// Check path is not attacked (d1, c1)
				// Note: b1 does not need to be safe, only empty
				if !mg.Board.IsSquareAttackedByColor(FileD, Rank1, opponent) &&
					!mg.Board.IsSquareAttackedByColor(FileC, Rank1, opponent) {
					*moves = append(*moves, NewMove(FileE, Rank1, FileC, Rank1))
				}
			}
		}
	}

	// Black Castling
	if color == Black && col == FileE && row == Rank8 {
		// Kingside (e8 -> g8)
		if mg.CastlingRights&BlackKingside != 0 {
			if mg.Board.IsEmpty(FileF, Rank8) && mg.Board.IsEmpty(FileG, Rank8) {
				if !mg.Board.IsSquareAttackedByColor(FileF, Rank8, opponent) &&
					!mg.Board.IsSquareAttackedByColor(FileG, Rank8, opponent) {
					*moves = append(*moves, NewMove(FileE, Rank8, FileG, Rank8))
				}
			}
		}
		// Queenside (e8 -> c8)
		if mg.CastlingRights&BlackQueenside != 0 {
			if mg.Board.IsEmpty(FileD, Rank8) && mg.Board.IsEmpty(FileC, Rank8) && mg.Board.IsEmpty(FileB, Rank8) {
				if !mg.Board.IsSquareAttackedByColor(FileD, Rank8, opponent) &&
					!mg.Board.IsSquareAttackedByColor(FileC, Rank8, opponent) {
					*moves = append(*moves, NewMove(FileE, Rank8, FileC, Rank8))
				}
			}
		}
	}
}

func (mg *MoveGenerator) generateKnightMoves(col, row int, color Color) []Move {
	var moves []Move
	knightMoves := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}

	for _, move := range knightMoves {
		newRow := row + move[0]
		newCol := col + move[1]

		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			if mg.Board.IsEmpty(newCol, newRow) || !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateBishopMoves(col, row int, color Color) []Move {
	var moves []Move
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRow := row + i*dir[0]
			newCol := col + i*dir[1]

			if newRow < 0 || newRow >= 8 || newCol < 0 || newCol >= 8 {
				break
			}

			if mg.Board.IsEmpty(newCol, newRow) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			} else if !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateRookMoves(col, row int, color Color) []Move {
	var moves []Move
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRow := row + i*dir[0]
			newCol := col + i*dir[1]

			if newRow < 0 || newRow >= 8 || newCol < 0 || newCol >= 8 {
				break
			}

			if mg.Board.IsEmpty(newCol, newRow) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			} else if !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateQueenMoves(col, row int, color Color) []Move {
	var moves []Move
	directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRow := row + i*dir[0]
			newCol := col + i*dir[1]

			if newRow < 0 || newRow >= 8 || newCol < 0 || newCol >= 8 {
				break
			}

			if mg.Board.IsEmpty(newCol, newRow) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			} else if !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generateKingMoves(col, row int, color Color) []Move {
	var moves []Move
	directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for _, dir := range directions {
		newRow := row + dir[0]
		newCol := col + dir[1]

		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			if mg.Board.IsEmpty(newCol, newRow) || !mg.Board.IsOccupiedByColor(newCol, newRow, color) {
				moves = append(moves, NewMove(col, row, newCol, newRow))
			}
		}
	}

	// Add castling moves
	mg.addCastlingMoves(col, row, color, &moves)

	return moves
}
