package search

import "arminia-chess-engine/internal/engine"

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

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
func Search(game *engine.Game, depth int) (engine.Move, int) {
	// We call negamax directly. It returns the best score and the best move.
	eval, bestMove := negamax(game, depth, -EvalInfinity, EvalInfinity, 0)
	return bestMove, eval
}

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
func negamax(game *engine.Game, depth int, alpha, beta int, ply int) (int, engine.Move) {
	alphaOrig := alpha

	// 1. Transposition Table Lookup
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok && entry.Depth >= depth {
		score := scoreFromTT(entry.Score, ply)
		if entry.Flag == FlagExact {
			return score, entry.BestMove
		}
		if entry.Flag == FlagLowerBound {
			if score > alpha {
				alpha = score
			}
		} else if entry.Flag == FlagUpperBound {
			if score < beta {
				beta = score
			}
		}
		if alpha >= beta {
			return score, entry.BestMove
		}
	}

	if depth == 0 {
		return Evaluate(game), engine.Move{}
	}

	moves := game.GetLegalMoves()

	if len(moves) == 0 {
		if game.Board.IsKingInCheck(game.CurrentTurn) {
			// Checkmate: return a very low score, adjusted by ply to prefer faster mates
			return -EvalMate + ply, engine.Move{}
		}
		// Stalemate
		return 0, engine.Move{}
	}

	bestScore := -EvalInfinity
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, _ := negamax(game, depth-1, -beta, -alpha, ply+1)
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

	return bestScore, bestMove
}
