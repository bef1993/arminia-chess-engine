package engine

import (
	"math/bits"
	"math/rand"
)

type Bitboard uint64

const (
	A1 = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
)

// Bitboard Constants
const (
	FileA_BB Bitboard = 0x0101010101010101
	FileB_BB Bitboard = FileA_BB << 1
	FileC_BB Bitboard = FileA_BB << 2
	FileD_BB Bitboard = FileA_BB << 3
	FileE_BB Bitboard = FileA_BB << 4
	FileF_BB Bitboard = FileA_BB << 5
	FileG_BB Bitboard = FileA_BB << 6
	FileH_BB Bitboard = FileA_BB << 7

	Rank1_BB Bitboard = 0xFF
	Rank2_BB Bitboard = Rank1_BB << 8
	Rank3_BB Bitboard = Rank1_BB << 16
	Rank4_BB Bitboard = Rank1_BB << 24
	Rank5_BB Bitboard = Rank1_BB << 32
	Rank6_BB Bitboard = Rank1_BB << 40
	Rank7_BB Bitboard = Rank1_BB << 48
	Rank8_BB Bitboard = Rank1_BB << 56
)

// Basic Operations
func (b *Bitboard) Set(sq int)        { *b |= (1 << sq) }
func (b *Bitboard) Clear(sq int)      { *b &= ^(1 << sq) }
func (b *Bitboard) IsSet(sq int) bool { return (*b & (1 << sq)) != 0 }

// Population Count (number of set bits)
func (b *Bitboard) Count() int { return bits.OnesCount64(uint64(*b)) }

// Get and Clear Least Significant Bit (used for iteration)
func (b *Bitboard) PopLSB() int {
	lsb := bits.TrailingZeros64(uint64(*b))
	*b &= *b - 1 // Clear the LSB
	return lsb
}

// Magic holds magic bitboard parameters for a square
type Magic struct {
	Mask   Bitboard
	Magic  uint64
	Shift  int
	Offset int
}

var (
	RookMagics   [64]Magic
	BishopMagics [64]Magic

	RookAttackTable   []Bitboard
	BishopAttackTable []Bitboard
)

// Pre-calculated attack tables
var (
	KnightAttacks [64]Bitboard
	KingAttacks   [64]Bitboard
	PawnAttacks   [2][64]Bitboard // [Color][Square]
)

func init() {
	initKnightAndKingAttacks()
	initPawnAttacks()
	initMagicBitboards()
}

func initKnightAndKingAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := GetFile(sq)
		rank := GetRank(sq)

		// Generate Knight Attacks
		knightOffsets := [][2]int{
			{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
			{1, -2}, {1, 2}, {2, -1}, {2, 1},
		}
		for _, off := range knightOffsets {
			f, r := file+off[0], rank+off[1]
			if IsOnBoard2D(f, r) {
				// Convert back to square index: r*8 + f
				KnightAttacks[sq].Set(GetSq(f, r))
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
			if IsOnBoard2D(f, r) {
				KingAttacks[sq].Set(GetSq(f, r))
			}
		}
	}
}

func initPawnAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := GetFile(sq)
		rank := GetRank(sq)

		// White Pawns (attack "up" / rank + 1)
		if rank < 7 {
			if file > 0 { PawnAttacks[White][sq].Set(GetSq(file-1, rank+1)) }
			if file < 7 { PawnAttacks[White][sq].Set(GetSq(file+1, rank+1)) }
		}

		// Black Pawns (attack "down" / rank - 1)
		if rank > 0 {
			if file > 0 { PawnAttacks[Black][sq].Set(GetSq(file-1, rank-1)) }
			if file < 7 { PawnAttacks[Black][sq].Set(GetSq(file+1, rank-1)) }
		}
	}
}

// GetBishopAttacks returns the attack bitboard for a bishop on a given square with specific occupancy
func GetBishopAttacks(sq int, occupancy Bitboard) Bitboard {
	m := &BishopMagics[sq]
	idx := (uint64(occupancy&m.Mask) * m.Magic) >> m.Shift
	return BishopAttackTable[m.Offset+int(idx)]
}

// GetRookAttacks returns the attack bitboard for a rook on a given square with specific occupancy
func GetRookAttacks(sq int, occupancy Bitboard) Bitboard {
	m := &RookMagics[sq]
	idx := (uint64(occupancy&m.Mask) * m.Magic) >> m.Shift
	return RookAttackTable[m.Offset+int(idx)]
}

// GetQueenAttacks returns the attack bitboard for a queen (combination of bishop and rook)
func GetQueenAttacks(sq int, occupancy Bitboard) Bitboard {
	return GetBishopAttacks(sq, occupancy) | GetRookAttacks(sq, occupancy)
}

// Magic Bitboard Initialization

func initMagicBitboards() {
	// Initialize tables
	// Sizes are approximate based on max bits for rook (12) and bishop (9)
	// We'll append to them dynamically or pre-allocate if we were strict.
	// For simplicity in this generator, we'll just let them grow or pre-calc offsets.
	// Actually, to keep it simple and contiguous:
	RookAttackTable = make([]Bitboard, 0, 102400)   // ~100k entries
	BishopAttackTable = make([]Bitboard, 0, 5248) // ~5k entries

	for sq := 0; sq < 64; sq++ {
		// Rooks
		mask := rookMask(sq)
		bits := mask.Count()
		RookMagics[sq] = findMagic(sq, mask, bits, true)

		// Bishops
		mask = bishopMask(sq)
		bits = mask.Count()
		BishopMagics[sq] = findMagic(sq, mask, bits, false)
	}
}

func findMagic(sq int, mask Bitboard, mBits int, isRook bool) Magic {
	var magic Magic
	magic.Mask = mask
	magic.Shift = 64 - mBits

	// Generate all occupancy variations
	occupancies := getOccupancyVariations(mask)
	attacks := make([]Bitboard, len(occupancies))

	for i, occ := range occupancies {
		if isRook {
			attacks[i] = rookAttacksSlow(sq, occ)
		} else {
			attacks[i] = bishopAttacksSlow(sq, occ)
		}
	}

	// Find magic number
	// Use a fixed seed for deterministic behavior
	rng := rand.New(rand.NewSource(int64(sq) + 1))

	table := make([]Bitboard, 1<<mBits)

	for {
		// Generate candidate magic
		cand := rng.Uint64() & rng.Uint64() & rng.Uint64() // Sparse random

		// Verify magic
		for i := range table {
			table[i] = 0
		}
		ok := true

		for i, occ := range occupancies {
			idx := (uint64(occ) * cand) >> magic.Shift
			if table[idx] != 0 && table[idx] != attacks[i] {
				ok = false
				break
			}
			table[idx] = attacks[i]
		}

		if ok {
			magic.Magic = cand
			// Append table to global table and set offset
			if isRook {
				magic.Offset = len(RookAttackTable)
				RookAttackTable = append(RookAttackTable, table...)
			} else {
				magic.Offset = len(BishopAttackTable)
				BishopAttackTable = append(BishopAttackTable, table...)
			}
			return magic
		}
	}
}

func getOccupancyVariations(mask Bitboard) []Bitboard {
	count := mask.Count()
	size := 1 << count
	variations := make([]Bitboard, size)

	// Map bits of mask to indices
	bitIndices := make([]int, count)
	tempMask := mask
	for i := 0; i < count; i++ {
		bitIndices[i] = tempMask.PopLSB()
	}

	for i := 0; i < size; i++ {
		var occ Bitboard
		for j := 0; j < count; j++ {
			if (i & (1 << j)) != 0 {
				occ.Set(bitIndices[j])
			}
		}
		variations[i] = occ
	}
	return variations
}

func rookMask(sq int) Bitboard {
	var mask Bitboard
	r, f := GetRank(sq), GetFile(sq)
	for r2 := r + 1; r2 < 7; r2++ { mask.Set(GetSq(f, r2)) }
	for r2 := r - 1; r2 > 0; r2-- { mask.Set(GetSq(f, r2)) }
	for f2 := f + 1; f2 < 7; f2++ { mask.Set(GetSq(f2, r)) }
	for f2 := f - 1; f2 > 0; f2-- { mask.Set(GetSq(f2, r)) }
	return mask
}

func bishopMask(sq int) Bitboard {
	var mask Bitboard
	r, f := GetRank(sq), GetFile(sq)
	for r2, f2 := r+1, f+1; r2 < 7 && f2 < 7; r2, f2 = r2+1, f2+1 { mask.Set(GetSq(f2, r2)) }
	for r2, f2 := r+1, f-1; r2 < 7 && f2 > 0; r2, f2 = r2+1, f2-1 { mask.Set(GetSq(f2, r2)) }
	for r2, f2 := r-1, f+1; r2 > 0 && f2 < 7; r2, f2 = r2-1, f2+1 { mask.Set(GetSq(f2, r2)) }
	for r2, f2 := r-1, f-1; r2 > 0 && f2 > 0; r2, f2 = r2-1, f2-1 { mask.Set(GetSq(f2, r2)) }
	return mask
}

func rookAttacksSlow(sq int, block Bitboard) Bitboard {
	var attacks Bitboard
	r, f := GetRank(sq), GetFile(sq)
	for r2 := r + 1; r2 < 8; r2++ { s := GetSq(f, r2); attacks.Set(s); if block.IsSet(s) { break } }
	for r2 := r - 1; r2 >= 0; r2-- { s := GetSq(f, r2); attacks.Set(s); if block.IsSet(s) { break } }
	for f2 := f + 1; f2 < 8; f2++ { s := GetSq(f2, r); attacks.Set(s); if block.IsSet(s) { break } }
	for f2 := f - 1; f2 >= 0; f2-- { s := GetSq(f2, r); attacks.Set(s); if block.IsSet(s) { break } }
	return attacks
}

func bishopAttacksSlow(sq int, block Bitboard) Bitboard {
	var attacks Bitboard
	r, f := GetRank(sq), GetFile(sq)
	for r2, f2 := r+1, f+1; r2 < 8 && f2 < 8; r2, f2 = r2+1, f2+1 { s := GetSq(f2, r2); attacks.Set(s); if block.IsSet(s) { break } }
	for r2, f2 := r+1, f-1; r2 < 8 && f2 >= 0; r2, f2 = r2+1, f2-1 { s := GetSq(f2, r2); attacks.Set(s); if block.IsSet(s) { break } }
	for r2, f2 := r-1, f+1; r2 >= 0 && f2 < 8; r2, f2 = r2-1, f2+1 { s := GetSq(f2, r2); attacks.Set(s); if block.IsSet(s) { break } }
	for r2, f2 := r-1, f-1; r2 >= 0 && f2 >= 0; r2, f2 = r2-1, f2-1 { s := GetSq(f2, r2); attacks.Set(s); if block.IsSet(s) { break } }
	return attacks
}
