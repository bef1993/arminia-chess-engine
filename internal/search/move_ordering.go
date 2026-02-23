package search

import (
	"arminia-chess-engine/internal/engine"
	"sort"
)

// Move Ordering Score Constants
const (
	ScoreHashMove      = 3000000
	ScoreCaptureBase   = 1000000
	ScorePromotionBase = 900000
	ScoreKiller1       = 800000
	ScoreKiller2       = 700000
)

type moveSorter struct {
	moves  []engine.Move
	scores []int
}

func (s *moveSorter) Len() int { return len(s.moves) }
func (s *moveSorter) Swap(i, j int) {
	s.moves[i], s.moves[j] = s.moves[j], s.moves[i]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}
func (s *moveSorter) Less(i, j int) bool { return s.scores[i] > s.scores[j] }

// orderMoves sorts moves based on heuristics to improve search efficiency (Alpha-Beta pruning).
//
// Scoring Hierarchy:
// 1. Hash Move (TT Move): ScoreHashMove (3,000,000)
//   - The best move from a previous search depth. Always searched first.
//
// 2. Capture Promotions: ScoreCaptureBase + ScorePromotionBase + MVV-LVA + PromotionValue
//   - Extremely valuable moves.
//
// 3. Captures (MVV-LVA): ScoreCaptureBase (1,000,000) + (10 * VictimValue - AttackerValue)
//   - Most Valuable Victim, Least Valuable Aggressor.
//   - Prioritizes capturing high-value pieces with low-value pieces (e.g., PxQ).
//
// 4. Quiet Promotions: ScorePromotionBase (900,000) + PromotionPieceValue * 10
//   - Queen promotion: ~909,000.
//
// 5. Killer Moves: ScoreKiller1 (800,000) and ScoreKiller2 (700,000)
//   - Quiet moves that caused a beta-cutoff at this ply in other branches.
//
// 6. History Heuristic: 0 to 700,000
//   - Quiet moves that have historically been good. Capped at ScoreKiller2.
//
// 7. Other Quiet Moves: 0 (or small random value if randomization is enabled)
func orderMoves(sc *SearchContext, moves []engine.Move, ttMove engine.Move, ply int) {
	scores := make([]int, len(moves))
	game := sc.Game

	for i, move := range moves {
		if move == ttMove {
			scores[i] = ScoreHashMove // Highest priority
			continue
		}

		score := 0

		attacker := game.Board.GetPiece(move.From)
		victim := game.Board.GetPiece(move.To)
		if victim != engine.NoPiece {
			// MVV-LVA score: 10 * victim - attacker
			// Offset by 1000000 to prioritize over quiet moves
			attackerValue := pieceValue(attacker.Type())
			// Cap King value for move ordering to ensure MVV (Most Valuable Victim) dominance.
			if attacker.Type() == engine.King {
				attackerValue = QueenValue + 100 // Ensure king is the "least valuable" attacker.
			}
			score = ScoreCaptureBase + (pieceValue(victim.Type()) * 10) - attackerValue
		} else if attacker.Type() == engine.Pawn && move.To == game.EnPassantTarget && (move.From%8) != (move.To%8) {
			// En Passant capture (victim is Pawn)
			score = ScoreCaptureBase + (PawnValue * 10) - PawnValue
		}

		// Promotions
		if move.PromotionPiece != engine.NoPiece {
			// Prioritize promotions. Queen promotion (900) is valuable.
			// Add to existing score (captures + promotion is very valuable)
			score += ScorePromotionBase + pieceValue(move.PromotionPiece.Type())*10
		}

		// Quiet Moves
		if score == 0 {
			// 1. Killer Moves
			if ply < MaxPly {
				if sc.KillerMoves[ply][0] == move {
					score = ScoreKiller1
				} else if sc.KillerMoves[ply][1] == move {
					score = ScoreKiller2
				}
			}

			// 2. History Heuristic
			if score == 0 {
				histScore := sc.History.Get(move, game.CurrentTurn)
				if histScore > HistoryMax {
					histScore = HistoryMax // Capped at HistoryMax (700,000) which is <= ScoreKiller2
				}
				score = histScore
			}
		}

		// Lazy SMP Randomness / Divergence
		// If a random generator is provided (helper threads), add small noise to ALL moves
		// to encourage exploring different branches, even among captures.
		if sc.Rand != nil {
			score += sc.Rand.Intn(10) // Small noise 0-9
		}

		scores[i] = score
	}

	// Sort moves based on scores (descending)
	sort.Sort(&moveSorter{moves: moves, scores: scores})
}

// storeKiller adds a move to the killer moves table for the given ply
func storeKiller(sc *SearchContext, ply int, move engine.Move) {
	if ply >= MaxPly || sc.KillerMoves[ply][0] == move {
		return
	}
	// Shift: Old 1st killer becomes 2nd killer, new move becomes 1st killer
	sc.KillerMoves[ply][1] = sc.KillerMoves[ply][0]
	sc.KillerMoves[ply][0] = move
}
