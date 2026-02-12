package engine

import "fmt"

// Game represents a chess game
type Game struct {
	Board              *Board
	CurrentTurn        Color
	MoveHistory        []Move         // Track actual Move objects
	LastMove           *Move          // For en passant target tracking
	EnPassantTargetCol int            // Column of en passant target (-1 if none)
	EnPassantTargetRow int            // Row of en passant target (-1 if none)
	CastlingRights     CastlingRights // Bitmask tracking which sides can castle
	HalfMoveClock      int            // For 50-move rule (reset on capture or pawn move)
	FullMoveNumber     int            // Increments after black's move
}

// NewGame creates a new chess game
func NewGame() *Game {
	return &Game{
		Board:              NewBoard(),
		CurrentTurn:        White,
		MoveHistory:        []Move{},
		LastMove:           nil,
		EnPassantTargetCol: -1,
		EnPassantTargetRow: -1,
		CastlingRights:     AllCastling,
		HalfMoveClock:      0,
		FullMoveNumber:     1,
	}
}

// PrintBoard prints the current board state to the console
func (g *Game) PrintBoard() {
	fmt.Println("  a b c d e f g h")
	fmt.Println("  ╔═╦═╦═╦═╦═╦═╦═╦═╗")

	for row := 0; row < 8; row++ {
		fmt.Print(8 - row)
		fmt.Print("║")

		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != NoPiece {
				fmt.Print(piece.GetSymbol())
			} else {
				fmt.Print(" ")
			}

			if col < 7 {
				fmt.Print("║")
			} else {
				fmt.Print("║")
			}
		}

		fmt.Print(8 - row)
		if row < 7 {
			fmt.Println()
			fmt.Println("  ╠═╬═╬═╬═╬═╬═╬═╬═╣")
		}
	}

	fmt.Println()
	fmt.Println("  ╚═╩═╩═╩═╩═╩═╩═╩═╝")
	fmt.Println("  a b c d e f g h")
	fmt.Println()

	if g.CurrentTurn == White {
		fmt.Println("Current turn: White")
	} else {
		fmt.Println("Current turn: Black")
	}
}

// SwitchTurn changes the current turn to the other player
func (g *Game) SwitchTurn() {
	if g.CurrentTurn == White {
		g.CurrentTurn = Black
	} else {
		g.CurrentTurn = White
	}
}

// ExecuteMove executes a move on the board and updates game state
// Returns true if move was successful, false otherwise
func (g *Game) ExecuteMove(move Move) bool {
	piece := g.Board.GetPiece(move.FromCol, move.FromRow)
	if piece == NoPiece || piece.Color() != g.CurrentTurn {
		return false
	}

	// Determine move type from board state
	targetPiece := g.Board.GetPiece(move.ToCol, move.ToRow)
	isCapture := targetPiece != NoPiece
	isPawnMove := piece.Type() == Pawn

	// Detect en passant capture (pawn moving diagonally to empty square at en passant target)
	isEnPassant := isPawnMove &&
		move.FromCol != move.ToCol &&
		targetPiece == NoPiece &&
		move.ToCol == g.EnPassantTargetCol &&
		move.ToRow == g.EnPassantTargetRow

	// Detect castling (King moving 2 squares horizontally)
	isCastling := piece.Type() == King && (move.ToCol-move.FromCol == 2 || move.FromCol-move.ToCol == 2)

	// Handle en passant capture (remove attacked pawn)
	if isEnPassant {
		isCapture = true
		epCaptureRow := move.FromRow
		g.Board.SetPiece(move.ToCol, epCaptureRow, NoPiece)
	}

	// Execute the move
	g.Board.MovePiece(move.FromCol, move.FromRow, move.ToCol, move.ToRow)

	// Handle castling rook movement
	if isCastling {
		row := move.FromRow
		if move.ToCol > move.FromCol { // Kingside
			// Move rook from H-file (7) to F-file (5)
			g.Board.MovePiece(FileH, row, FileF, row)
		} else { // Queenside
			// Move rook from A-file (0) to D-file (3)
			g.Board.MovePiece(FileA, row, FileD, row)
		}
	}

	// Handle pawn promotion
	if move.PromotionPiece != NoPiece {
		g.Board.SetPiece(move.ToCol, move.ToRow, move.PromotionPiece)
	}

	// Update castling rights if king or rook moved
	g.updateCastlingRights(move, piece, targetPiece)

	// Update half-move clock (reset on captures or pawn moves, increment otherwise)
	if isCapture || isPawnMove {
		g.HalfMoveClock = 0
	} else {
		g.HalfMoveClock++
	}

	// Detect double pawn move (creates en passant target)
	rowDiff := move.ToRow - move.FromRow
	if rowDiff < 0 {
		rowDiff = -rowDiff
	}
	isDoublePawnMove := isPawnMove && rowDiff == 2

	// Update en passant target
	if isDoublePawnMove {
		g.EnPassantTargetCol = move.ToCol
		g.EnPassantTargetRow = move.FromRow + (move.ToRow-move.FromRow)/2
	} else {
		g.EnPassantTargetCol = -1
		g.EnPassantTargetRow = -1
	}

	// Add to move history
	g.MoveHistory = append(g.MoveHistory, move)
	g.LastMove = &move

	// Increment full move number after black moves
	if g.CurrentTurn == Black {
		g.FullMoveNumber++
	}

	// Switch turns
	g.SwitchTurn()

	return true
}

// updateCastlingRights revokes castling rights if king or rook moves or is captured
func (g *Game) updateCastlingRights(move Move, piece Piece, targetPiece Piece) {
	if piece.Type() == King {
		// King moved - revoke all castling rights for this color
		if piece.Color() == White {
			g.CastlingRights &= ^(WhiteKingside | WhiteQueenside)
		} else {
			g.CastlingRights &= ^(BlackKingside | BlackQueenside)
		}
	} else if piece.Type() == Rook {
		// Rook moved - revoke castling on that side
		if piece.Color() == White {
			if move.FromCol == FileA && move.FromRow == Rank1 {
				g.CastlingRights &= ^WhiteQueenside
			} else if move.FromCol == FileH && move.FromRow == Rank1 {
				g.CastlingRights &= ^WhiteKingside
			}
		} else {
			if move.FromCol == FileA && move.FromRow == Rank8 {
				g.CastlingRights &= ^BlackQueenside
			} else if move.FromCol == FileH && move.FromRow == Rank8 {
				g.CastlingRights &= ^BlackKingside
			}
		}
	}

	// If a rook is captured, revoke castling rights
	if targetPiece != NoPiece && targetPiece.Type() == Rook {
		if targetPiece.Color() == White {
			if move.ToCol == FileA && move.ToRow == Rank1 {
				g.CastlingRights &= ^WhiteQueenside
			} else if move.ToCol == FileH && move.ToRow == Rank1 {
				g.CastlingRights &= ^WhiteKingside
			}
		} else {
			if move.ToCol == FileA && move.ToRow == Rank8 {
				g.CastlingRights &= ^BlackQueenside
			} else if move.ToCol == FileH && move.ToRow == Rank8 {
				g.CastlingRights &= ^BlackKingside
			}
		}
	}
}

// IsDrawByFiftyMoveRule checks if 50 moves have passed without capture or pawn move
func (g *Game) IsDrawByFiftyMoveRule() bool {
	return g.HalfMoveClock >= 100 // 50 full moves = 100 half-moves
}

// CanClaimDrawByThreefoldRepetition checks if current position has occurred 3 times
func (g *Game) CanClaimDrawByThreefoldRepetition() bool {
	// TODO: Implement by comparing board positions in move history
	return false
}

// IsCheckmate checks if the current turn player is in checkmate
func (g *Game) IsCheckmate() bool {
	if !g.Board.IsKingInCheck(g.CurrentTurn) {
		return false
	}

	// We assume IsCheckmate is called for the current turn player
	moves := g.GetLegalMoves()
	return len(moves) == 0
}

// IsStalemate checks if the current turn player is in stalemate
func (g *Game) IsStalemate() bool {
	if g.Board.IsKingInCheck(g.CurrentTurn) {
		return false
	}

	// We assume IsStalemate is called for the current turn player
	moves := g.GetLegalMoves()
	return len(moves) == 0
}

// GetLegalMoves returns all legal moves for the current turn, considering game state
func (g *Game) GetLegalMoves() []Move {
	var legalMoves []Move
	// Use the full generator with game state (Castling, En Passant)
	mg := NewMoveGeneratorFull(g.Board, g.EnPassantTargetCol, g.EnPassantTargetRow, g.CastlingRights)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != NoPiece && piece.Color() == g.CurrentTurn {
				moves := mg.GenerateMovesForPiece(col, row)

				// Filter out moves that leave the king in check
				for _, move := range moves {
					// Simulate the move to check legality

					targetPiece := g.Board.GetPiece(move.ToCol, move.ToRow)

					// Handle En Passant simulation (remove the captured pawn)
					isEnPassant := piece.Type() == Pawn && move.ToCol == g.EnPassantTargetCol && move.ToRow == g.EnPassantTargetRow && move.ToCol != move.FromCol
					var epCapturedPiece Piece
					if isEnPassant {
						epCapturedPiece = g.Board.GetPiece(move.ToCol, move.FromRow)
						g.Board.SetPiece(move.ToCol, move.FromRow, NoPiece)
					}

					g.Board.MovePiece(move.FromCol, move.FromRow, move.ToCol, move.ToRow)

					if !g.Board.IsKingInCheck(g.CurrentTurn) {
						legalMoves = append(legalMoves, move)
					}

					// Undo the move
					g.Board.SetPiece(move.FromCol, move.FromRow, piece)
					g.Board.SetPiece(move.ToCol, move.ToRow, targetPiece)

					if isEnPassant {
						g.Board.SetPiece(move.ToCol, move.FromRow, epCapturedPiece)
					}
				}
			}
		}
	}
	return legalMoves
}
