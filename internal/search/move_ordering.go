package search

import (
	"arminia-chess-engine/internal/engine"
	"sort"
)

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
	scores := make(map[string]int)

	for _, move := range moves {
		moveKey := move.String()
		if move == ttMove {
			scores[moveKey] = 3000000 // Highest priority
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

		scores[moveKey] = score
	}

	// Sort moves based on scores (descending)
	sort.Slice(moves, func(i, j int) bool {
		return scores[moves[i].String()] > scores[moves[j].String()]
	})
}
