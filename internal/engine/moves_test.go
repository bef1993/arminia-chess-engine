package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoveGenerator(t *testing.T) {
	board := NewBoard()
	mg := NewMoveGenerator(board)

	assert.NotNil(t, mg, "NewMoveGenerator returned nil")
	assert.Equal(t, board, mg.Board, "MoveGenerator board does not match provided board")
}

func TestGeneratePawnMoves(t *testing.T) {
	board := NewBoard()
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("e2"))

	assert.Len(t, moves, 2, "Expected 2 moves for white pawn at starting position")
	assert.Contains(t, moves, Move{From: Sq("e2"), To: Sq("e3"), PromotionPiece: NoPiece}, "Expected move from e2 to e3")
	assert.Contains(t, moves, Move{From: Sq("e2"), To: Sq("e4"), PromotionPiece: NoPiece}, "Expected move from e2 to e4")
}

func TestGenerateKnightMoves(t *testing.T) {
	board := NewBoard()
	mg := NewMoveGenerator(board)

	// Test white knight at (1, 7) - should have 2 moves (can't move forward much)
	moves := mg.GenerateMovesForPiece(Sq("b1"))

	assert.Len(t, moves, 2, "Expected 2 moves for white knight at starting position")

	// Remove pawn at d2 to allow knight to move there
	board.SetPieceAt("d2", NoPiece)

	moves = mg.GenerateMovesForPiece(Sq("b1"))

	// Knight should have 3 moves (a3, c3, d2)
	assert.Len(t, moves, 3, "Expected 3 moves for knight after removing pawn at d2")
}

func TestGenerateBishopMovesFromMiddle(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white bishop in the middle of an empty board
	board.SetPieceAt("d5", WhiteBishop)
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// Bishop in the middle should have 13 moves (4 diagonals with 3 or 4 squares each)
	assert.Len(t, moves, 13, "Expected 13 moves for bishop in center")

	// Verify all moves are diagonal
	for _, move := range moves {
		colDiff := (move.To % 8) - (move.From % 8)
		rowDiff := (move.To / 8) - (move.From / 8)

		assert.True(t, colDiff != 0 && rowDiff != 0 && (colDiff == rowDiff || colDiff == -rowDiff), "Invalid bishop move")
	}
}

func TestGenerateRookMovesFromMiddle(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white rook in the middle of an empty board
	board.SetPieceAt("d5", WhiteRook)
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// Rook in the middle should have 14 moves (4 directions with 3 or 4 squares each)
	assert.Len(t, moves, 14, "Expected 14 moves for rook in center")

	// Verify all moves are horizontal or vertical
	for _, move := range moves {
		colDiff := (move.To % 8) - (move.From % 8)
		rowDiff := (move.To / 8) - (move.From / 8)

		assert.True(t, (colDiff == 0 && rowDiff != 0) || (colDiff != 0 && rowDiff == 0), "Invalid rook move")
	}
}

func TestGenerateQueenMovesFromMiddle(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white queen in the middle of an empty board
	board.SetPieceAt("d5", WhiteQueen)
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// Queen should have 27 moves (combination of rook and bishop moves)
	assert.Len(t, moves, 27, "Expected 27 moves for queen in center")
}

func TestGenerateKingMoves(t *testing.T) {
	board := NewBoard()
	// Place a white king in the middle of an empty board
	board.SetPieceAt("d5", WhiteKing)
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// King in the middle should have 8 moves
	assert.Len(t, moves, 8, "Expected 8 moves for king in center")

	// Verify all moves are 1 square away
	for _, move := range moves {
		colDiff := (move.To % 8) - (move.From % 8)
		rowDiff := (move.To / 8) - (move.From / 8)

		assert.True(t, colDiff >= -1 && colDiff <= 1 && rowDiff >= -1 && rowDiff <= 1, "King moved more than 1 square")
	}
}

func TestGenerateKingMovesCorner(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white king in the corner
	board.SetPieceAt("a1", WhiteKing)
	mg := NewMoveGenerator(board)

	moves := mg.GenerateMovesForPiece(Sq("a1"))

	// King in corner should have 3 moves
	assert.Len(t, moves, 3, "Expected 3 moves for king in corner")
}

func TestGenerateMovesForNilPiece(t *testing.T) {
	board := NewBoard()
	mg := NewMoveGenerator(board)

	// Try to generate moves for empty square
	moves := mg.GenerateMovesForPiece(Sq("d4"))

	assert.Empty(t, moves, "Expected 0 moves for empty square")
}

func TestPawnCapturesDiagonally(t *testing.T) {
	board := NewEmptyBoard()

	// Place a white pawn
	board.SetPieceAt("d4", WhitePawn)
	// Place black pieces to capture
	board.SetPieceAt("c5", BlackPawn)
	board.SetPieceAt("e5", BlackPawn)

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d4"))

	// Should have 3 moves: 1 forward, 2 captures
	assert.Len(t, moves, 3, "Expected 3 moves (1 forward + 2 captures)")

	// Verify at least one move goes to (2, 3) and one to (4, 3)
	captureFound := [2]bool{false, false}
	for _, move := range moves {
		if move.To == Sq("c5") {
			captureFound[0] = true
		}
		if move.To == Sq("e5") {
			captureFound[1] = true
		}
	}

	assert.True(t, captureFound[0], "Pawn should be able to capture diagonally left")
	assert.True(t, captureFound[1], "Pawn should be able to capture diagonally right")
}

func TestBlockedPawnCannotMove(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white pawn
	board.SetPieceAt("d2", WhitePawn)
	// Place blocking pieces
	board.SetPieceAt("d3", BlackPawn) // Block forward move
	board.SetPieceAt("d4", BlackPawn) // Block double move

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d2"))

	assert.Empty(t, moves, "Expected 0 moves for blocked pawn")
}

func TestBishopBlockedByOwnPiece(t *testing.T) {
	board := NewEmptyBoard()

	board.SetPieceAt("d5", WhiteBishop)
	// Block one diagonal
	board.SetPieceAt("f3", WhitePawn)

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// Verify f3 and beyond on that diagonal are NOT in the moves
	for _, move := range moves {
		assert.False(t, move.To == Sq("f3"), "Bishop should not move to square occupied by own piece")
		assert.False(t, move.To == Sq("g2"), "Bishop should not move past own piece")
		assert.False(t, move.To == Sq("h1"), "Bishop should not move past own piece")
	}


	assert.Equal(t, 10, len(moves), "Bishop should have exactly 10 available moves")
}

func TestBishopStopsAtEnemyPiece(t *testing.T) {
	board := NewEmptyBoard()

	board.SetPieceAt("d5", WhiteBishop)
	// Place black piece to capture
	board.SetPieceAt("f3", BlackPawn)
	// Place another piece beyond to verify bishop stops at capture
	board.SetPieceAt("g2", BlackPawn)

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d5"))

	assert.Contains(t, moves, Move{From: Sq("d5"), To: Sq("f3"), PromotionPiece: NoPiece}, "Expected move to capture enemy piece at f3")
	assert.NotContains(t, moves, Move{From: Sq("d5"), To: Sq("g2"), PromotionPiece: NoPiece}, "Expected no move to square occupied by enemy piece")
	assert.Equal(t, 11, len(moves), "Bishop should have exactly 11 available moves (including capture)")
}

func TestRookBlockedAlongRank(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white rook at (3, 3)
	board.SetPieceAt("d5", WhiteRook)
	// Block the rank to the right at (6, 3)
	board.SetPieceAt("g5", BlackPawn)

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d5"))
	// Rook should reach (6,3) but NOT (7,3)
	canCapture := false
	canPassThrough := false

	for _, move := range moves {
		if move.To == Sq("g5") {
			canCapture = true
		}
		if move.To == Sq("h5") {
			canPassThrough = true
		}
	}

	assert.True(t, canCapture, "Rook should be able to capture enemy piece on rank")
	assert.False(t, canPassThrough, "Rook should not move past captured piece on rank")
}

func TestRookBlockedAlongFile(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white rook at (3, 4)
	board.SetPieceAt("d4", WhiteRook)
	// Block the file upward at (3, 1)
	board.SetPieceAt("d7", BlackKnight)
	// Place another piece beyond to verify rook stops at capture
	board.SetPieceAt("d8", BlackPawn)

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d4"))

	// Rook should reach (3,1) but NOT (3,0)
	canCapture := false
	canPassThrough := false

	for _, move := range moves {
		if move.To == Sq("d7") {
			canCapture = true
		}
		if move.To == Sq("d8") {
			canPassThrough = true
		}
	}

	assert.True(t, canCapture, "Rook should be able to capture enemy piece on file")
	assert.False(t, canPassThrough, "Rook should not move past captured piece on file")
}

func TestQueenMultipleBlockingDirections(t *testing.T) {
	board := NewEmptyBoard()
	// Place a white queen at (3, 3)
	board.SetPieceAt("d5", WhiteQueen)
	// Block different directions with different pieces
	board.SetPieceAt("d7", BlackPawn) // Block downward file
	board.SetPieceAt("f5", WhitePawn) // Block rightward rank with own piece
	board.SetPieceAt("f3", BlackPawn) // Block diagonal
	board.SetPieceAt("b5", BlackPawn) // Block leftward rank

	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(Sq("d5"))

	// Verify specific blocking behavior
	canCaptureDown := false
	canReachRightOwn := false
	canCaptureDiagonal := false
	canCaptureLeft := false

	for _, move := range moves {
		if move.To == Sq("d7") {

			canCaptureDown = true
		}
		if move.To == Sq("f5") {
			canReachRightOwn = true // Should NOT be in moves
		}
		if move.To == Sq("f3") {
			canCaptureDiagonal = true
		}
		if move.To == Sq("b5") {
			canCaptureLeft = true
		}
	}

	assert.True(t, canCaptureDown, "Queen should capture downward")
	assert.False(t, canReachRightOwn, "Queen should not move to own piece")
	assert.True(t, canCaptureDiagonal, "Queen should capture on diagonal")
	assert.True(t, canCaptureLeft, "Queen should capture leftward")
}

func TestCastlingMoves(t *testing.T) {
	board := NewEmptyBoard()

	// Setup White King and Rooks for testing
	board.SetPieceAt("e1", WhiteKing)
	board.SetPieceAt("h1", WhiteRook)
	board.SetPieceAt("a1", WhiteRook)

	// 1. Test basic castling rights (both sides available)
	// We use NewMoveGeneratorFull to inject castling rights
	mg := NewMoveGeneratorFull(board, -1, WhiteKingside|WhiteQueenside)
	moves := mg.GenerateMovesForPiece(Sq("e1"))

	hasKingside := false
	hasQueenside := false
	for _, m := range moves {
		if m.From == Sq("e1") {
			if m.To == Sq("g1") {
				hasKingside = true
			}
			if m.To == Sq("c1") {
				hasQueenside = true
			}
		}
	}

	assert.True(t, hasKingside, "Expected White Kingside castling move (e1->g1)")
	assert.True(t, hasQueenside, "Expected White Queenside castling move (e1->c1)")

	// 2. Test blocked path
	board.SetPieceAt("f1", WhiteBishop)
	mg = NewMoveGeneratorFull(board, -1, WhiteKingside)
	moves = mg.GenerateMovesForPiece(Sq("e1"))
	for _, m := range moves {
		assert.False(t, m.To == Sq("g1"), "Should not castle kingside when blocked at f1")
	}
	board.RemovePieceAt("f1")

	// 3. Test path attacked (castling through check)
	// Place a black rook at f2 to attack f1
	board.SetPieceAt("f7", BlackRook)
	mg = NewMoveGeneratorFull(board, -1, WhiteKingside)
	moves = mg.GenerateMovesForPiece(Sq("e1"))

	for _, m := range moves {
		assert.False(t, m.To == Sq("g1"), "Should not castle kingside when f1 is attacked")
	}
	board.RemovePieceAt("f7")

	// 4. Test King in check
	// Place a black rook at e2 to attack e1 (King's square)
	board.SetPieceAt("e7", BlackRook)
	mg = NewMoveGeneratorFull(board, -1, WhiteKingside)
	moves = mg.GenerateMovesForPiece(Sq("e1"))
	for _, m := range moves {
		assert.False(t, m.To == Sq("g1"), "Should not castle when King is in check")
	}
}

func TestCastlingQueensideRules(t *testing.T) {
	board := NewEmptyBoard()
	// Setup White King and Rook
	board.SetPieceAt("e1", WhiteKing)
	board.SetPieceAt("a1", WhiteRook)

	// Helper to check if queenside castling is generated
	hasQueensideCastling := func() bool {
		mg := NewMoveGeneratorFull(board, -1, WhiteQueenside)
		moves := mg.GenerateMovesForPiece(Sq("e1"))
		for _, m := range moves {
			if m.To == Sq("c1") {
				return true
			}
		}
		return false
	}

	// 1. Basic valid castling
	assert.True(t, hasQueensideCastling(), "Should allow queenside castling on empty board")

	// 2. Blocked path tests
	// Block d1
	board.SetPieceAt("d1", WhitePawn)
	assert.False(t, hasQueensideCastling(), "Should not castle when d1 is blocked")
	board.RemovePieceAt("d1")

	// Block c1
	board.SetPieceAt("c1", WhitePawn)
	assert.False(t, hasQueensideCastling(), "Should not castle when c1 is blocked")
	board.RemovePieceAt("c1")

	// Block b1 (Rook path)
	board.SetPieceAt("b1", WhitePawn)
	assert.False(t, hasQueensideCastling(), "Should not castle when b1 is blocked")
	board.RemovePieceAt("b1")

	// 3. Attacked path tests
	// Attack d1 (pass-through square)
	board.SetPieceAt("d8", BlackRook) // Rook at d8 attacks d1
	assert.False(t, hasQueensideCastling(), "Should not castle when d1 is attacked")
	board.RemovePieceAt("d8")

	// Attack c1 (destination square)
	board.SetPieceAt("c8", BlackRook) // Rook at c8 attacks c1
	assert.False(t, hasQueensideCastling(), "Should not castle when c1 is attacked")
	board.RemovePieceAt("c8")

	// Attack b1 (rook square/path) - Should still be allowed!
	// The King does not pass through b1, so it doesn't matter if it's attacked.
	board.SetPieceAt("b8", BlackRook) // Rook at b8 attacks b1
	assert.True(t, hasQueensideCastling(), "Should allow castling even if b1 is attacked (King doesn't pass through it)")
}

func TestCastlingBlack(t *testing.T) {
	board := NewEmptyBoard()
	board.SetPieceAt("e8", BlackKing)
	board.SetPieceAt("h8", BlackRook)

	// Test blocked path for Black Kingside
	board.SetPieceAt("f8", BlackBishop)
	mg := NewMoveGeneratorFull(board, -1, BlackKingside)
	moves := mg.GenerateMovesForPiece(Sq("e8"))
	for _, m := range moves {
		assert.False(t, m.To == Sq("g8"), "Black should not castle kingside when blocked")
	}
}
