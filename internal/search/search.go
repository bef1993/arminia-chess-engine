package search

import "arminia-chess-engine/internal/engine"



// Search finds the best move for the current position.
// This is the entry point for the search algorithm.
func Search(game *engine.Game, depth int) engine.Move {
	// We call negamax directly. It returns the best score and the best move.
	// We discard the score here as we only need the move for the UCI protocol.
	_, bestMove := negamax(game, depth, -EvalInfinity, EvalInfinity, 0)
	return bestMove
}

// negamax implements the Negamax algorithm with alpha-beta pruning.
// ply is the distance from the root of the search tree.
func negamax(game *engine.Game, depth int, alpha, beta int, ply int) (int, engine.Move) {
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

	return bestScore, bestMove
}
