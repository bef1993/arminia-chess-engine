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
// Returns the best move, the score, and the number of nodes visited.
func Search(game *engine.Game, depth int) (engine.Move, int, int) {
	nodes := 0
	// We call negamax directly. It returns the best score and the best move.
	eval, bestMove := negamax(game, depth, -EvalInfinity, EvalInfinity, 0, &nodes)
	return bestMove, eval, nodes
}

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
func negamax(game *engine.Game, depth int, alpha, beta int, ply int, nodes *int) (int, engine.Move) {
	*nodes++
	alphaOrig := alpha
	var ttMove engine.Move

	// 1. Transposition Table Lookup
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
		ttMove = entry.BestMove
		if entry.Depth >= depth {
			score := scoreFromTT(entry.Score, ply)
			if entry.Flag == FlagExact {
				return score, entry.BestMove
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
				return score, entry.BestMove
			}
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
		score, _ := negamax(game, depth-1, -beta, -alpha, ply+1, nodes)
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
