package search

import (
	"arminia-chess-engine/internal/engine"
	"context"
)

// scoreToTT converts a search score to a TT score (independent of ply)
func scoreToTT(score, ply int) int {
	if score > MateBound {
		return score + ply
	}
	if score < -MateBound {
		return score - ply
	}
	return score
}

// scoreFromTT converts a TT score to a search score (relative to ply)
func scoreFromTT(score, ply int) int {
	if score > MateBound {
		return score - ply
	}
	if score < -MateBound {
		return score + ply
	}
	return score
}

// SearchInfo holds progress information sent via channel
type SearchInfo struct {
	Depth    int
	Score    int
	Nodes    int
	BestMove engine.Move
	PV       []engine.Move
}

// SearchOptions holds configuration for the search
type SearchOptions struct {
	MaxDepth int
}

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
// Returns the best move, the score, and the number of nodes visited.
// infoCh is a channel to report search progress (can be nil).
func Search(ctx context.Context, game *engine.Game, options SearchOptions, infoCh chan<- SearchInfo) (engine.Move, int, int) {
	var bestMove engine.Move
	var score int
	totalNodes := 0

	legalMoves := game.GetLegalMoves()

	// Initialize bestMove with a fallback in case the search is interrupted before depth 1 completes.
	// 1. Try to retrieve a move from the Transposition Table (e.g. from previous search)
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok && entry.BestMove != (engine.Move{}) {
		// Validate TT move to ensure it's legal (guards against hash collisions)
		for _, m := range legalMoves {
			if m == entry.BestMove {
				bestMove = entry.BestMove
				break
			}
		}
	}

	// 2. If no TT move, use the first legal move
	// in case the search is interrupted before depth 1 completes.
	if bestMove == (engine.Move{}) {
		if len(legalMoves) > 0 {
			bestMove = legalMoves[0]
		}
	}

	// Iterative Deepening
	for depth := 1; depth <= options.MaxDepth; depth++ {
		// We use a new node counter for each iteration to track work done
		eval, move, interrupted := negamax(ctx, game, depth, -EvalInfinity, EvalInfinity, 0, &totalNodes)

		if interrupted {
			break
		}

		bestMove = move
		score = eval

		if infoCh != nil {
			infoCh <- SearchInfo{
				Depth:    depth,
				Score:    score,
				Nodes:    totalNodes,
				BestMove: bestMove,
				PV:       getPV(game, bestMove, depth),
			}
		}

		if ctx.Err() != nil {
			break
		}
	}
	return bestMove, score, totalNodes
}

// getPV extracts the Principal Variation (best line of play) from the TT
func getPV(game *engine.Game, firstMove engine.Move, depth int) []engine.Move {
	pv := []engine.Move{firstMove}
	movesMade := 0

	// We must execute moves to update the hash and probe the TT for the next move
	// We will undo them all at the end to restore game state
	if !game.ExecuteMove(firstMove) {
		return pv
	}
	movesMade++

	for i := 0; i < depth-1; i++ {
		entry, ok := GlobalTT.Probe(game.ZobristHash)
		if !ok || entry.BestMove == (engine.Move{}) {
			break
		}
		// In a real engine, we should verify legality here
		if !game.ExecuteMove(entry.BestMove) {
			break
		}
		pv = append(pv, entry.BestMove)
		movesMade++
	}

	for i := 0; i < movesMade; i++ {
		game.UnmakeMove()
	}
	return pv
}

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
// Returns: score, bestMove, interrupted
func negamax(ctx context.Context, game *engine.Game, depth int, alpha, beta int, ply int, nodes *int) (int, engine.Move, bool) {
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
		score, interrupted := quiescence(ctx, game, alpha, beta, ply, nodes)
		return score, engine.Move{}, interrupted
	}

	moves := game.GetLegalMoves()

	if len(moves) == 0 {
		if game.Board.IsKingInCheck(game.CurrentTurn) {
			// Checkmate: return a very low score, adjusted by ply to prefer faster mates
			return -EvalMate + ply, engine.Move{}, false
		}
		// Stalemate
		return 0, engine.Move{}, false
	}

	// Move Ordering: Try the move from TT first (Hash Move)
	if ttMove != (engine.Move{}) {
		for i, m := range moves {
			if m == ttMove {
				// Swap the hash move to the front
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	bestScore := -EvalInfinity
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, _, interrupted := negamax(ctx, game, depth-1, -beta, -alpha, ply+1, nodes)
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

// quiescence search extends the search at leaf nodes to avoid the horizon effect.
// It only considers "noisy" moves (captures and promotions).
// Returns: score, interrupted
func quiescence(ctx context.Context, game *engine.Game, alpha, beta, ply int, nodes *int) (int, bool) {
	// Check for timeout every 2048 nodes
	if (*nodes & 2047) == 0 {
		if ctx.Err() != nil {
			return 0, true
		}
	}
	*nodes++
	alphaOrig := alpha

	// 1. Transposition Table Lookup
	// We can use any entry from the TT because QS is effectively depth 0,
	// and all TT entries have depth >= 0.
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
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
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, interrupted := quiescence(ctx, game, -beta, -alpha, ply+1, nodes)
		if interrupted {
			game.UnmakeMove()
			return 0, true
		}
		game.UnmakeMove()

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
