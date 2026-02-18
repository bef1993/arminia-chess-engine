package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
	"sync/atomic"
)

// quiescence search extends the search at leaf nodes to avoid the horizon effect.
// It only considers "noisy" moves (captures and promotions).
// Returns: score, interrupted
func quiescence(ctx context.Context, game *engine.Game, alpha, beta, ply int, nodes *atomic.Int64, selDepth *int, threadID int) (int, bool) {
	if ply > *selDepth {
		*selDepth = ply
	}

	// Check for timeout every 2048 nodes
	if (nodes.Load() & 2047) == 0 {
		if ctx.Err() != nil {
			return 0, true
		}
	}
	nodes.Add(1)

	// --- Draw Detection ---
	if game.IsDrawByFiftyMoveRule() || game.CanClaimDrawByThreefoldRepetition() || game.IsInsufficientMaterial() {
		return 0, false
	}

	alphaOrig := alpha
	var ttMove engine.Move

	// 1. Transposition Table Lookup
	// We can use any entry from the TT because QS is effectively depth 0,
	// and all TT entries have depth >= 0.
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
		ttMove = entry.BestMove
		score := scoreFromTT(entry.Score, ply)
		if entry.Flag == FlagExact {
			return score, false
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
			return score, false
		}
	}

	standPat := Evaluate(game)
	if standPat >= beta {
		ttScore := scoreToTT(standPat, ply)
		GlobalTT.Store(game.ZobristHash, 0, ttScore, FlagLowerBound, engine.Move{})
		return beta, false
	}
	if alpha < standPat {
		alpha = standPat
	}

	moves := game.GetNoisyMoves()
	orderMoves(game, moves, ttMove, threadID)
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, interrupted := quiescence(ctx, game, -beta, -alpha, ply+1, nodes, selDepth, threadID)
		if interrupted {
			game.UnmakeMove()
			return 0, true
		}
		game.UnmakeMove()
		score = -score

		if score >= beta {
			ttScore := scoreToTT(score, ply)
			GlobalTT.Store(game.ZobristHash, 0, ttScore, FlagLowerBound, move)
			return beta, false
		}
		if score > alpha {
			alpha = score
			bestMove = move
		}
	}

	// 2. Store in Transposition Table
	flag := FlagUpperBound
	if alpha > alphaOrig {
		flag = FlagExact
	}
	ttScore := scoreToTT(alpha, ply)
	GlobalTT.Store(game.ZobristHash, 0, ttScore, flag, bestMove)

	return alpha, false
}
