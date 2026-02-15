package engine

// Generate all legal moves for the piece on given square
func (g *Game) generateMovesForPiece(sq int) []Move {
	piece := g.Board.GetPiece(sq)
	if piece == NoPiece {
		return []Move{}
	}

	var moves []Move

	switch piece.Type() {
	case Pawn:
		moves = g.generatePawnMoves(sq, piece.Color())
	case Knight:
		moves = g.generateKnightMoves(sq, piece.Color())
	case Bishop:
		moves = g.generateBishopMoves(sq, piece.Color())
	case Rook:
		moves = g.generateRookMoves(sq, piece.Color())
	case Queen:
		moves = g.generateQueenMoves(sq, piece.Color())
	case King:
		moves = g.generateKingMoves(sq, piece.Color())
	case NoType:
	}

	return moves
}

func (g *Game) generatePawnMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)

	direction := 1
	startRank := int(Rank2)
	promotionRank := int(Rank8)

	if color == Black {
		direction = -1
		startRank = int(Rank7)
		promotionRank = int(Rank1)
	}

	// Helper function to add promotion moves
	addPromotionMoves := func(toSq int) {
		// Pawn can promote to Queen, Rook, Bishop, or Knight
		for _, piece := range []PieceType{Queen, Rook, Bishop, Knight} {
			moves = append(moves, NewPromotionMove(sq, toSq, piece.FromColor(color)))
		}
	}

	// Move forward one square
	newRank := rank + direction
	if IsOnBoard2D(file, newRank) {
		toSq := GetSq(file, newRank)
		if g.Board.IsEmpty(toSq) {
			if newRank == promotionRank {
				addPromotionMoves(toSq)
			} else {
				moves = append(moves, NewMove(sq, toSq))
			}

			// Move forward two squares from starting position
			if rank == startRank {
				newRank2 := rank + 2*direction
				toSq2 := GetSq(file, newRank2)
				if g.Board.IsEmpty(toSq2) {
					moves = append(moves, NewMove(sq, toSq2))
				}
			}
		}
	}

	// Capture diagonally
	for dfile := -1; dfile <= 1; dfile += 2 {
		newFile := file + dfile
		newRank := rank + direction
		if IsOnBoard2D(newFile, newRank) {
			toSq := GetSq(newFile, newRank)
			// Regular capture
			if !g.Board.IsEmpty(toSq) && !g.Board.IsOccupiedByColor(toSq, color) {
				if newRank == promotionRank {
					addPromotionMoves(toSq)
				} else {
					moves = append(moves, NewMove(sq, toSq))
				}
			}

			// En passant capture (never a promotion)
			if toSq == g.EnPassantTarget {
				moves = append(moves, NewMove(sq, toSq))
			}
		}
	}

	return moves
}

func (g *Game) addCastlingMoves(sq int, color Color, moves *[]Move) {
	if g.CastlingRights == NoCastling {
		return
	}

	// Determine opponent color for attack checks
	opponent := Black
	if color == Black {
		opponent = White
	}

	// Cannot castle if King is currently in check
	if g.Board.IsSquareAttackedByColor(sq, opponent) {
		return
	}

	// White Castling
	if color == White && sq == E1 {
		// Kingside (e1 -> g1)
		if g.CastlingRights&WhiteKingside != 0 {
			// Check path is empty (f1, g1)
			if g.Board.IsEmpty(F1) && g.Board.IsEmpty(G1) {
				// Check path is not attacked (f1, g1)
				// Note: We already checked e1 (current pos) above
				if !g.Board.IsSquareAttackedByColor(F1, opponent) &&
					!g.Board.IsSquareAttackedByColor(G1, opponent) {
					*moves = append(*moves, NewMove(E1, G1))
				}
			}
		}
		// Queenside (e1 -> c1)
		if g.CastlingRights&WhiteQueenside != 0 {
			// Check path is empty (d1, c1, b1)
			if g.Board.IsEmpty(D1) && g.Board.IsEmpty(C1) && g.Board.IsEmpty(B1) {
				// Check path is not attacked (d1, c1)
				// Note: b1 does not need to be safe, only empty
				if !g.Board.IsSquareAttackedByColor(D1, opponent) &&
					!g.Board.IsSquareAttackedByColor(C1, opponent) {
					*moves = append(*moves, NewMove(E1, C1))
				}
			}
		}
	}

	// Black Castling
	if color == Black && sq == E8 {
		// Kingside (e8 -> g8)
		if g.CastlingRights&BlackKingside != 0 {
			if g.Board.IsEmpty(F8) && g.Board.IsEmpty(G8) {
				if !g.Board.IsSquareAttackedByColor(F8, opponent) &&
					!g.Board.IsSquareAttackedByColor(G8, opponent) {
					*moves = append(*moves, NewMove(E8, G8))
				}
			}
		}
		// Queenside (e8 -> c8)
		if g.CastlingRights&BlackQueenside != 0 {
			if g.Board.IsEmpty(D8) && g.Board.IsEmpty(C8) && g.Board.IsEmpty(B8) {
				if !g.Board.IsSquareAttackedByColor(D8, opponent) &&
					!g.Board.IsSquareAttackedByColor(C8, opponent) {
					*moves = append(*moves, NewMove(E8, C8))
				}
			}
		}
	}
}

func (g *Game) generateKnightMoves(sq int, color Color) []Move {
	var moves []Move

	// Use pre-calculated attacks
	attacks := KnightAttacks[sq]

	// Mask out own pieces
	validMoves := attacks & ^g.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	return moves
}

func (g *Game) generateBishopMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if g.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !g.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (g *Game) generateRookMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if g.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !g.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (g *Game) generateQueenMoves(sq int, color Color) []Move {
	var moves []Move
	rank := GetRank(sq)
	file := GetFile(sq)
	directions := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + i*dir[0]
			newFile := file + i*dir[1]

			if !IsOnBoard2D(newFile, newRank) {
				break
			}

			toSq := GetSq(newFile, newRank)
			if g.Board.IsEmpty(toSq) {
				moves = append(moves, NewMove(sq, toSq))
			} else if !g.Board.IsOccupiedByColor(toSq, color) {
				moves = append(moves, NewMove(sq, toSq))
				break
			} else {
				break
			}
		}
	}

	return moves
}

func (g *Game) generateKingMoves(sq int, color Color) []Move {
	var moves []Move

	// Use pre-calculated attacks
	attacks := KingAttacks[sq]

	// Mask out own pieces
	validMoves := attacks & ^g.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	// Add castling moves
	g.addCastlingMoves(sq, color, &moves)

	return moves
}
