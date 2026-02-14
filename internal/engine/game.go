package engine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// GameStatus represents the current state of the game
type GameStatus int

const (
	StatusActive GameStatus = iota
	StatusCheckmate
	StatusStalemate
	StatusDraw50Move
	StatusDrawRepetition
	StatusDrawInsufficientMaterial
)

// GameState captures the state of the game before a move is made, for undo purposes
type GameState struct {
	CastlingRights     CastlingRights
	EnPassantTargetCol int
	EnPassantTargetRow int
	HalfMoveClock      int
	CapturedPiece      Piece
	PositionHistory    []string
	ZobristHash        uint64
}

// Game represents a chess game
type Game struct {
	Board          *Board
	CurrentTurn    Color
	MoveHistory    []Move      // Track actual Move objects
	LastMove       *Move       // For en passant target tracking
	FullMoveNumber int         // Increments after black's move
	StateHistory   []GameState // Stack of states for undoing moves

	// GameState is embedded to hold the current state fields like castling rights,
	// en passant target, half-move clock, and position history.
	GameState
}

// NewGame creates a new chess game
func NewGame() *Game {
	g := &Game{
		Board:          NewBoard(),
		CurrentTurn:    White,
		MoveHistory:    []Move{},
		LastMove:       nil,
		FullMoveNumber: 1,
		StateHistory:   []GameState{},
		GameState: GameState{
			EnPassantTargetCol: -1,
			EnPassantTargetRow: -1,
			CastlingRights:     AllCastling,
			HalfMoveClock:      0,
			PositionHistory:    []string{},
		},
	}
	g.PositionHistory = append(g.PositionHistory, g.GeneratePositionKey())
	g.ZobristHash = g.ComputeZobristHash()
	return g
}

// PrintBoard prints the current board state to the console
func (g *Game) PrintBoard(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintln(w, "  a b c d e f g h")
	fmt.Fprintln(w, "  ╔═╦═╦═╦═╦═╦═╦═╦═╗")

	for row := 0; row < 8; row++ {
		fmt.Fprint(w, 8-row)
		fmt.Fprint(w, " ")
		fmt.Fprint(w, "║")

		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != NoPiece {
				fmt.Fprint(w, piece.GetSymbol())
			} else {
				fmt.Fprint(w, " ")
			}

			if col < 7 {
				fmt.Fprint(w, "║")
			} else {
				fmt.Fprint(w, "║")
			}
		}

		fmt.Fprint(w, " ")
		fmt.Fprint(w, 8-row)
		if row < 7 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "  ╠═╬═╬═╬═╬═╬═╬═╬═╣")
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "  ╚═╩═╩═╩═╩═╩═╩═╩═╝")
	fmt.Fprintln(w, "  a b c d e f g h")
	fmt.Fprintln(w)

	if g.CurrentTurn == White {
		fmt.Fprintln(w, "Current turn: White")
	} else {
		fmt.Fprintln(w, "Current turn: Black")
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

	// Start updating hash incrementally
	hash := g.ZobristHash

	// XOR out old state (Castling, EP, Turn)
	// We use the current state values before they are modified
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTargetCol != -1 {
		hash ^= zobristEnPassant[g.EnPassantTargetCol]
	} else {
		hash ^= zobristEnPassant[8]
	}
	hash ^= zobristBlackTurn // Toggle turn

	// XOR out moving piece from source
	srcSq := move.FromRow*8 + move.FromCol
	hash ^= zobristPiece[piece.Color()][piece.Type()][srcSq]

	// XOR in moving piece at dest (we'll handle promotion replacement later)
	// Note: If it's a capture, we'll XOR out the captured piece below
	dstSq := move.ToRow*8 + move.ToCol
	hash ^= zobristPiece[piece.Color()][piece.Type()][dstSq]

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

	// Handle Capture (Remove target piece from dest)
	if isCapture {
		hash ^= zobristPiece[targetPiece.Color()][targetPiece.Type()][dstSq]
	}

	// Detect castling (King moving 2 squares horizontally)
	isCastling := piece.Type() == King && (move.ToCol-move.FromCol == 2 || move.FromCol-move.ToCol == 2)

	// Capture state before changes
	state := g.GameState // This is a copy.
	state.CapturedPiece = targetPiece
	if isEnPassant {
		state.CapturedPiece = g.Board.GetPiece(move.ToCol, move.FromRow)
	}

	// Handle en passant capture (remove attacked pawn)
	if isEnPassant {
		isCapture = true
		epCaptureRow := move.FromRow
		g.Board.SetPiece(move.ToCol, epCaptureRow, NoPiece)

		// Update Hash: Remove captured EP pawn
		epPawn := state.CapturedPiece
		epSq := epCaptureRow*8 + move.ToCol
		hash ^= zobristPiece[epPawn.Color()][epPawn.Type()][epSq]
	}

	// Execute the move
	g.Board.MovePiece(move.FromCol, move.FromRow, move.ToCol, move.ToRow)

	// Handle castling rook movement
	if isCastling {
		row := move.FromRow
		if move.ToCol > move.FromCol { // Kingside
			// Move rook from H-file (7) to F-file (5)
			g.Board.MovePiece(FileH, row, FileF, row)

			// Update Hash: Move Rook
			rook := g.Board.GetPiece(FileF, row)
			oldRookSq := row*8 + FileH
			newRookSq := row*8 + FileF
			hash ^= zobristPiece[rook.Color()][rook.Type()][oldRookSq]
			hash ^= zobristPiece[rook.Color()][rook.Type()][newRookSq]
		} else { // Queenside
			// Move rook from A-file (0) to D-file (3)
			g.Board.MovePiece(FileA, row, FileD, row)

			// Update Hash: Move Rook
			rook := g.Board.GetPiece(FileD, row)
			oldRookSq := row*8 + FileA
			newRookSq := row*8 + FileD
			hash ^= zobristPiece[rook.Color()][rook.Type()][oldRookSq]
			hash ^= zobristPiece[rook.Color()][rook.Type()][newRookSq]
		}
	}

	// Handle pawn promotion
	if move.PromotionPiece != NoPiece {
		g.Board.SetPiece(move.ToCol, move.ToRow, move.PromotionPiece)

		// Update Hash: Replace Pawn with Promoted Piece
		// We already added the Pawn at dstSq above. Remove it.
		hash ^= zobristPiece[piece.Color()][piece.Type()][dstSq]
		// Add promoted piece
		promPiece := move.PromotionPiece
		hash ^= zobristPiece[promPiece.Color()][promPiece.Type()][dstSq]
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

	// Update position history for repetition check
	// If the move is irreversible (capture or pawn move), we clear the history
	// because previous positions can never be reached again.
	// This is an optimization and also correct according to FIDE rules for 3-fold repetition.
	newPosKey := g.GeneratePositionKey()
	if isCapture || isPawnMove {
		g.PositionHistory = []string{newPosKey}
	} else {
		g.PositionHistory = append(g.PositionHistory, newPosKey)
	}

	// Push state to history
	g.StateHistory = append(g.StateHistory, state)

	// XOR in new state (Castling, EP)
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTargetCol != -1 {
		hash ^= zobristEnPassant[g.EnPassantTargetCol]
	} else {
		hash ^= zobristEnPassant[8]
	}

	g.ZobristHash = hash

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

// UnmakeMove undoes the last move
func (g *Game) UnmakeMove() {
	if len(g.StateHistory) == 0 || len(g.MoveHistory) == 0 {
		return
	}

	// Pop state and move
	state := g.StateHistory[len(g.StateHistory)-1]
	g.StateHistory = g.StateHistory[:len(g.StateHistory)-1]

	move := g.MoveHistory[len(g.MoveHistory)-1]
	g.MoveHistory = g.MoveHistory[:len(g.MoveHistory)-1]

	// Restore game state fields
	g.GameState = state

	// Switch turn back
	g.SwitchTurn()
	if g.CurrentTurn == Black {
		g.FullMoveNumber--
	}

	// Restore LastMove
	if len(g.MoveHistory) > 0 {
		lm := g.MoveHistory[len(g.MoveHistory)-1]
		g.LastMove = &lm
	} else {
		g.LastMove = nil
	}

	// Restore pieces on board
	// 1. Move the piece back from To -> From
	// Note: The piece at ToCol, ToRow is the one that moved (or promoted piece)
	movedPiece := g.Board.GetPiece(move.ToCol, move.ToRow)
	g.Board.MovePiece(move.ToCol, move.ToRow, move.FromCol, move.FromRow)

	// 2. If it was a promotion, revert the piece at From to a Pawn
	if move.PromotionPiece != NoPiece {
		pawn := WhitePawn
		if g.CurrentTurn == Black {
			pawn = BlackPawn
		}
		g.Board.SetPiece(move.FromCol, move.FromRow, pawn)
		movedPiece = pawn // Update for En Passant check below
	}

	// 3. Restore captured piece
	if state.CapturedPiece != NoPiece {
		// Check if it was En Passant
		if movedPiece.Type() == Pawn && move.ToCol == g.EnPassantTargetCol && move.ToRow == g.EnPassantTargetRow {
			// En Passant capture: restore pawn at (ToCol, FromRow)
			g.Board.SetPiece(move.ToCol, move.FromRow, state.CapturedPiece)
		} else {
			// Normal capture: restore piece at ToCol, ToRow
			g.Board.SetPiece(move.ToCol, move.ToRow, state.CapturedPiece)
		}
	}

	// Castling undo
	// If King moved 2 squares, move the Rook back
	if movedPiece.Type() == King && (move.ToCol-move.FromCol == 2 || move.FromCol-move.ToCol == 2) {
		row := move.FromRow
		if move.ToCol > move.FromCol { // Kingside castling (e1->g1 or e8->g8)
			// Rook moved from H to F. Move back F -> H.
			g.Board.MovePiece(FileF, row, FileH, row)
		} else { // Queenside castling (e1->c1 or e8->c8)
			// Rook moved from A to D. Move back D -> A.
			g.Board.MovePiece(FileD, row, FileA, row)
		}
	}
}

// GetGameStatus returns the current status of the game
func (g *Game) GetGameStatus() GameStatus {
	// 1. Check for Checkmate and Stalemate (require move generation)
	legalMoves := g.GetLegalMoves()
	if len(legalMoves) == 0 {
		if g.Board.IsKingInCheck(g.CurrentTurn) {
			return StatusCheckmate
		}
		return StatusStalemate
	}

	// 2. Check for Insufficient Material
	if g.IsInsufficientMaterial() {
		return StatusDrawInsufficientMaterial
	}

	// 3. Check for 50-Move Rule
	if g.IsDrawByFiftyMoveRule() {
		return StatusDraw50Move
	}

	// 4. Check for Threefold Repetition
	if g.CanClaimDrawByThreefoldRepetition() {
		return StatusDrawRepetition
	}

	return StatusActive
}

// IsDrawByFiftyMoveRule checks if 50 moves have passed without capture or pawn move
func (g *Game) IsDrawByFiftyMoveRule() bool {
	return g.HalfMoveClock >= 100 // 50 full moves = 100 half-moves
}

// CanClaimDrawByThreefoldRepetition checks if current position has occurred 3 times
func (g *Game) CanClaimDrawByThreefoldRepetition() bool {
	currentKey := g.GeneratePositionKey()
	count := 0
	for _, key := range g.PositionHistory {
		if key == currentKey {
			count++
		}
	}
	return count >= 3
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

// IsDraw checks for any draw condition (stalemate, 50-move rule, insufficient material, repetition)
func (g *Game) IsDraw() bool {
	status := g.GetGameStatus()
	return status != StatusActive && status != StatusCheckmate
}

// IsInsufficientMaterial checks if there are enough pieces to force a checkmate
func (g *Game) IsInsufficientMaterial() bool {
	// Count pieces
	whitePieces := 0
	blackPieces := 0
	whiteBishops := 0
	blackBishops := 0
	whiteKnights := 0
	blackKnights := 0

	// Also need to track bishop square colors for same-colored bishop ending
	whiteBishopSquareColor := -1 // 0 for light, 1 for dark
	blackBishopSquareColor := -1

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece == NoPiece {
				continue
			}

			// If there's a pawn, rook, or queen, it's not insufficient material
			if piece.Type() == Pawn || piece.Type() == Rook || piece.Type() == Queen {
				return false
			}

			if piece.Color() == White {
				whitePieces++
				if piece.Type() == Bishop {
					whiteBishops++
					whiteBishopSquareColor = (row + col) % 2
				} else if piece.Type() == Knight {
					whiteKnights++
				}
			} else {
				blackPieces++
				if piece.Type() == Bishop {
					blackBishops++
					blackBishopSquareColor = (row + col) % 2
				} else if piece.Type() == Knight {
					blackKnights++
				}
			}
		}
	}

	// King vs King
	if whitePieces == 1 && blackPieces == 1 {
		return true
	}

	// King + Knight vs King
	if (whitePieces == 2 && whiteKnights == 1 && blackPieces == 1) ||
		(blackPieces == 2 && blackKnights == 1 && whitePieces == 1) {
		return true
	}

	// King + Bishop vs King
	if (whitePieces == 2 && whiteBishops == 1 && blackPieces == 1) ||
		(blackPieces == 2 && blackBishops == 1 && whitePieces == 1) {
		return true
	}

	// King + Bishop vs King + Bishop (same color squares)
	if whitePieces == 2 && whiteBishops == 1 && blackPieces == 2 && blackBishops == 1 {
		return whiteBishopSquareColor == blackBishopSquareColor
	}

	return false
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

// GetNoisyMoves returns all legal capture and promotion moves for the current turn.
// This is optimized for Quiescence Search to avoid validating quiet moves.
func (g *Game) GetNoisyMoves() []Move {
	var noisyMoves []Move
	// Use the full generator with game state (Castling, En Passant)
	mg := NewMoveGeneratorFull(g.Board, g.EnPassantTargetCol, g.EnPassantTargetRow, g.CastlingRights)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != NoPiece && piece.Color() == g.CurrentTurn {
				moves := mg.GenerateMovesForPiece(col, row)

				for _, move := range moves {
					// Filter: Only keep Captures and Promotions
					targetPiece := g.Board.GetPiece(move.ToCol, move.ToRow)
					isCapture := targetPiece != NoPiece
					isPromotion := move.PromotionPiece != NoPiece

					// Check En Passant (special capture case)
					isEnPassant := piece.Type() == Pawn && move.ToCol == g.EnPassantTargetCol && move.ToRow == g.EnPassantTargetRow && move.ToCol != move.FromCol
					if isEnPassant {
						isCapture = true
					}

					if !isCapture && !isPromotion {
						continue
					}

					// Simulate the move to check legality (King safety)
					var epCapturedPiece Piece
					if isEnPassant {
						epCapturedPiece = g.Board.GetPiece(move.ToCol, move.FromRow)
						g.Board.SetPiece(move.ToCol, move.FromRow, NoPiece)
					}

					g.Board.MovePiece(move.FromCol, move.FromRow, move.ToCol, move.ToRow)

					if !g.Board.IsKingInCheck(g.CurrentTurn) {
						noisyMoves = append(noisyMoves, move)
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
	return noisyMoves
}

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
		row := r // FEN starts from rank 8 (row 0) to rank 1 (row 7)
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
				g.Board.SetPiece(col, row, piece)
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
	g.EnPassantTargetCol = -1
	g.EnPassantTargetRow = -1
	if parts[3] != "-" {
		col, row := Sq(parts[3])
		if col != -1 && row != -1 {
			g.EnPassantTargetCol = col
			g.EnPassantTargetRow = row
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
	g.PositionHistory = []string{g.GeneratePositionKey()}

	// Compute initial hash
	g.ZobristHash = g.ComputeZobristHash()

	return nil
}

// GeneratePositionKey returns a string representing the unique position (pieces, turn, castling, ep)
func (g *Game) GeneratePositionKey() string {
	var sb strings.Builder

	// 1. Piece placement
	for row := 0; row < 8; row++ {
		emptyCount := 0
		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
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
		if row < 7 {
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
	if g.EnPassantTargetCol != -1 && g.EnPassantTargetRow != -1 {
		col := rune('a' + g.EnPassantTargetCol)
		row := 8 - g.EnPassantTargetRow
		sb.WriteString(fmt.Sprintf("%c%d", col, row))
	} else {
		sb.WriteString("-")
	}

	return sb.String()
}
