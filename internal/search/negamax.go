package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"sync/atomic"
)

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
// Returns: score, bestMove, interrupted
func negamax(ctx context.Context, game *engine.Game, depth int, alpha, beta int, ply int, nodes *atomic.Int64, selDepth *int) (int, engine.Move, bool) {
	if ply > *selDepth {
		*selDepth = ply
	}

	// Check for timeout every 2048 nodes
	if (nodes.Load() & 2047) == 0 {
		if ctx.Err() != nil {
			return 0, engine.Move{}, true
		}
	}
	nodes.Add(1)

	// --- Draw Detection ---
	// We check for draws before probing the TT. A draw is a draw regardless of whose turn it is.
	// We check at ply > 0 because we want to search for a win from the root even if it's a 2-fold repetition.
	if ply > 0 && (game.IsDrawByFiftyMoveRule() || game.CanClaimDrawByThreefoldRepetition() || game.IsInsufficientMaterial()) {
		// Store draw score in TT
		GlobalTT.Store(game.ZobristHash, depth, 0, FlagExact, engine.Move{})
		return 0, engine.Move{}, false
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
		score, interrupted := quiescence(ctx, game, alpha, beta, ply, nodes, selDepth)
		return score, engine.Move{}, interrupted
	}

	moves := game.GenerateLegalMoves()

	if len(moves) == 0 {
		score := 0
		if game.Board.IsKingInCheck(game.CurrentTurn) {
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
	orderMoves(game, moves, ttMove)

	bestScore := -EvalInfinity
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, _, interrupted := negamax(ctx, game, depth-1, -beta, -alpha, ply+1, nodes, selDepth)
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
