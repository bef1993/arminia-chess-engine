package search

import "arminia-chess-engine/internal/engine"

// HistoryTable stores history heuristic scores
type HistoryTable [2][64][64]int

const HistoryMax = 700000

func (h *HistoryTable) Add(move engine.Move, depth int, player engine.Color) {
	bonus := depth * depth
	h[player][move.From][move.To] += bonus
	if h[player][move.From][move.To] > HistoryMax {
		h.ScaleDown()
	}
}

func (h *HistoryTable) Get(move engine.Move, player engine.Color) int {
	return h[player][move.From][move.To]
}

func (h *HistoryTable) ScaleDown() {
	for c := 0; c < 2; c++ {
		for f := 0; f < 64; f++ {
			for t := 0; t < 64; t++ {
				h[c][f][t] /= 2
			}
		}
	}
}