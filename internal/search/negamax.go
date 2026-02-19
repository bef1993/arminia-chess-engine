package search

import (
	"arminia-chess-engine/internal/engine"
)

// negamax implements the Negamax algorithm with alpha-beta pruning.
// sc holds the search context (game state, cancellation, stats, RNG).
// depth is the remaining search depth.
// alpha and beta are the current bounds for pruning.
// ply is the current depth in the tree (used for TT score adjustments to prefer faster checkmates).
// Returns: score, bestMove, interrupted
func negamax(sc *SearchContext, depth int, alpha, beta int, ply int) (int, engine.Move, bool) {
	game := sc.Game
	if ply > *sc.SelDepth {
		*sc.SelDepth = ply
	}

	// Check for timeout every 2048 nodes
	if (sc.Nodes.Load() & 2047) == 0 {
		if sc.Ctx.Err() != nil {
			return 0, engine.Move{}, true
		}
	}
	sc.Nodes.Add(1)

	// --- Draw Detection ---
	// We do not want to store repetion draws in the TT because the draw depends on how the position was reached
	// Since we check this before probing the TT, we can avoid storing these positions in the TT and just return a score of 0 immediately.
	// If we are at the root node, we want to search the position anyway to report stats, so we only return 0 for draw positions at ply > 0.
	if ply > 0 && (game.IsDrawByFiftyMoveRule() || game.CanClaimDrawByThreefoldRepetition() || game.IsInsufficientMaterial()) {
		return 0, engine.Move{}, false
	}

	// Check Extensions
	// If the side to move is in check, we extend the search depth by 1 to find a defense or mate.
	inCheck := game.Board.IsKingInCheck(game.CurrentTurn)
	if inCheck {
		depth++
	}

	alphaOrig := alpha
	var ttMove engine.Move

	// 1. Transposition Table Lookup
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
		ttMove = entry.BestMove
		// Don't cut off at the root (ply 0) to ensure we search and report stats
		if ply > 0 && entry.Depth >= depth {
			score := scoreFromTT(entry.Score, ply)
			if entry.Flag == FlagExact {
				return score, entry.BestMove, false
			}
			switch entry.Flag {
			case FlagLowerBound:
				if score > alpha {
					alpha = score
				}
			case FlagUpperBound:
				if score < beta {
					beta = score
				}
			}
			if alpha >= beta {
				return score, entry.BestMove, false
			}
		}
	}

	if depth == 0 {
		score, interrupted := quiescence(sc, alpha, beta, ply)
		return score, engine.Move{}, interrupted
	}

	moves := sc.Game.GenerateLegalMoves()

	if len(moves) == 0 {
		score := 0
		if inCheck {
			// Checkmate: return a very low score, adjusted by ply to prefer faster mates
			score = -EvalMate + ply
		}
		// Stalemate score is 0

		// Store in TT (Exact score, no move)
		ttScore := scoreToTT(score, ply)
		GlobalTT.Store(game.ZobristHash, depth, ttScore, FlagExact, engine.Move{})

		return score, engine.Move{}, false
	}

	// Move Ordering: Try the move from TT first (Hash Move)
	// TODO: Improve move ordering for quiet moves (e.g., Killer Moves, History Heuristic)
	orderMoves(sc, moves, ttMove)

	bestScore := -EvalInfinity
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, _, interrupted := negamax(sc, depth-1, -beta, -alpha, ply+1)
		if interrupted {
			game.UnmakeMove()
			return 0, engine.Move{}, true
		}
		score = -score
		game.UnmakeMove()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break // Beta cutoff
		}
	}

	// 2. Store in Transposition Table
	var flag TTFlag
	if bestScore <= alphaOrig {
		flag = FlagUpperBound
	} else if bestScore >= beta {
		flag = FlagLowerBound
	} else {
		flag = FlagExact
	}

	ttScore := scoreToTT(bestScore, ply)
	GlobalTT.Store(game.ZobristHash, depth, ttScore, flag, bestMove)

	return bestScore, bestMove, false
}
