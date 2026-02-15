package engine

// GenerateAllMoves generates all pseudo-legal moves for the current position
func (g *Game) GenerateAllMoves() []Move {
	var moves []Move
	color := g.CurrentTurn

	// Pawns
	pawns := g.Board.Pieces[color][Pawn]
	for pawns != 0 {
		sq := pawns.PopLSB()
		moves = append(moves, g.generatePawnMoves(sq, color)...)
	}

	// Knights
	knights := g.Board.Pieces[color][Knight]
	for knights != 0 {
		sq := knights.PopLSB()
		moves = append(moves, g.generateKnightMoves(sq, color)...)
	}

	// Bishops
	bishops := g.Board.Pieces[color][Bishop]
	for bishops != 0 {
		sq := bishops.PopLSB()
		moves = append(moves, g.generateBishopMoves(sq, color)...)
	}

	// Rooks
	rooks := g.Board.Pieces[color][Rook]
	for rooks != 0 {
		sq := rooks.PopLSB()
		moves = append(moves, g.generateRookMoves(sq, color)...)
	}

	// Queens
	queens := g.Board.Pieces[color][Queen]
	for queens != 0 {
		sq := queens.PopLSB()
		moves = append(moves, g.generateQueenMoves(sq, color)...)
	}

	// King
	king := g.Board.Pieces[color][King]
	if king != 0 {
		sq := king.PopLSB()
		moves = append(moves, g.generateKingMoves(sq, color)...)
	}

	return moves
}

// GenerateMovesForPiece generates all possible moves for a piece at the given square
// Kept for testing and specific piece logic
func (g *Game) GenerateMovesForPiece(sq int) []Move {
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

	// Helper function to add promotion moves
	addPromotionMoves := func(toSq int) {
		// Pawn can promote to Queen, Rook, Bishop, or Knight
		for _, piece := range []PieceType{Queen, Rook, Bishop, Knight} {
			moves = append(moves, NewPromotionMove(sq, toSq, piece.FromColor(color)))
		}
	}

	if color == White {
		// Single Push
		toSq := sq + 8
		if toSq < 64 && !g.Board.Occupancy[AnyColor].IsSet(toSq) {
			if toSq >= 56 { // Rank 8 (Promotion)
				addPromotionMoves(toSq)
			} else {
				moves = append(moves, NewMove(sq, toSq))
				// Double Push (only if single push was valid and on Rank 2)
				if sq >= 8 && sq <= 15 { // Rank 2
					toSq2 := sq + 16
					if !g.Board.Occupancy[AnyColor].IsSet(toSq2) {
						moves = append(moves, NewMove(sq, toSq2))
					}
				}
			}
		}

		// Captures
		// Capture Left (sq + 7) - Valid if not on File A
		if (sq & 7) != 0 { // sq % 8 != 0
			toSq := sq + 7
			if toSq < 64 {
				if g.Board.Occupancy[Black].IsSet(toSq) {
					if toSq >= 56 {
						addPromotionMoves(toSq)
					} else {
						moves = append(moves, NewMove(sq, toSq))
					}
				} else if toSq == g.EnPassantTarget {
					moves = append(moves, NewMove(sq, toSq))
				}
			}
		}
		// Capture Right (sq + 9) - Valid if not on File H
		if (sq & 7) != 7 { // sq % 8 != 7
			toSq := sq + 9
			if toSq < 64 {
				if g.Board.Occupancy[Black].IsSet(toSq) {
					if toSq >= 56 {
						addPromotionMoves(toSq)
					} else {
						moves = append(moves, NewMove(sq, toSq))
					}
				} else if toSq == g.EnPassantTarget {
					moves = append(moves, NewMove(sq, toSq))
				}
			}
		}

	} else { // Black
		// Single Push
		toSq := sq - 8
		if toSq >= 0 && !g.Board.Occupancy[AnyColor].IsSet(toSq) {
			if toSq <= 7 { // Rank 1 (Promotion)
				addPromotionMoves(toSq)
			} else {
				moves = append(moves, NewMove(sq, toSq))
				// Double Push (only if single push was valid and on Rank 7)
				if sq >= 48 && sq <= 55 { // Rank 7
					toSq2 := sq - 16
					if !g.Board.Occupancy[AnyColor].IsSet(toSq2) {
						moves = append(moves, NewMove(sq, toSq2))
					}
				}
			}
		}

		// Captures
		// Capture Right (sq - 7) - Valid if not on File H
		if (sq & 7) != 7 {
			toSq := sq - 7
			if toSq >= 0 {
				if g.Board.Occupancy[White].IsSet(toSq) {
					if toSq <= 7 {
						addPromotionMoves(toSq)
					} else {
						moves = append(moves, NewMove(sq, toSq))
					}
				} else if toSq == g.EnPassantTarget {
					moves = append(moves, NewMove(sq, toSq))
				}
			}
		}
		// Capture Left (sq - 9) - Valid if not on File A
		if (sq & 7) != 0 {
			toSq := sq - 9
			if toSq >= 0 {
				if g.Board.Occupancy[White].IsSet(toSq) {
					if toSq <= 7 {
						addPromotionMoves(toSq)
					} else {
						moves = append(moves, NewMove(sq, toSq))
					}
				} else if toSq == g.EnPassantTarget {
					moves = append(moves, NewMove(sq, toSq))
				}
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

	attacks := GetBishopAttacks(sq, g.Board.Occupancy[AnyColor])
	validMoves := attacks & ^g.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	return moves
}

func (g *Game) generateRookMoves(sq int, color Color) []Move {
	var moves []Move

	attacks := GetRookAttacks(sq, g.Board.Occupancy[AnyColor])
	validMoves := attacks & ^g.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
	}

	return moves
}

func (g *Game) generateQueenMoves(sq int, color Color) []Move {
	var moves []Move

	attacks := GetQueenAttacks(sq, g.Board.Occupancy[AnyColor])
	validMoves := attacks & ^g.Board.Occupancy[color]

	for validMoves != 0 {
		toSq := validMoves.PopLSB()
		moves = append(moves, NewMove(sq, toSq))
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
