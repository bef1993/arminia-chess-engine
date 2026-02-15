package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
)

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
// Returns: score, bestMove, interrupted
func negamax(ctx context.Context, game *engine.Game, depth int, alpha, beta int, ply int, nodes *int, selDepth *int) (int, engine.Move, bool) {
	if ply > *selDepth {
		*selDepth = ply
	}

	// Check for timeout every 2048 nodes
	if (*nodes & 2047) == 0 {
		if ctx.Err() != nil {
			return 0, engine.Move{}, true
		}
	}
	*nodes++

	alphaOrig := alpha
	var ttMove engine.Move

	// 1. Transposition Table Lookup
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
		ttMove = entry.BestMove
		if entry.Depth >= depth {
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
		if game.Board.IsKingInCheck(game.CurrentTurn) {
			// Checkmate: return a very low score, adjusted by ply to prefer faster mates
			return -EvalMate + ply, engine.Move{}, false
		}
		// Stalemate
		return 0, engine.Move{}, false
	}

	// Move Ordering: Try the move from TT first (Hash Move)
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
