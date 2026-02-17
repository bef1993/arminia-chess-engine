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

// GameState captures the state of the game before a move is made, for undo purposes
type GameState struct {
	CastlingRights  CastlingRights
	EnPassantTarget int // Square index, -1 if none
	HalfMoveClock   int
	CapturedPiece   Piece
	PositionHistory []uint64
	ZobristHash     uint64
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
			EnPassantTarget: -1,
			CastlingRights:  AllCastling,
			HalfMoveClock:   0,
			PositionHistory: []uint64{},
		},
	}
	g.ZobristHash = g.ComputeZobristHash()
	g.PositionHistory = append(g.PositionHistory, g.ZobristHash)
	return g
}

// NewEmptyGame creates a new chess game with an empty board
func NewEmptyGame() *Game {
	g := &Game{
		Board:          NewEmptyBoard(),
		CurrentTurn:    White,
		MoveHistory:    []Move{},
		LastMove:       nil,
		FullMoveNumber: 1,
		StateHistory:   []GameState{},
		GameState: GameState{
			EnPassantTarget: -1,
			CastlingRights:  NoCastling,
			HalfMoveClock:   0,
			PositionHistory: []uint64{},
		},
	}
	g.ZobristHash = g.ComputeZobristHash()
	g.PositionHistory = append(g.PositionHistory, g.ZobristHash)
	return g
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
	piece := g.Board.GetPiece(move.From)
	if piece == NoPiece || piece.Color() != g.CurrentTurn {
		return false
	}

	// Start updating hash incrementally
	hash := g.ZobristHash

	// XOR out old state (Castling, EP, Turn)
	// We use the current state values before they are modified
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTarget != -1 {
		hash ^= zobristEnPassant[GetFile(g.EnPassantTarget)] // Use column for hash
	} else {
		hash ^= zobristEnPassant[8]
	}
	hash ^= zobristBlackTurn // Toggle turn

	// XOR out moving piece from source
	hash ^= zobristPiece[piece.Color()][piece.Type()][move.From]

	// XOR in moving piece at dest (we'll handle promotion replacement later)
	// Note: If it's a capture, we'll XOR out the captured piece below
	hash ^= zobristPiece[piece.Color()][piece.Type()][move.To]

	// Determine move type from board state
	targetPiece := g.Board.GetPiece(move.To)
	isCapture := targetPiece != NoPiece
	isPawnMove := piece.Type() == Pawn

	// Detect en passant capture (pawn moving diagonally to empty square at en passant target)
	isEnPassant := isPawnMove &&
		move.To == g.EnPassantTarget &&
		targetPiece == NoPiece

	// Handle Capture (Remove target piece from dest)
	if isCapture {
		hash ^= zobristPiece[targetPiece.Color()][targetPiece.Type()][move.To]
	}

	// Detect castling (King moving 2 squares horizontally)
	isCastling := piece.Type() == King && (move.To == move.From+2 || move.To == move.From-2)

	// Capture state before changes
	// We must perform a deep copy of the PositionHistory slice.
	// Otherwise, when we append to g.PositionHistory later, we might be modifying
	// the same underlying array that is part of the state we are trying to save.
	// This would corrupt the history for the UnmakeMove operation.
	stateToSave := GameState{
		CastlingRights:  g.CastlingRights,
		EnPassantTarget: g.EnPassantTarget,
		HalfMoveClock:   g.HalfMoveClock,
		ZobristHash:     g.ZobristHash,
		CapturedPiece:   targetPiece,
		PositionHistory: make([]uint64, len(g.PositionHistory)),
	}
	copy(stateToSave.PositionHistory, g.PositionHistory)

	// Handle en passant capture (remove attacked pawn)
	if isEnPassant {
		// Captured pawn is at [ToCol, FromRow]
		captureSq := GetSq(GetFile(move.To), GetRank(move.From))
		stateToSave.CapturedPiece = g.Board.GetPiece(captureSq)

		isCapture = true
		g.Board.SetPiece(captureSq, NoPiece)

		// Update Hash: Remove captured EP pawn
		epPawn := stateToSave.CapturedPiece
		hash ^= zobristPiece[epPawn.Color()][epPawn.Type()][captureSq]
	}

	// Execute the move
	g.Board.MovePiece(move.From, move.To)

	// Handle castling rook movement
	if isCastling {
		rank := GetRank(move.From)
		if move.To > move.From { // Kingside
			// Move rook from H-file (7) to F-file (5)
			g.Board.MovePiece(GetSq(FileH, rank), GetSq(FileF, rank))

			// Update Hash: Move Rook
			rook := g.Board.GetPiece(GetSq(FileF, rank))
			oldRookSq := GetSq(FileH, rank)
			newRookSq := GetSq(FileF, rank)
			hash ^= zobristPiece[rook.Color()][rook.Type()][oldRookSq]
			hash ^= zobristPiece[rook.Color()][rook.Type()][newRookSq]
		} else { // Queenside
			// Move rook from A-file (0) to D-file (3)
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
		// We already added the Pawn at dstSq above. Remove it.
		hash ^= zobristPiece[piece.Color()][piece.Type()][move.To]
		// Add promoted piece
		promPiece := move.PromotionPiece
		hash ^= zobristPiece[promPiece.Color()][promPiece.Type()][move.To]
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
	g.LastMove = &move

	// Increment full move number after black moves
	if g.CurrentTurn == Black {
		g.FullMoveNumber++
	}

	// Switch turns
	g.SwitchTurn()

	// XOR in new state (Castling, EP)
	hash ^= zobristCastling[g.CastlingRights]
	if g.EnPassantTarget != -1 {
		hash ^= zobristEnPassant[GetFile(g.EnPassantTarget)]
	} else {
		hash ^= zobristEnPassant[8]
	}

	g.ZobristHash = hash

	// Update position history for repetition check
	// If the move is irreversible (capture or pawn move), we clear the history
	// because previous positions can never be reached again.
	// This is an optimization and also correct according to FIDE rules for 3-fold repetition.
	if isCapture || isPawnMove {
		g.PositionHistory = []uint64{g.ZobristHash}
	} else {
		g.PositionHistory = append(g.PositionHistory, g.ZobristHash)
	}

	// Push state to history
	g.StateHistory = append(g.StateHistory, stateToSave)

	return true
}

// updateCastlingRights revokes castling rights if king or rook moves or is captured
func (g *Game) updateCastlingRights(move Move, piece Piece, targetPiece Piece) {
	fromFile := GetFile(move.From)
	fromRank := GetRank(move.From)
	toFile := GetFile(move.To)
	toRank := GetRank(move.To)

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

	// If a rook is captured, revoke castling rights
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
		g.LastMove = new(g.MoveHistory[len(g.MoveHistory)-1])
	} else {
		g.LastMove = nil
	}

	// Restore pieces on board
	// 1. Move the piece back from To -> From
	movedPiece := g.Board.GetPiece(move.To)
	g.Board.MovePiece(move.To, move.From)

	// 2. If it was a promotion, revert the piece at From to a Pawn
	if move.PromotionPiece != NoPiece {
		pawn := Pawn.FromColor(move.PromotionPiece.Color())
		g.Board.SetPiece(move.From, pawn)
		movedPiece = pawn // Update for En Passant check below
	}

	// 3. Restore captured piece
	if state.CapturedPiece != NoPiece {
		// Check if it was En Passant
		if movedPiece.Type() == Pawn && move.To == g.EnPassantTarget {
			// En Passant capture: restore pawn at (ToCol, FromRow)
			captureSq := GetSq(GetFile(move.To), GetRank(move.From))
			g.Board.SetPiece(captureSq, state.CapturedPiece)
		} else {
			// Normal capture: restore piece at To
			g.Board.SetPiece(move.To, state.CapturedPiece)
		}
	}

	// Castling undo
	// If King moved 2 squares, move the Rook back
	if movedPiece.Type() == King && (move.To == move.From+2 || move.To == move.From-2) {
		rank := GetRank(move.From)
		if move.To > move.From { // Kingside castling (e1->g1 or e8->g8)
			// Rook moved from H to F. Move back F -> H.
			g.Board.MovePiece(GetSq(FileF, rank), GetSq(FileH, rank))
		} else { // Queenside castling (e1->c1 or e8->c8)
			// Rook moved from A to D. Move back D -> A.
			g.Board.MovePiece(GetSq(FileD, rank), GetSq(FileA, rank))
		}
	}
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
	currentHash := g.ZobristHash
	count := 0
	for _, hash := range g.PositionHistory {
		if hash == currentHash {
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
	// Fast check: If there are any Pawns, Rooks, or Queens, it's not insufficient material.
	// We can check all at once using bitwise OR.
	heavyPieces := g.Board.Pieces[White][Pawn] | g.Board.Pieces[White][Rook] | g.Board.Pieces[White][Queen] |
		g.Board.Pieces[Black][Pawn] | g.Board.Pieces[Black][Rook] | g.Board.Pieces[Black][Queen]

	if heavyPieces != 0 {
		return false
	}

	// Count minor pieces using Population Count (very fast)
	whiteKnights := g.Board.Pieces[White][Knight].Count()
	blackKnights := g.Board.Pieces[Black][Knight].Count()
	whiteBishops := g.Board.Pieces[White][Bishop].Count()
	blackBishops := g.Board.Pieces[Black][Bishop].Count()

	whiteMinors := whiteKnights + whiteBishops
	blackMinors := blackKnights + blackBishops
	totalMinors := whiteMinors + blackMinors

	// King vs King (No minors)
	if totalMinors == 0 {
		return true
	}

	// King + Knight vs King OR King + Bishop vs King
	// (Exactly one minor piece on the board)
	if totalMinors == 1 {
		return true
	}

	// King + Bishop vs King + Bishop (same color squares)
	if whiteBishops == 1 && blackBishops == 1 && whiteKnights == 0 && blackKnights == 0 {
		// Get the square of the white bishop
		wbBB := g.Board.Pieces[White][Bishop]
		wbSq := wbBB.PopLSB() // Safe because we know count is 1

		// Get the square of the black bishop
		bbBB := g.Board.Pieces[Black][Bishop]
		bbSq := bbBB.PopLSB()

		wbColor := (GetRank(wbSq) + GetFile(wbSq)) % 2
		bbColor := (GetRank(bbSq) + GetFile(bbSq)) % 2

		return wbColor == bbColor
	}

	return false
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
	newG := &Game{
		Board:          g.Board.Clone(),
		CurrentTurn:    g.CurrentTurn,
		MoveHistory:    make([]Move, len(g.MoveHistory)),
		LastMove:       nil,
		FullMoveNumber: g.FullMoveNumber,
		StateHistory:   make([]GameState, len(g.StateHistory)),
		GameState:      g.GameState,
	}
	copy(newG.MoveHistory, g.MoveHistory)
	copy(newG.StateHistory, g.StateHistory)
	if g.LastMove != nil {
		lm := *g.LastMove
		newG.LastMove = &lm
	}

	// Deep copy the PositionHistory slice within the GameState to prevent aliasing issues.
	newG.GameState.PositionHistory = make([]uint64, len(g.GameState.PositionHistory))
	copy(newG.GameState.PositionHistory, g.GameState.PositionHistory)

	return newG
}
