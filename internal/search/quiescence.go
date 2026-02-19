package search

import (
	"arminia-chess-engine/internal/engine"
)

// quiescence search extends the search at leaf nodes to avoid the horizon effect.
// It only considers "noisy" moves (captures and promotions).
// Returns: score, interrupted
func quiescence(sc *SearchContext, alpha, beta, ply int) (int, bool) {
	game := sc.Game
	if ply > *sc.SelDepth {
		*sc.SelDepth = ply
	}

	// Check for timeout every 2048 nodes
	if (sc.Nodes.Load() & 2047) == 0 {
		if sc.Ctx.Err() != nil {
			return 0, true
		}
	}
	sc.Nodes.Add(1)

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
	orderMoves(sc, moves, ttMove)
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, interrupted := quiescence(sc, -beta, -alpha, ply+1)
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
