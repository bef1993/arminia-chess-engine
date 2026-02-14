package engine

import "math/bits"

type Bitboard uint64

const (
    A1 = iota; B1; C1; D1; E1; F1; G1; H1
    A2; B2; C2; D2; E2; F2; G2; H2
	A3; B3; C3; D3; E3; F3; G3; H3
	A4; B4; C4; D4; E4; F4; G4; H4
	A5; B5; C5; D5; E5; F5; G5; H5
	A6; B6; C6; D6; E6; F6; G6; H6
	A7; B7; C7; D7; E7; F7; G7; H7
	A8; B8; C8; D8; E8; F8; G8; H8
    
)

// Basic Operations
func (b *Bitboard) Set(sq int)      { *b |= (1 << sq) }
func (b *Bitboard) Clear(sq int)    { *b &= ^(1 << sq) }
func (b Bitboard) IsSet(sq int) bool { return (b & (1 << sq)) != 0 }

// Population Count (number of set bits)
func (b Bitboard) Count() int { return bits.OnesCount64(uint64(b)) }

// Get and Clear Least Significant Bit (used for iteration)
func (b *Bitboard) PopLSB() int {
    lsb := bits.TrailingZeros64(uint64(*b))
    *b &= *b - 1 // Clear the LSB
    return lsb
}

// Pre-calculated attack tables
var (
	KnightAttacks [64]Bitboard
	KingAttacks   [64]Bitboard
)

func init() {
	initKnightAndKingAttacks()
}

func initKnightAndKingAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8

		// Generate Knight Attacks
		knightOffsets := [][2]int{
			{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
			{1, -2}, {1, 2}, {2, -1}, {2, 1},
		}
		for _, off := range knightOffsets {
			f, r := file+off[0], rank+off[1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				// Convert back to square index: r*8 + f
				KnightAttacks[sq].Set(r*8 + f)
			}
		}

		// Generate King Attacks
		kingOffsets := [][2]int{
			{-1, -1}, {-1, 0}, {-1, 1},
			{0, -1}, {0, 1},
			{1, -1}, {1, 0}, {1, 1},
		}
		for _, off := range kingOffsets {
			f, r := file+off[0], rank+off[1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				KingAttacks[sq].Set(r*8 + f)
			}
		}
	}
}
