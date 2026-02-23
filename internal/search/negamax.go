package search

import (
	"arminia-chess-engine/internal/engine"
)

// Negamax with Alpha-Beta Pruning
// Current player is always the maximizing player
//
// Alpha-Beta Pruning Optimization:
// Alpha (α) represents the minimum score that the maximizing player is assured of.
// Beta (β) represents the maximum score that the minimizing player is assured of.
//
// The search maintains a window [alpha, beta].
//   - If a move results in a score >= beta, it means this position is "too good" for the current player,
//     implying the opponent would have avoided this branch earlier. This triggers a "Beta Cutoff" (pruning).
//   - If a move results in a score > alpha, it means we found a better move than before. We update alpha.
//
// Parameters:
// - sc: Search context (context, game, stats, rng, killerMoves).
// - depth: Remaining search depth.
// - alpha: The best score the side to move can guarantee so far (Lower Bound).
// - beta: The worst score the opponent can force the side to move to accept (Upper Bound).
// - ply: Current depth in the tree (used for TT score adjustments to prefer faster checkmates).
// - extensions: Count of Check Extensions used in the current line.
// - allowNull: Whether Null Move Pruning is allowed in this node (disabled in child nodes after a null move).
//
// Returns: score, bestMove, interrupted
func negamax(sc *SearchContext, depth int, alpha, beta int, ply int, extensions int, allowNull bool) (int, engine.Move, bool) {
	game := sc.Game
	if ply > *sc.SelDepth {
		*sc.SelDepth = ply
	}

	if shouldStop(sc) {
		return 0, engine.Move{}, true
	}

	// Draw Detection
	if isDraw(game, ply) {
		return 0, engine.Move{}, false
	}

	alphaOrig := alpha
	var ttMove engine.Move

	// Transposition Table Lookup
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

	// Check Extensions
	// If the side to move is in check, we extend the search depth by 1 (maximum 3 extensions)
	inCheck := game.Board.IsKingInCheck(game.CurrentTurn)
	if inCheck && extensions < 3 {
		depth++
		extensions++
	}

	// Null Move Pruning
	if score, cutoff, interrupted := nullMovePruning(sc, depth, beta, ply, extensions, allowNull, inCheck); cutoff || interrupted {
		if interrupted {
			return 0, engine.Move{}, true
		}
		return score, engine.Move{}, false
	}

	// Final depth reached: Call Quiescence Search to resolve captures before final evaluation
	if depth == 0 {
		score, interrupted := quiescence(sc, alpha, beta, ply)
		return score, engine.Move{}, interrupted
	}

	moves := sc.Game.GenerateLegalMoves()

	// No legal moves: Checkmate or Stalemate
	if len(moves) == 0 {
		score := 0
		if inCheck {
			score = -EvalMate + ply // Prefer faster mates by adjusting the score based on ply
		}
		ttScore := scoreToTT(score, ply)
		GlobalTT.Store(game.ZobristHash, depth, ttScore, FlagExact, engine.Move{})
		return score, engine.Move{}, false
	}

	// Move Ordering: Try the move from TT first (Hash Move)
	orderMoves(sc, moves, ttMove, ply)

	bestScore := -EvalInfinity
	var bestMove engine.Move

	for _, move := range moves {
		game.ExecuteMove(move)
		score, _, interrupted := negamax(sc, depth-1, -beta, -alpha, ply+1, extensions, true)
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
			// Found a better move for the current player.
			// Update the lower bound of the search window.
			alpha = score
		}
		if alpha >= beta {
			// Beta Cutoff (Fail-High):
			// The opponent has a better alternative at the previous node that prevents
			// us from reaching this position. We stop searching this branch.
			updateKillerMoves(sc, game, move, ply)
			break
		}
	}

	// Store result in Transposition Table
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

// updateKillerMoves checks if a beta-cutoff move is quiet and stores it as a killer move.
func updateKillerMoves(sc *SearchContext, game *engine.Game, move engine.Move, ply int) {
	// We need to check if it was a capture.
	// Since we already unmade the move, we check the board state (which is restored).
	target := game.Board.GetPiece(move.To)
	isCapture := target != engine.NoPiece
	if !isCapture {
		// Check En Passant (Pawn move to EP target)
		piece := game.Board.GetPiece(move.From)
		if piece.Type() == engine.Pawn && move.To == game.EnPassantTarget && (move.To%8 != move.From%8) {
			isCapture = true
		}
	}

	if !isCapture {
		storeKiller(sc, ply, move)
	}
}

// shouldStop checks if the search should be interrupted due to timeout or cancellation.
func shouldStop(sc *SearchContext) bool {
	// Check for timeout every 2048 nodes
	if (sc.Nodes.Load() & 2047) == 0 {
		if sc.Ctx.Err() != nil {
			return true
		}
	}
	sc.Nodes.Add(1)
	return false
}

// isDraw checks for immediate draw conditions (repetition, 50-move, insufficient material).
func isDraw(game *engine.Game, ply int) bool {
	// We do not want to store repetition draws in the TT because the draw depends on how the position was reached
	// Since we check this before probing the TT, we can avoid storing these positions in the TT and just return a score of 0 immediately.
	// If we are at the root node, we want to search the position anyway to report stats, so we only return 0 for draw positions at ply > 0.
	// Already consider 2 repetitions a draw
	return ply > 0 && (game.IsDrawByFiftyMoveRule() || game.GetRepetitionCount() >= 2 || game.IsInsufficientMaterial())
}

// nullMovePruning attempts to prune the search early by making a "null move"
// Returns score, cutoff (true if pruned), interrupted.
func nullMovePruning(sc *SearchContext, depth, beta, ply, extensions int, allowNull, inCheck bool) (int, bool, bool) {
	game := sc.Game
	// Conditions:
	// 1. Not in check (null move while in check is illegal).
	// 2. Depth is sufficient (>= 3).
	// 3. Not at root (ply > 0).
	// 4. Side to move has non-pawn material (avoid Zugzwang).
	// 5. Null move is allowed (allowNull flag).
	if allowNull && depth >= 3 && !inCheck && ply > 0 && game.Board.HasNonPawnMaterial(game.CurrentTurn) {
		game.ExecuteNullMove()
		reduction := 2
		if depth > 6 {
			reduction = 3
		}
		score, _, interrupted := negamax(sc, depth-1-reduction, -beta, -beta+1, ply+1, extensions, false)
		game.UnmakeMove()
		score = -score
		return score, score >= beta, interrupted
	}
	return 0, false, false
}
