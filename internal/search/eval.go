package search

import "arminia-chess-engine/internal/engine"

const (
	// Infinity is a score that is higher than any possible evaluation.
	// Used for alpha-beta pruning.
	EvalInfinity = 30000

	// Mate is a score indicating a checkmate. It's slightly less than infinity
	// to allow for distinguishing between mates at different depths.
	EvalMate = 29000

	// MateBound is the threshold for considering a score a mate score
	MateBound = EvalMate - 1000
)

// Piece values for standard material counting and move ordering
const (
	PawnValue   = 100
	KnightValue = 320
	BishopValue = 330
	RookValue   = 500
	QueenValue  = 900
	KingValue   = 20000 // Represents "Infinity" (Checkmate)
)

type GamePhase int

const (
	Opening GamePhase = iota
	Middlegame
	Endgame
)

const (
	KnightWeight = 1
	BishopWeight = 1
	RookWeight   = 2
	QueenWeight  = 4
)

// pieceSquareTables[phase][pieceType] gives the piece-square table for a given piece type and game phase.
type PieceSquareTable [64]int

var pieceSquareTables [3][6]PieceSquareTable
var adjacentFilesBB [8]engine.Bitboard
var whitePassedMask[64] engine.Bitboard
var blackPassedMask[64] engine.Bitboard

func init() {
	initPieceSquareTables()
	initAdjacentFiles()
	initPassedPawnMasks()
}

func initPieceSquareTables() {
	pieceSquareTables[Opening][engine.Pawn] = pawnPST
	pieceSquareTables[Middlegame][engine.Pawn] = pawnPST // TODO: Add middlegame PST
	pieceSquareTables[Endgame][engine.Pawn] = pawnPSTEndgame

	pieceSquareTables[Opening][engine.Knight] = knightPST
	pieceSquareTables[Middlegame][engine.Knight] = knightPST
	pieceSquareTables[Endgame][engine.Knight] = knightPST

	pieceSquareTables[Opening][engine.Bishop] = bishopPST
	pieceSquareTables[Middlegame][engine.Bishop] = bishopPST
	pieceSquareTables[Endgame][engine.Bishop] = bishopPST

	pieceSquareTables[Opening][engine.Rook] = rookPST
	pieceSquareTables[Middlegame][engine.Rook] = rookPST
	pieceSquareTables[Endgame][engine.Rook] = rookPST

	pieceSquareTables[Opening][engine.Queen] = queenPST
	pieceSquareTables[Middlegame][engine.Queen] = queenPST
	pieceSquareTables[Endgame][engine.Queen] = queenPST

	pieceSquareTables[Opening][engine.King] = kingPST // TODO: Add opening PST
	pieceSquareTables[Middlegame][engine.King] = kingPST
	pieceSquareTables[Endgame][engine.King] = kingPSTEndgame
}

func initAdjacentFiles() {
	for i := 0; i < 8; i++ {
		if i > 0 {
			adjacentFilesBB[i] |= engine.FileA_BB << (i - 1)
		}
		if i < 7 {
			adjacentFilesBB[i] |= engine.FileA_BB << (i + 1)
		}
	}
}

func initPassedPawnMasks() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8

		// White
		bb := adjacentFilesBB[file] | (engine.FileA_BB << file)
		for r := 0; r <= sq/8; r++ {
			bb &= ^(engine.Rank1_BB << (r * 8))
		}
		whitePassedMask[sq] = bb

		// Black
		bb = adjacentFilesBB[file] | (engine.FileA_BB << file)
		for r := 7; r >= sq/8; r-- {
			bb &= ^(engine.Rank1_BB << (r * 8))
		}
		blackPassedMask[sq] = bb
	}
}

// Piece-Square Tables (PST)
// The PST are defined in human readable format (rank 1 at the bottom) and are indexed by square.
// This means for White we mirror the square (sq ^ 56) to flip the rank.

var pawnPST = PieceSquareTable{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightPST = PieceSquareTable{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopPST = PieceSquareTable{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookPST = PieceSquareTable{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenPST = PieceSquareTable{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingPST = PieceSquareTable{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

var kingPSTEndgame = PieceSquareTable{
	-50, -30, -10, -10, -10, -10, -30, -50,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-10, 20, 40, 50, 50, 40, 20, -10,
	-10, 30, 50, 60, 60, 50, 30, -10,
	-10, 30, 50, 60, 60, 50, 30, -10,
	-10, 20, 40, 50, 50, 40, 20, -10,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-50, -30, -10, -10, -10, -10, -30, -50,
}

var pawnPSTEndgame = PieceSquareTable{
	0, 0, 0, 0, 0, 0, 0, 0,
	150, 150, 150, 150, 150, 150, 150, 150,
	100, 100, 100, 100, 100, 100, 100, 100,
	50, 50, 60, 70, 70, 60, 50, 50,
	20, 20, 30, 40, 40, 30, 20, 20,
	10, 10, 10, 20, 20, 10, 10, 10,
	5, 5, 5, 5, 5, 5, 5, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// Evaluate calculates the score of the current board position from the perspective
// of the current player. A positive score means the current player has an advantage.
func Evaluate(game *engine.Game) int {
	score := evaluatePosition(game.Board)

	// Convert the absolute score to a perspective-based score.
	if game.CurrentTurn == engine.White {
		return score
	}
	return -score
}

func determineGamePhase(board *engine.Board) GamePhase {
	weights := 0
	weights += KnightWeight * (board.Pieces[engine.White][engine.Knight].Count() + board.Pieces[engine.Black][engine.Knight].Count())
	weights += BishopWeight * (board.Pieces[engine.White][engine.Bishop].Count() + board.Pieces[engine.Black][engine.Bishop].Count())
	weights += RookWeight * (board.Pieces[engine.White][engine.Rook].Count() + board.Pieces[engine.Black][engine.Rook].Count())
	weights += QueenWeight * (board.Pieces[engine.White][engine.Queen].Count() + board.Pieces[engine.Black][engine.Queen].Count())

	// Max weight is 24
	if weights > 20 {
		return Opening
	} else if weights > 10 {
		return Middlegame
	}
	return Endgame
}

// evaluatePosition calculates the score based on material and piece-square tables.
// Positive values indicate an advantage for White,
// negative values indicate an advantage for Black.
func evaluatePosition(board engine.Board) int {
	score := 0

	phase := determineGamePhase(&board)

	// Evaluate pieces for White
	for pieceType := engine.Pawn; pieceType <= engine.King; pieceType++ {
		bb := board.Pieces[engine.White][pieceType]
		for bb != 0 {
			sq := bb.PopLSB()
			score += evaluatePiece(sq, pieceType.White(), phase)
		}
	}

	// Evaluate pieces for Black
	for pieceType := engine.Pawn; pieceType <= engine.King; pieceType++ {
		bb := board.Pieces[engine.Black][pieceType]
		for bb != 0 {
			sq := bb.PopLSB()
			score -= evaluatePiece(sq, pieceType.Black(), phase)
		}
	}

	score += evaluatePawnStructure(&board)

	return score
}

const (
	coveredPawnBonus    = 10
	doubledPawnPenalty  = -20
	isolatedPawnPenalty = -20
)

var passedPawnBonus = [8]int{0, 5, 10, 20, 35, 60, 100, 0}

// evaluatePawnStructure evaluates the pawn structure of the position, including:
// - Covered pawns: Pawns that are protected by another pawn.
// - Doubled pawns: Multiple pawns on the same file.
// - Isolated pawns: Pawns with no friendly pawns on adjacent files.
// - Passed pawns: Pawns with no enemy pawns in front of them on the same or adjacent files.
// Positive scores indicate an advantage for White, negative scores indicate an advantage for Black.
func evaluatePawnStructure(board *engine.Board) int {
	score := 0
	whitePawns := board.Pieces[engine.White][engine.Pawn]
	blackPawns := board.Pieces[engine.Black][engine.Pawn]

	// Covered pawns and passed pawns
	whitePawnsCopy := whitePawns
	for whitePawnsCopy != 0 {
		sq := whitePawnsCopy.PopLSB()
		// Covered pawns
		protectedPawns := engine.PawnAttacks[engine.White][sq] & whitePawns
		score += protectedPawns.Count() * coveredPawnBonus

		// Passed pawns
		if (whitePassedMask[sq] & blackPawns) == 0 {
			score += passedPawnBonus[engine.GetRank(sq)]
		}
	}
	blackPawnsCopy := blackPawns
	for blackPawnsCopy != 0 {
		sq := blackPawnsCopy.PopLSB()
		// Covered pawns
		protectedPawns := engine.PawnAttacks[engine.Black][sq] & blackPawns
		score -= protectedPawns.Count() * coveredPawnBonus

		// Passed pawns
		if (blackPassedMask[sq] & whitePawns) == 0 {
			score -= passedPawnBonus[7-engine.GetRank(sq)]
		}
	}

	// Doubled and isolated pawns
	for file := 0; file < 8; file++ {
		fileBB := engine.FileA_BB << file
		whitePawnsOnFile := whitePawns & fileBB
		blackPawnsOnFile := blackPawns & fileBB
		whitePawnsOnFileCount := whitePawnsOnFile.Count()
		blackPawnsOnFileCount := blackPawnsOnFile.Count()

		// Doubled pawns
		if whitePawnsOnFileCount > 1 {
			score += (whitePawnsOnFileCount - 1) * doubledPawnPenalty
		}
		if blackPawnsOnFileCount > 1 {
			score -= (blackPawnsOnFileCount - 1) * doubledPawnPenalty
		}

		// Isolated pawns
		if whitePawnsOnFileCount > 0 {
			if (whitePawns & adjacentFilesBB[file]) == 0 {
				score += isolatedPawnPenalty
			}
		}
		if blackPawnsOnFileCount > 0 {
			if (blackPawns & adjacentFilesBB[file]) == 0 {
				score -= isolatedPawnPenalty
			}
		}
	}

	return score
}

func evaluatePiece(sq int, piece engine.Piece, phase GamePhase) int {
	if piece.Color() == engine.White {
		sq ^= 56 // Mirror square for white
	}

	score := pieceSquareTables[phase][piece.Type()][sq] + pieceValue(piece.Type())
	return score
}

func pieceValue(pieceType engine.PieceType) int {
	switch pieceType {
	case engine.Pawn:
		return PawnValue
	case engine.Knight:
		return KnightValue
	case engine.Bishop:
		return BishopValue
	case engine.Rook:
		return RookValue
	case engine.Queen:
		return QueenValue
	case engine.King:
		return KingValue
	default:
		return 0
	}
}
