package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	game := NewGame()

	assert.NotNil(t, game, "NewGame returned nil")
	assert.NotNil(t, game.Board, "Game board is nil")
	assert.Equal(t, White, game.CurrentTurn, "Expected current turn to be White")
	assert.Empty(t, game.MoveHistory, "Expected empty move history")
}

func TestGameBoardInitialization(t *testing.T) {
	game := NewGame()

	// Check that the board is properly initialized with starting position
	// White pawn on (4, 6)
	piece := game.Board.GetPieceAt("e2")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Pawn, piece.Type())
	assert.Equal(t, White, piece.Color())

	// Black pawn on (4, 1)
	piece = game.Board.GetPieceAt("e7")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Pawn, piece.Type())
	assert.Equal(t, Black, piece.Color())

	// White king on (4, 7)
	piece = game.Board.GetPieceAt("e1")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, King, piece.Type())
	assert.Equal(t, White, piece.Color())

	// Black king on (4, 0)
	piece = game.Board.GetPieceAt("e8")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, King, piece.Type())
	assert.Equal(t, Black, piece.Color())
}

func TestSwitchTurn(t *testing.T) {
	game := NewGame()

	assert.Equal(t, White, game.CurrentTurn, "Expected initial turn to be White")

	game.SwitchTurn()
	assert.Equal(t, Black, game.CurrentTurn, "Expected turn to be Black after switch")

	game.SwitchTurn()
	assert.Equal(t, White, game.CurrentTurn, "Expected turn to be White after second switch")
}

func TestSwitchTurnMultipleTimes(t *testing.T) {
	game := NewGame()

	for i := 0; i < 10; i++ {
		expectedColor := White
		if i%2 == 1 {
			expectedColor = Black
		}

		assert.Equal(t, expectedColor, game.CurrentTurn, "Iteration %d", i)

		game.SwitchTurn()
	}
}

func TestGameEnPassantTarget(t *testing.T) {
	game := NewGame()

	// Initially no en passant target
	assert.Equal(t, -1, game.EnPassantTargetCol, "Expected no en passant target col initially")
	assert.Equal(t, -1, game.EnPassantTargetRow, "Expected no en passant target row initially")

	// Move white pawn 2 squares forward (e2 to e4 = col 4, row 6 to row 4)
	move := NewMove(FileE, Rank2, FileE, Rank4)
	game.ExecuteMove(move)

	// En passant target should be set to (4, 5)
	assert.Equal(t, FileE, game.EnPassantTargetCol, "Expected en passant target col at FileE")
	assert.Equal(t, Rank3, game.EnPassantTargetRow, "Expected en passant target row at Rank3")

	// Make a non-pawn move to clear en passant target
	blackMove := NewMove(1, 0, 2, 2) // Move black knight
	game.ExecuteMove(blackMove)

	// En passant target should be cleared
	assert.Equal(t, -1, game.EnPassantTargetCol, "Expected en passant target col cleared")
	assert.Equal(t, -1, game.EnPassantTargetRow, "Expected en passant target row cleared")
}

func TestGameExecuteMove(t *testing.T) {
	game := NewGame()

	// Initial state
	assert.Empty(t, game.MoveHistory, "Expected empty move history at start")
	assert.Nil(t, game.LastMove, "Expected nil LastMove at start")

	// Execute a white pawn move
	move1 := NewMove(FileE, Rank2, FileE, Rank3) // e2 to e3
	success := game.ExecuteMove(move1)

	assert.True(t, success, "Failed to execute valid move")

	// Check move was recorded
	assert.Len(t, game.MoveHistory, 1, "Expected 1 move in history")

	assert.NotNil(t, game.LastMove, "LastMove should be set")
	assert.Equal(t, FileE, game.LastMove.FromCol, "LastMove FromCol incorrect")

	// Check turn switched
	assert.Equal(t, Black, game.CurrentTurn, "Expected turn to be Black after white move")

	// Full move number should still be 1
	assert.Equal(t, 1, game.FullMoveNumber, "Expected full move number 1")

	// Execute a black move
	move2 := NewMove(FileE, Rank7, FileE, Rank5) // e7 to e5
	game.ExecuteMove(move2)

	// Now full move number should be 2
	assert.Equal(t, 2, game.FullMoveNumber, "Expected full move number 2 after black moves twice")

	assert.Equal(t, White, game.CurrentTurn, "Expected turn to be White after black moves")
}

func TestGameHalfMoveClockCapture(t *testing.T) {
	game := NewGame()

	// Setup: clear and place specific pieces for testing
	game.Board.Clear()
	game.Board.SetPieceAt("e5", WhiteRook)
	game.Board.SetPieceAt("d5", BlackPawn)
	game.CurrentTurn = White

	// Regular non-capture non-pawn move
	game.ExecuteMove(NewMove(FileE, Rank5, FileE, Rank6)) // Rook e5-e4 (not capturing, not pawn)

	assert.Equal(t, 1, game.HalfMoveClock, "Expected half-move clock to be 1 after non-capture move")

	// Another regular move
	game.CurrentTurn = Black
	game.ExecuteMove(NewMove(FileD, Rank5, FileD, Rank6)) // Pawn d5-d4 (pawn move resets clock)

	assert.Equal(t, 0, game.HalfMoveClock, "Expected half-move clock to reset to 0 on pawn move")

	// Non-pawn move by white
	game.CurrentTurn = White
	game.ExecuteMove(NewMove(FileE, Rank6, FileE, Rank7)) // Rook e4-e3

	assert.Equal(t, 1, game.HalfMoveClock, "Expected half-move clock to be 1 after another non-capture move")

	// Capture move
	game.CurrentTurn = Black
	game.ExecuteMove(NewMove(FileD, Rank6, FileE, Rank7)) // Pawn d4 captures rook on e3

	assert.Equal(t, 0, game.HalfMoveClock, "Expected half-move clock to reset to 0 on capture")
}

func TestGameHalfMoveClockPawnMove(t *testing.T) {
	game := NewGame()

	// Half-move clock starts at 0
	assert.Equal(t, 0, game.HalfMoveClock, "Expected half-move clock 0")

	// Pawn move should reset half-move clock
	game.ExecuteMove(NewMove(FileE, Rank2, FileE, Rank4)) // White pawn e2-e4 (double move)

	assert.Equal(t, 0, game.HalfMoveClock, "Expected half-move clock to reset on pawn move")
}

func TestGameDrawByFiftyMoveRule(t *testing.T) {
	game := NewGame()

	// Manual setup: set half-move clock to 99 (just before draw)
	game.HalfMoveClock = 99

	assert.False(t, game.IsDrawByFiftyMoveRule(), "Should not be draw at 99 half-moves")

	// Set to 100 (50 full moves)
	game.HalfMoveClock = 100

	assert.True(t, game.IsDrawByFiftyMoveRule(), "Should be draw at 100 half-moves (50-move rule)")
}

func TestEnPassantMoveGeneration(t *testing.T) {
	board := NewEmptyBoard()
	// Set up white pawn on e5 and black pawn on d7
	board.SetPieceAt("e5", WhitePawn)
	board.SetPieceAt("d7", BlackPawn)

	// Black pawn moves to d5 (double move), creating en passant target at d6
	board.MovePiece(FileD, Rank7, FileD, Rank5) // d7 to d5

	// Now generate moves for white with en passant target
	mg := NewMoveGeneratorWithEnPassant(board, FileD, Rank6) // d6 is en passant target
	moves := mg.GenerateMovesForPiece(FileE, Rank5)

	// Should have regular moves plus en passant capture (move to 3,2)
	hasEnPassant := false
	for _, move := range moves {
		if move.ToCol == FileD && move.ToRow == Rank6 {
			hasEnPassant = true
		}
	}

	assert.True(t, hasEnPassant, "Expected en passant capture move to (3, 2) to be generated")
}

func TestEnPassantCaptureExecution(t *testing.T) {
	game := NewGame()

	// Set up: white pawn on e5, black pawn on d7
	game.Board.Clear()
	game.Board.SetPieceAt("e5", WhitePawn)
	game.Board.SetPieceAt("d7", BlackPawn)
	game.CurrentTurn = Black

	// Black moves pawn to d5 (double move)
	move := NewMove(FileD, Rank7, FileD, Rank5)
	game.ExecuteMove(move)

	// Check piece is at d5
	piece := game.Board.GetPieceAt("d5")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Pawn, piece.Type())
	assert.Equal(t, Black, piece.Color())

	// Check en passant target is set
	assert.Equal(t, FileD, game.EnPassantTargetCol)
	assert.Equal(t, Rank6, game.EnPassantTargetRow)

	// White captures en passant (e5 to d6 en passant)
	epMove := NewMove(FileE, Rank5, FileD, Rank6)
	game.ExecuteMove(epMove)

	// Check white pawn is at d6
	piece = game.Board.GetPieceAt("d6")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, WhitePawn, piece)

	// Check black pawn was captured (should be at original row, not at capture square)
	piece = game.Board.GetPieceAt("d5")
	assert.Equal(t, NoPiece, piece, "Black pawn should have been captured")
}

func TestPawnPromotionForward(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Place white pawn on e2 (col 4, row 6), one move away from promotion
	game.Board.SetPieceAt("e7", WhitePawn)

	// Move pawn forward to promotion rank (e1 = row 0)
	move := NewPromotionMove(FileE, Rank7, FileE, Rank8, WhiteQueen)
	game.ExecuteMove(move)

	// Check pawn was replaced with Queen
	piece := game.Board.GetPieceAt("e8")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Queen, piece.Type())
	assert.Equal(t, White, piece.Color())

	// Check old square is empty
	piece = game.Board.GetPieceAt("e7")
	assert.Equal(t, NoPiece, piece, "Pawn square should be empty after promotion")
}

func TestPawnPromotionCapture(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Place white pawn on e2 (col 4, row 1), one move away from promotion
	game.Board.SetPieceAt("e7", WhitePawn)
	// Place black pawn on d1 (col 3, row 0) to capture
	game.Board.SetPieceAt("d8", BlackPawn)

	// Pawn captures and promotes to Rook
	move := NewPromotionMove(FileE, Rank7, FileD, Rank8, WhiteRook)
	game.ExecuteMove(move)

	// Check pawn was replaced with Rook on capture square
	piece := game.Board.GetPieceAt("d8")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Rook, piece.Type())
	assert.Equal(t, White, piece.Color())

	// Check old square is empty
	piece = game.Board.GetPieceAt("e7")
	assert.Equal(t, NoPiece, piece, "Pawn square should be empty after promotion capture")
}

func TestPawnPromotionMoveGeneration(t *testing.T) {
	board := NewEmptyBoard()

	// Place white pawn on e2 (col 4, row 1), one move away from promotion
	board.SetPieceAt("e7", WhitePawn)

	// Generate moves
	mg := NewMoveGenerator(board)
	moves := mg.GenerateMovesForPiece(FileE, Rank7)

	// Should have 4 promotion moves (Queen, Rook, Bishop, Knight)
	promotionMoves := []Move{}
	for _, move := range moves {
		if move.PromotionPiece != 0 {
			promotionMoves = append(promotionMoves, move)
		}
	}

	assert.Len(t, promotionMoves, 4, "Expected 4 promotion moves")

	// Check we have all 4 promotion pieces
	promotionPieces := make(map[PieceType]bool)
	for _, move := range promotionMoves {
		promotionPieces[move.PromotionPiece.Type()] = true
	}

	expected := []PieceType{Queen, Rook, Bishop, Knight}
	for _, piece := range expected {
		assert.True(t, promotionPieces[piece], "Missing promotion to %v", piece)
	}
}

func TestBlackPawnPromotionForward(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Place black pawn on e7 (col 4, row 6), one move away from promotion (row 7)
	game.Board.SetPieceAt("e2", BlackPawn)
	game.CurrentTurn = Black

	// Move pawn forward to promotion rank (e8 = row 7)
	move := NewPromotionMove(FileE, Rank2, FileE, Rank1, BlackQueen)
	game.ExecuteMove(move)

	// Check pawn was replaced with Queen
	piece := game.Board.GetPieceAt("e1")
	assert.NotEqual(t, NoPiece, piece)
	assert.Equal(t, Queen, piece.Type())
	assert.Equal(t, Black, piece.Color())

	// Check old square is empty
	piece = game.Board.GetPieceAt("e2")
	assert.Equal(t, NoPiece, piece, "Pawn square should be empty after promotion")
}

func TestCastlingExecution(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup White Kingside Castling
	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("h1", WhiteRook)
	game.CastlingRights = WhiteKingside

	// Execute castling move (e1 to g1)
	move := NewMove(FileE, Rank1, FileG, Rank1)
	game.ExecuteMove(move)

	// Check King position
	king := game.Board.GetPieceAt("g1")
	assert.NotEqual(t, NoPiece, king)
	assert.Equal(t, King, king.Type())

	// Check Rook position (should be at f1)
	rook := game.Board.GetPieceAt("f1")
	assert.NotEqual(t, NoPiece, rook)
	assert.Equal(t, Rook, rook.Type())

	// Check castling rights revoked
	assert.Zero(t, game.CastlingRights&WhiteKingside, "Castling rights not revoked after castling")
}

func TestCastlingExecutionQueenside(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup White Queenside Castling
	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("a1", WhiteRook)
	game.CastlingRights = WhiteQueenside

	// Execute castling move (e1 to c1)
	move := NewMove(FileE, Rank1, FileC, Rank1)
	game.ExecuteMove(move)

	// Check King position
	king := game.Board.GetPieceAt("c1")
	assert.NotEqual(t, NoPiece, king)
	assert.Equal(t, King, king.Type())

	// Check Rook position (should be at d1)
	rook := game.Board.GetPieceAt("d1")
	assert.NotEqual(t, NoPiece, rook)
	assert.Equal(t, Rook, rook.Type())

	// Check castling rights revoked
	assert.Zero(t, game.CastlingRights&WhiteQueenside, "Castling rights not revoked after queenside castling")
}

func TestIsCheckmate(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup Mate (King blocked by own pawns and attacked by Queen)
	// White King at e1, White Pawns at d1, f1
	// Black Queen at e3
	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("d1", WhitePawn)
	game.Board.SetPieceAt("f1", WhitePawn)
	game.Board.SetPieceAt("e3", BlackQueen)

	assert.True(t, game.IsCheckmate(), "Expected checkmate for White")
	assert.False(t, game.IsStalemate(), "Checkmate is not stalemate")
}

func TestIsStalemate(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup Stalemate
	// Black King at a8 (0,0)
	// White Queen at c7 (2,1) -> Attacks a7, b7, b8, c8, d8.
	// King at a8 is NOT in check (knight jump away from c7), but has no moves.
	game.Board.SetPieceAt("a8", BlackKing)
	game.Board.SetPieceAt("c7", WhiteQueen)
	game.Board.SetPieceAt("h1", WhiteKing)

	game.CurrentTurn = Black
	assert.True(t, game.IsStalemate(), "Expected stalemate for Black")
	assert.False(t, game.IsCheckmate(), "Stalemate is not checkmate")
}

func TestNotCheckmateWhenCanBlock(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup: White King trapped by own pawns and attacked square
	// White King at e1
	game.Board.SetPieceAt("e1", WhiteKing)

	// Block King with own pawns
	game.Board.SetPieceAt("d1", WhitePawn)
	game.Board.SetPieceAt("f1", WhitePawn)
	game.Board.SetPieceAt("d2", WhitePawn)
	game.Board.SetPieceAt("f2", WhitePawn)

	// Black Rook at e8 delivers check
	game.Board.SetPieceAt("e8", BlackRook)

	// Black Pawn at d3 attacks e2, preventing King from escaping forward
	game.Board.SetPieceAt("d3", BlackPawn)

	// White Knight at c3 can block at e2 or e4
	game.Board.SetPieceAt("c3", WhiteKnight)

	assert.False(t, game.IsCheckmate(), "Should not be checkmate if Knight can block")
}

func TestGetLegalMovesFiltersCheckMoves(t *testing.T) {
	tests := []struct {
		name      string
		setupFn   func(*Game)
		color     Color
		expectMin int // At least this many legal moves
		expectMax int // At most this many legal moves
	}{
		{
			name:      "white has moves at start",
			setupFn:   func(g *Game) {},
			color:     White,
			expectMin: 20, // 16 pawn moves + 4 knight moves
			expectMax: 20,
		},
		{
			name: "black has moves at start",
			setupFn: func(g *Game) {
				g.CurrentTurn = Black
			},
			color:     Black,
			expectMin: 20, // 16 pawn moves + 4 knight moves
			expectMax: 20,
		},
		{
			name: "king cannot move into check",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e4", WhiteKing)
				g.Board.SetPieceAt("e7", BlackRook)
				// King can move left/right but not up into check
			},
			color:     White,
			expectMin: 5, // Limited moves to avoid check
			expectMax: 7,
		},
		{
			name: "king has limited moves in corner",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("a8", WhiteKing)
				// King in corner can only move to 3 adjacent squares (inside board)
				// Plus moving to diagonals and horizontally/vertically
			},
			color:     White,
			expectMin: 3,
			expectMax: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			tt.setupFn(game)

			got := game.GetLegalMoves()

			assert.GreaterOrEqual(t, len(got), tt.expectMin, "Too few moves")
			assert.LessOrEqual(t, len(got), tt.expectMax, "Too many moves")
		})
	}
}

func TestIsInsufficientMaterial(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func(*Game)
		expected bool
	}{
		{
			name: "King vs King",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
			},
			expected: true,
		},
		{
			name: "King + Knight vs King",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("d4", WhiteKnight)
			},
			expected: true,
		},
		{
			name: "King vs King + Knight",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("d4", BlackKnight)
			},
			expected: true,
		},
		{
			name: "King + Bishop vs King",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("c1", WhiteBishop)
			},
			expected: true,
		},
		{
			name: "King + Bishop vs King + Bishop (same color)",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("c1", WhiteBishop) // Dark square (2,7) -> (2+7)%2 = 1
				g.Board.SetPieceAt("f8", BlackBishop) // Dark square (5,0) -> (5+0)%2 = 1
			},
			expected: true,
		},
		{
			name: "King + Bishop vs King + Bishop (different color)",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("c1", WhiteBishop) // Dark square
				g.Board.SetPieceAt("c8", BlackBishop) // Light square
			},
			expected: false,
		},
		{
			name: "King + Pawn vs King",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("e2", WhitePawn)
			},
			expected: false,
		},
		{
			name: "King + Rook vs King",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
				g.Board.SetPieceAt("a1", WhiteRook)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			tt.setupFn(game)
			assert.Equal(t, tt.expected, game.IsInsufficientMaterial())
		})
	}
}

func TestThreefoldRepetition(t *testing.T) {
	game := NewGame()
	game.Board.Clear()
	game.CastlingRights = NoCastling // Ensure castling rights don't change

	// Setup a position where pieces can move back and forth
	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("e8", BlackKing)
	game.Board.SetPieceAt("a1", WhiteRook)
	game.Board.SetPieceAt("a8", BlackRook)

	// Initial position (1st occurrence)
	// We need to manually set the position history because we cleared the board
	// and set pieces manually, bypassing NewGame's initialization
	game.ZobristHash = game.ComputeZobristHash()
	game.PositionHistory = []uint64{game.ZobristHash}

	// Move 1: White Rook a1-b1
	game.ExecuteMove(NewMove(FileA, Rank1, FileB, Rank1))
	// Move 1: Black Rook a8-b8
	game.ExecuteMove(NewMove(FileA, Rank8, FileB, Rank8))

	// Move 2: White Rook b1-a1 (2nd occurrence of initial position)
	game.ExecuteMove(NewMove(FileB, Rank1, FileA, Rank1))
	// Move 2: Black Rook b8-a8
	game.ExecuteMove(NewMove(FileB, Rank8, FileA, Rank8))

	assert.False(t, game.CanClaimDrawByThreefoldRepetition(), "Should not be draw yet (2 occurrences)")

	// Move 3: White Rook a1-b1
	game.ExecuteMove(NewMove(FileA, Rank1, FileB, Rank1))
	// Move 3: Black Rook a8-b8
	game.ExecuteMove(NewMove(FileA, Rank8, FileB, Rank8))

	// Move 4: White Rook b1-a1 (3rd occurrence of initial position)
	game.ExecuteMove(NewMove(FileB, Rank1, FileA, Rank1))
	// Move 4: Black Rook b8-a8
	game.ExecuteMove(NewMove(FileB, Rank8, FileA, Rank8))

	assert.True(t, game.CanClaimDrawByThreefoldRepetition(), "Should be draw (3 occurrences)")
}

func TestThreefoldRepetitionFromStart(t *testing.T) {
	game := NewGame()
	// Start position is 1st occurrence

	// Move 1: White Knight g1-f3
	game.ExecuteMove(NewMove(FileG, Rank1, FileF, Rank3))
	// Move 1: Black Knight g8-f6
	game.ExecuteMove(NewMove(FileG, Rank8, FileF, Rank6))

	// Move 2: White Knight f3-g1 (2nd occurrence of start position)
	game.ExecuteMove(NewMove(FileF, Rank3, FileG, Rank1))
	// Move 2: Black Knight f6-g8
	game.ExecuteMove(NewMove(FileF, Rank6, FileG, Rank8))

	assert.False(t, game.CanClaimDrawByThreefoldRepetition(), "Should not be draw yet (2 occurrences)")

	// Move 3: White Knight g1-f3
	game.ExecuteMove(NewMove(FileG, Rank1, FileF, Rank3))
	// Move 3: Black Knight g8-f6
	game.ExecuteMove(NewMove(FileG, Rank8, FileF, Rank6))

	// Move 4: White Knight f3-g1 (3rd occurrence of start position)
	game.ExecuteMove(NewMove(FileF, Rank3, FileG, Rank1))
	// Move 4: Black Knight f6-g8
	game.ExecuteMove(NewMove(FileF, Rank6, FileG, Rank8))

	assert.True(t, game.CanClaimDrawByThreefoldRepetition(), "Should be draw (3 occurrences)")
}

func TestGetGameStatus(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func(*Game)
		expected GameStatus
	}{
		{
			name:     "Active Game",
			setupFn:  func(g *Game) {},
			expected: StatusActive,
		},
		{
			name: "Checkmate",
			setupFn: func(g *Game) {
				g.Board.Clear()
				// Fool's Mate setup
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("d1", WhitePawn)
				g.Board.SetPieceAt("f1", WhitePawn)
				g.Board.SetPieceAt("e3", BlackQueen)
			},
			expected: StatusCheckmate,
		},
		{
			name: "Stalemate",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("a8", BlackKing)
				g.Board.SetPieceAt("c7", WhiteQueen)
				g.Board.SetPieceAt("h1", WhiteKing)
				g.CurrentTurn = Black
			},
			expected: StatusStalemate,
		},
		{
			name: "Insufficient Material",
			setupFn: func(g *Game) {
				g.Board.Clear()
				g.Board.SetPieceAt("e1", WhiteKing)
				g.Board.SetPieceAt("e8", BlackKing)
			},
			expected: StatusDrawInsufficientMaterial,
		},
		{
			name: "50-Move Rule",
			setupFn: func(g *Game) {
				g.HalfMoveClock = 100
			},
			expected: StatusDraw50Move,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			tt.setupFn(game)
			assert.Equal(t, tt.expected, game.GetGameStatus())
		})
	}
}

func TestGetNoisyMoves(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup:
	// White King at e1
	// White Pawn at e4
	// Black Pawn at d5 (capture target)
	// White Knight at g1 (quiet move available to f3)
	// Black King at e8

	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("e4", WhitePawn)
	game.Board.SetPieceAt("d5", BlackPawn)
	game.Board.SetPieceAt("g1", WhiteKnight)
	game.Board.SetPieceAt("e8", BlackKing)
	game.CurrentTurn = White

	noisyMoves := game.GetNoisyMoves()

	// Expected: e4xd5 only. g1-f3 is quiet. e4-e5 is quiet.
	assert.Len(t, noisyMoves, 1, "Expected exactly 1 noisy move")
	if len(noisyMoves) > 0 {
		assert.Equal(t, FileE, noisyMoves[0].FromCol)
		assert.Equal(t, Rank4, noisyMoves[0].FromRow)
		assert.Equal(t, FileD, noisyMoves[0].ToCol)
		assert.Equal(t, Rank5, noisyMoves[0].ToRow)
	}
}

func TestGetNoisyMovesPromotions(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup: White Pawn at a7
	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("a7", WhitePawn)
	game.Board.SetPieceAt("e8", BlackKing)
	game.CurrentTurn = White

	noisyMoves := game.GetNoisyMoves()

	// Expected: 4 promotion moves (Q, R, B, N)
	assert.Len(t, noisyMoves, 4, "Expected 4 promotion moves")
	for _, m := range noisyMoves {
		assert.NotEqual(t, NoPiece, m.PromotionPiece, "Move should be a promotion")
	}
}

func TestGetNoisyMovesEnPassant(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("e5", WhitePawn)
	game.Board.SetPieceAt("d5", BlackPawn) // Just moved d7-d5
	game.Board.SetPieceAt("e8", BlackKing)

	game.CurrentTurn = White
	game.EnPassantTargetCol = FileD
	game.EnPassantTargetRow = Rank6

	noisyMoves := game.GetNoisyMoves()

	// Expected: exd6 e.p.
	// Note: e5-e6 is quiet, so not included.
	assert.Len(t, noisyMoves, 1, "Expected 1 noisy move (en passant)")
	if len(noisyMoves) > 0 {
		assert.Equal(t, FileE, noisyMoves[0].FromCol)
		assert.Equal(t, FileD, noisyMoves[0].ToCol)
		assert.Equal(t, Rank6, noisyMoves[0].ToRow)
	}
}

func TestGetNoisyMovesPinnedCapture(t *testing.T) {
	game := NewGame()
	game.Board.Clear()

	// Setup:
	// White King at e1
	// White Rook at e2 (pinned by Black Rook at e8)
	// Black Rook at e8
	// Black Knight at a2 (capture target for White Rook, not adjacent to King)
	// White Pawn at d7 (can promote or capture Rook at e8)
	// White Pawn at h2 (can make quiet move)

	game.Board.SetPieceAt("e1", WhiteKing)
	game.Board.SetPieceAt("e2", WhiteRook)
	game.Board.SetPieceAt("e8", BlackRook)
	game.Board.SetPieceAt("a2", BlackKnight)
	game.Board.SetPieceAt("d7", WhitePawn)
	game.Board.SetPieceAt("h2", WhitePawn)

	game.CurrentTurn = White


	noisyMoves := game.GetNoisyMoves()

	// Convert to map for easy lookup
	moveMap := make(map[string]bool)
	for _, m := range noisyMoves {
		moveMap[m.String()] = true
	}

	// 1. Check Rxe8 (Capture of pinner) - Should be present
	assert.True(t, moveMap["e2e8"], "Should find valid capture of pinner (Rxe8)")

	// 2. Check Rxa2 (Capture by pinned piece moving off-line) - Should NOT be present
	assert.False(t, moveMap["e2a2"], "Should not find illegal capture by pinned piece (Rxa2)")

	// 3. Check d7xe8 (Pawn capture promotion) - Should be present
	assert.True(t, moveMap["d7e8q"], "Should find pawn capture promotion (d7xe8=Q)")

	// 4. Check d7-d8 (Pawn push promotion) - Should be present
	assert.True(t, moveMap["d7d8q"], "Should find pawn promotion (d7-d8=Q)")

	// 5. Check h2-h3 (Quiet move) - Should NOT be present
	assert.False(t, moveMap["h2h3"], "Should not find quiet pawn move (h2-h3)")
}
