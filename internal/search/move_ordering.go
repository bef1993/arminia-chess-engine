package search

import (
	"arminia-chess-engine/internal/engine"
	"sort"
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
// 1. Hash Move (TT Move): 3,000,000
//   - The best move from a previous search depth. Always searched first.
//
// 2. Captures (MVV-LVA): 1,000,000 + (10 * VictimValue - AttackerValue)
//   - Most Valuable Victim, Least Valuable Aggressor.
//   - Prioritizes capturing high-value pieces with low-value pieces (e.g., PxQ).
//
// 3. Promotions: 1,000,000 + PromotionPieceValue * 10
//   - Queen promotion: 1,009,000.
//
// 4. Quiet Moves: 0
func orderMoves(game *engine.Game, moves []engine.Move, ttMove engine.Move) {
	scores := make([]int, len(moves))

	for i, move := range moves {
		if move == ttMove {
			scores[i] = 3000000 // Highest priority
			continue
		}

		score := 0

		attacker := game.Board.GetPiece(move.From)
		victim := game.Board.GetPiece(move.To)
		if victim != engine.NoPiece {
			// MVV-LVA score: 10 * victim - attacker
			// Offset by 1000000 to prioritize over quiet moves
			attackerValue := attacker.Value()
			// Cap King value for move ordering to ensure MVV (Most Valuable Victim) dominance.
			if attacker.Type() == engine.King {
				attackerValue = engine.QueenValue + 100 // Ensure king is the "least valuable" attacker.
			}
			score = 1000000 + (victim.Value() * 10) - attackerValue
		} else if attacker.Type() == engine.Pawn && move.To == game.EnPassantTarget && (move.From%8) != (move.To%8) {
			// En Passant capture (victim is Pawn)
			score = 1000000 + (engine.PawnValue * 10) - engine.PawnValue
		}

		// Promotions
		if move.PromotionPiece != engine.NoPiece {
			// Prioritize promotions. Queen promotion (900) is valuable.
			// Add to existing score (captures + promotion is very valuable)
			score += 1000000 + move.PromotionPiece.Value()*10
		}

		scores[i] = score
	}

	// Sort moves based on scores (descending)
	sort.Sort(&moveSorter{moves: moves, scores: scores})
}
