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

// SearchCallback is a function type for reporting search progress
type SearchCallback func(depth, score, nodes int, bestMove engine.Move)

// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
// Returns the best move, the score, and the number of nodes visited.
// onInfo is a callback function to report search progress (can be nil).
func Search(game *engine.Game, maxDepth int, onInfo SearchCallback) (engine.Move, int, int) {
	var bestMove engine.Move
	var score int
	totalNodes := 0

	// Iterative Deepening
	for depth := 1; depth <= maxDepth; depth++ {
		// We use a new node counter for each iteration to track work done
		// Note: In a real time-managed engine, we would check for timeout here
		eval, move := negamax(game, depth, -EvalInfinity, EvalInfinity, 0, &totalNodes)

		bestMove = move
		score = eval

		if onInfo != nil {
			onInfo(depth, score, totalNodes, bestMove)
		}
	}
	return bestMove, score, totalNodes
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
		return quiescence(game, alpha, beta, nodes), engine.Move{}
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

// quiescence search extends the search at leaf nodes to avoid the horizon effect.
// It only considers "noisy" moves (captures and promotions).
func quiescence(game *engine.Game, alpha, beta int, nodes *int) int {
	*nodes++
	alphaOrig := alpha

	// 1. Transposition Table Lookup
	// We can use any entry from the TT because QS is effectively depth 0,
	// and all TT entries have depth >= 0.
	if entry, ok := GlobalTT.Probe(game.ZobristHash); ok {
		score := entry.Score
		if entry.Flag == FlagExact {
			return score
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
			return score
		}
	}

	standPat := Evaluate(game)
	if standPat >= beta {
		GlobalTT.Store(game.ZobristHash, 0, standPat, FlagLowerBound, engine.Move{})
		return beta
	}
	if alpha < standPat {
		alpha = standPat
	}

	moves := game.GetLegalMoves() // TODO implement game.GetLegalCapturesAndPromotions() for efficiency
	var bestMove engine.Move

	for _, move := range moves {
		// Filter for captures and promotions
		isCapture := game.Board.GetPiece(move.ToCol, move.ToRow) != engine.NoPiece

		// Check En Passant (special capture case where target square is empty)
		if !isCapture {
			piece := game.Board.GetPiece(move.FromCol, move.FromRow)
			if piece.Type() == engine.Pawn &&
				move.ToCol == game.EnPassantTargetCol &&
				move.ToRow == game.EnPassantTargetRow {
				isCapture = true
			}
		}

		isPromotion := move.PromotionPiece != engine.NoPiece

		if !isCapture && !isPromotion {
			continue
		}

		game.ExecuteMove(move)
		score := -quiescence(game, -beta, -alpha, nodes)
		game.UnmakeMove()

		if score >= beta {
			GlobalTT.Store(game.ZobristHash, 0, score, FlagLowerBound, move)
			return beta
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
	GlobalTT.Store(game.ZobristHash, 0, alpha, flag, bestMove)

	return alpha
}
