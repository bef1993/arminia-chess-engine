package engine

import (
	"fmt"
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

// Game represents a chess game
type Game struct {
	Board           Board
	CurrentTurn     Color
	MoveHistory     []Move // Track actual Move objects
	FullMoveCounter int    // Tracks the number of the current full turn, starting with 1
	CastlingRights  CastlingRights
	EnPassantTarget int // Square index, -1 if none
	HalfMoveClock   int // Tracks the number of half turns no capture or pawn push was made, starting with 0
	ZobristHash     uint64
	PreviousState   *Game
}

// NewGame creates a new chess game
func NewGame() *Game {
	g := &Game{
		Board:           NewBoard(),
		CurrentTurn:     White,
		MoveHistory:     []Move{},
		FullMoveCounter: 1,
		CastlingRights:  AllCastling,
		EnPassantTarget: -1,
		HalfMoveClock:   0,
		PreviousState:   nil,
	}
	g.ZobristHash = g.ComputeZobristHash()
	return g
}

// NewEmptyGame creates a new chess game with an empty board
func NewEmptyGame() *Game {
	g := &Game{
		Board:           NewEmptyBoard(),
		CurrentTurn:     White,
		MoveHistory:     []Move{},
		FullMoveCounter: 1,
		CastlingRights:  NoCastling,
		EnPassantTarget: -1,
		HalfMoveClock:   0,
		PreviousState:   nil,
	}
	g.ZobristHash = g.ComputeZobristHash()
	return g
}

// SwitchTurn changes the current turn to the other player, changes the hash
func (g *Game) SwitchTurn() {
	g.ZobristHash ^= zobristBlackTurn
	g.CurrentTurn = g.CurrentTurn.Opposite()
}

// ExecuteMove executes a move on the board and updates game state
// Returns true if move was successful, false otherwise
func (g *Game) ExecuteMove(move Move) bool {

	// Copy current state and set it as PreviousGameState
	oldGame := g.Clone()
	g.PreviousState = oldGame

	piece := g.Board.GetPiece(move.From)
	if piece == NoPiece || piece.Color() != g.CurrentTurn {
		return false
	}

	hash := g.ZobristHash

	// XOR out old state (Castling, EP, Turn)
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTarget != -1 {
		hash ^= zobristEnPassant[GetFile(g.EnPassantTarget)]
	} else {
		hash ^= zobristEnPassant[8]
	}

	// XOR out moving piece from source
	hash ^= zobristPiece[piece.Color()][piece.Type()][move.From]

	// XOR in moving piece at dest (we'll handle promotion replacement later)
	hash ^= zobristPiece[piece.Color()][piece.Type()][move.To]

	// Determine move type from board state
	targetPiece := g.Board.GetPiece(move.To)
	isCapture := targetPiece != NoPiece
	isPawnMove := piece.Type() == Pawn

	// Detect en passant capture
	isEnPassant := isPawnMove &&
		move.To == g.EnPassantTarget &&
		targetPiece == NoPiece

	// Handle Capture (Remove target piece from dest)
	if isCapture {
		hash ^= zobristPiece[targetPiece.Color()][targetPiece.Type()][move.To]
	}

	// Detect castling
	isCastling := piece.Type() == King && (move.To == move.From+2 || move.To == move.From-2)

	// Handle en passant capture (remove attacked pawn)
	if isEnPassant {
		captureSq := GetSq(GetFile(move.To), GetRank(move.From))
		isCapture = true
		g.Board.SetPiece(captureSq, NoPiece)

		// Update Hash: Remove captured EP pawn
		hash ^= zobristPiece[g.CurrentTurn][Pawn][captureSq]
	}

	// Execute the move on the board
	g.Board.MovePiece(move.From, move.To)

	// Handle castling rook movement
	if isCastling {
		rank := GetRank(move.From)
		if move.To > move.From { // Kingside
			g.Board.MovePiece(GetSq(FileH, rank), GetSq(FileF, rank))

			// Update Hash: Move Rook
			rook := g.Board.GetPiece(GetSq(FileF, rank))
			oldRookSq := GetSq(FileH, rank)
			newRookSq := GetSq(FileF, rank)
			hash ^= zobristPiece[rook.Color()][rook.Type()][oldRookSq]
			hash ^= zobristPiece[rook.Color()][rook.Type()][newRookSq]
		} else { // Queenside
			g.Board.MovePiece(GetSq(FileA, rank), GetSq(FileD, rank))

			// Update Hash: Move Rook
			rook := g.Board.GetPiece(GetSq(FileD, rank))
			oldRookSq := GetSq(FileA, rank)
			newRookSq := GetSq(FileD, rank)
			hash ^= zobristPiece[rook.Color()][rook.Type()][oldRookSq]
			hash ^= zobristPiece[rook.Color()][rook.Type()][newRookSq]
		}
	}

	// Handle pawn promotion
	if move.PromotionPiece != NoPiece {
		g.Board.SetPiece(move.To, move.PromotionPiece)

		// Update Hash: Replace Pawn with Promoted Piece
		hash ^= zobristPiece[piece.Color()][piece.Type()][move.To] // Remove Pawn
		promPiece := move.PromotionPiece
		hash ^= zobristPiece[promPiece.Color()][promPiece.Type()][move.To] // Add Promoted Piece
	}

	// Update castling rights
	g.updateCastlingRights(move, piece, targetPiece)

	// Update half-move clock
	if isCapture || isPawnMove {
		g.HalfMoveClock = 0
	} else {
		g.HalfMoveClock++
	}

	// Detect double pawn move
	rankDiff := GetRank(move.To) - GetRank(move.From)
	if rankDiff < 0 {
		rankDiff = -rankDiff
	}
	isDoublePawnMove := isPawnMove && rankDiff == 2

	// Update en passant target
	if isDoublePawnMove {
		g.EnPassantTarget = (move.From + move.To) / 2
	} else {
		g.EnPassantTarget = -1
	}

	// Add to move history
	g.MoveHistory = append(g.MoveHistory, move)

	// Increment full move number
	if g.CurrentTurn == Black {
		g.FullMoveCounter++
	}

	// XOR in new state (Castling, EP)
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTarget != -1 {
		hash ^= zobristEnPassant[GetFile(g.EnPassantTarget)]
	} else {
		hash ^= zobristEnPassant[8]
	}

	g.ZobristHash = hash
	g.SwitchTurn()
	return true
}

// ExecuteNullMove executes a null move (passing the turn).
// Used for Null Move Pruning in search.
func (g *Game) ExecuteNullMove() {
	oldGame := g.Clone()
	g.PreviousState = oldGame

	// Update Hash: Remove old EP target
	if g.EnPassantTarget != -1 {
		g.ZobristHash ^= zobristEnPassant[GetFile(g.EnPassantTarget)]
	} else {
		g.ZobristHash ^= zobristEnPassant[8]
	}

	// Set new EP target to none
	g.EnPassantTarget = -1
	g.ZobristHash ^= zobristEnPassant[8]

	// Update counters
	if g.CurrentTurn == Black {
		g.FullMoveCounter++
	}
	g.HalfMoveClock++

	// Switch turn (updates hash)
	g.SwitchTurn()
}

// updateCastlingRights revokes castling rights if king or rook moves or is captured
func (g *Game) updateCastlingRights(move Move, piece Piece, targetPiece Piece) {
	fromFile := GetFile(move.From)
	fromRank := GetRank(move.From)
	toFile := GetFile(move.To)
	toRank := GetRank(move.To)

	if piece.Type() == King {
		if piece.Color() == White {
			g.CastlingRights &= ^(WhiteKingside | WhiteQueenside)
		} else {
			g.CastlingRights &= ^(BlackKingside | BlackQueenside)
		}
	} else if piece.Type() == Rook {
		if piece.Color() == White {
			if fromFile == FileA && fromRank == Rank1 {
				g.CastlingRights &= ^WhiteQueenside
			} else if fromFile == FileH && fromRank == Rank1 {
				g.CastlingRights &= ^WhiteKingside
			}
		} else {
			if fromFile == FileA && fromRank == Rank8 {
				g.CastlingRights &= ^BlackQueenside
			} else if fromFile == FileH && fromRank == Rank8 {
				g.CastlingRights &= ^BlackKingside
			}
		}
	}

	if targetPiece != NoPiece && targetPiece.Type() == Rook {
		if targetPiece.Color() == White {
			if toFile == FileA && toRank == Rank1 {
				g.CastlingRights &= ^WhiteQueenside
			} else if toFile == FileH && toRank == Rank1 {
				g.CastlingRights &= ^WhiteKingside
			}
		} else {
			if toFile == FileA && toRank == Rank8 {
				g.CastlingRights &= ^BlackQueenside
			} else if toFile == FileH && toRank == Rank8 {
				g.CastlingRights &= ^BlackKingside
			}
		}
	}
}

// UnmakeMove undoes the last move by overwriting the current state with the previous game state
func (g *Game) UnmakeMove() {
	if g.PreviousState == nil {
		return
	}
	*g = *g.PreviousState
}

// GetGameStatus returns the current status of the game
func (g *Game) GetGameStatus() GameStatus {
	// 1. Check for Checkmate and Stalemate (require move generation)
	legalMoves := g.GenerateLegalMoves()
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
	return g.GetRepetitionCount() >= 3
}

// GetRepetitionCount returns the number of times the current position has occurred.
func (g *Game) GetRepetitionCount() int {
	currentHash := g.ZobristHash
	count := 1
	for prev := g.PreviousState; prev != nil; prev = prev.PreviousState {
		if prev.ZobristHash == currentHash {
			count++
		}
	}
	return count
}

// IsCheckmate checks if the current turn player is in checkmate
func (g *Game) IsCheckmate() bool {
	if !g.Board.IsKingInCheck(g.CurrentTurn) {
		return false
	}

	// We assume IsCheckmate is called for the current turn player
	moves := g.GenerateLegalMoves()
	return len(moves) == 0
}

// IsStalemate checks if the current turn player is in stalemate
func (g *Game) IsStalemate() bool {
	if g.Board.IsKingInCheck(g.CurrentTurn) {
		return false
	}

	// We assume IsStalemate is called for the current turn player
	moves := g.GenerateLegalMoves()
	return len(moves) == 0
}

// IsDraw checks for any draw condition (stalemate, 50-move rule, insufficient material, repetition)
func (g *Game) IsDraw() bool {
	status := g.GetGameStatus()
	return status != StatusActive && status != StatusCheckmate
}

// IsInsufficientMaterial checks if there are enough pieces to force a checkmate
func (g *Game) IsInsufficientMaterial() bool {
	return g.Board.IsInsufficientMaterial()
}

// isKingInCheckAfterMove is a fast check to see if a move is legal.
// It performs a simplified make/unmake on the board to see if the king is left in check.
// TODO maybe use ExecuteMove() instead of custom code
func (g *Game) isKingInCheckAfterMove(move Move) bool {
	movedPiece := g.Board.GetPiece(move.From)
	targetPiece := g.Board.GetPiece(move.To)

	// Handle En Passant simulation (remove the captured pawn)
	isEnPassant := movedPiece.Type() == Pawn && move.To == g.EnPassantTarget && (move.To%8) != (move.From%8)
	var epCapturedPiece Piece
	var epCaptureSq int
	if isEnPassant {
		epCaptureSq = GetSq(GetFile(move.To), GetRank(move.From))
		epCapturedPiece = g.Board.GetPiece(epCaptureSq)
		g.Board.SetPiece(epCaptureSq, NoPiece)
	}

	// Make the move on the board
	g.Board.SetPiece(move.From, NoPiece)
	if move.PromotionPiece != NoPiece {
		g.Board.SetPiece(move.To, move.PromotionPiece)
	} else {
		g.Board.SetPiece(move.To, movedPiece)
	}

	kingInCheck := g.Board.IsKingInCheck(g.CurrentTurn)

	// Undo the move
	g.Board.SetPiece(move.From, movedPiece)
	g.Board.SetPiece(move.To, targetPiece) // Restore target
	if isEnPassant {
		g.Board.SetPiece(epCaptureSq, epCapturedPiece)
	}
	return kingInCheck
}

// GenerateLegalMoves returns all legal moves for the current turn, considering game state
func (g *Game) GenerateLegalMoves() []Move {
	pseudoLegalMoves := g.GenerateAllPseudoLegalMoves()

	var legalMoves []Move
	for _, move := range pseudoLegalMoves {
		if !g.isKingInCheckAfterMove(move) {
			legalMoves = append(legalMoves, move)
		}
	}
	return legalMoves
}

// GetNoisyMoves returns all legal capture and promotion moves for the current turn.
// This is optimized for Quiescence Search to avoid validating quiet moves.
func (g *Game) GetNoisyMoves() []Move {
	var noisyMoves []Move

	pseudoLegalMoves := g.GenerateAllPseudoLegalMoves()

	for _, move := range pseudoLegalMoves {
		movedPiece := g.Board.GetPiece(move.From)
		// Filter: Only keep Captures and Promotions
		targetPiece := g.Board.GetPiece(move.To)
		isCapture := targetPiece != NoPiece
		isPromotion := move.PromotionPiece != NoPiece

		// Check En Passant (special capture case)
		isEnPassant := movedPiece.Type() == Pawn && move.To == g.EnPassantTarget && (move.To%8) != (move.From%8)
		if isEnPassant {
			isCapture = true
		}

		if !isCapture && !isPromotion {
			continue
		}

		if !g.isKingInCheckAfterMove(move) {
			noisyMoves = append(noisyMoves, move)
		}
	}
	return noisyMoves
}

func (g *Game) ValidateMove(move Move) error {
	// Check if legalMove exists in legal legalMoves
	legalMoves := g.GenerateLegalMoves()

	for _, legalMove := range legalMoves {
		if legalMove.From == move.From && legalMove.To == move.To &&
			legalMove.PromotionPiece == move.PromotionPiece {
			return nil
		}
	}

	// If we have a promotion legalMove but user didn't specify promotion piece
	// Check if there are any promotion legalMoves for these coordinates
	if move.PromotionPiece == NoPiece {
		for _, legalMove := range legalMoves {
			if legalMove.From == move.From && legalMove.To == move.To &&
				legalMove.PromotionPiece != NoPiece {
				return fmt.Errorf("promotion move does not have a promotion piece specified")
			}
		}
	}

	return fmt.Errorf("illegal move")
}

// Clone creates a deep copy of the game state for parallel search
func (g *Game) Clone() *Game {
	newGame := *g
	// Deep copy MoveHistory
	newGame.MoveHistory = make([]Move, len(g.MoveHistory))
	copy(newGame.MoveHistory, g.MoveHistory)

	return &newGame
}
