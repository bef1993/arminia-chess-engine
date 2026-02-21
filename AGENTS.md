# Arminia Chess Engine - Developer & Agent Guide

This document provides technical guidance for developers and agents working on the Arminia chess engine codebase.

## Project Architecture

Arminia is a bitboard-based chess engine written in Go, implementing the UCI protocol.

### Directory Structure

```text
cmd/
├── uci/                  # UCI entry point (for Lichess/GUIs)
└── cli/                  # CLI entry point (for manual testing)

internal/
├── engine/               # Core chess logic
│   ├── bitboard.go       # Bitboard definitions, constants, and attack tables
│   ├── board.go          # Board struct (Pieces, Occupancy) and operations
│   ├── fen.go            # FEN parsing and generation
│   ├── game.go           # Game state, move execution, legality checking
│   ├── move_generation.go# Pseudo-legal move generation using bitboards
│   ├── moves.go          # Move struct and UCI string parsing
│   ├── piece.go          # Piece definitions and constants
│   └── zobrist.go        # Zobrist hashing for transposition tables
│
├── search/               # Search algorithms
│   ├── search.go         # Iterative deepening entry point
│   ├── negamax.go        # Alpha-beta search loop
│   ├── quiescence.go     # Quiescence search for stable positions
│   ├── eval.go           # Static evaluation function (PSTs)
│   ├── tt.go             # Transposition Table implementation
│   └── move_ordering.go  # Move ordering heuristics (MVV-LVA)
│
└── uci/                  # UCI Protocol
    └── protocol.go       # Command parsing and main loop
```

## Build and Dependencies

This project targets Go 1.26 or newer. It leverages modern language features, including:

- **`sync.WaitGroup.Go`**: This method, introduced in Go 1.26, provides a concise way to launch a goroutine and automatically manage the WaitGroup counter (Add(1) before, Done() after). All new concurrent code should prefer this method.

## Core Concepts

### 1. Board Representation (Bitboards)

The board is represented using **Bitboards** (`uint64`), where each bit corresponds to a square (0=A1, 63=H8).

- **`Board` Struct**:
  - `Pieces [2][6]Bitboard`: Bitboards for each piece type (Pawn..King) and color (White/Black).
  - `Occupancy [3]Bitboard`: Aggregate bitboards for White, Black, and Both.

- **Helpers**:
  - `GetRank(sq)`, `GetFile(sq)`, `GetSq(file, rank)`: Coordinate conversion.
  - `PopLSB()`: Efficiently iterates over set bits in a bitboard.

### 2. Move Generation

Move generation is **pseudo-legal** followed by **legality filtering**.

1. **`GenerateAllPseudoLegalMoves()`** (`move_generation.go`):
    - Iterates over piece bitboards.
    - Uses pre-calculated attack tables (`KnightAttacks`, `KingAttacks`, `PawnAttacks`) and sliding piece generators.
    - Generates moves that *might* leave the king in check.
2. **`GenerateLegalMoves()`** (`game.go`):
    - Calls `GenerateAllPseudoLegalMoves()`.
    - Filters moves using `isKingInCheckAfterMove()`, which temporarily makes the move and checks if the king is attacked.

### 3. Search & Evaluation

- **Algorithm**: Negamax with Alpha-Beta pruning and Iterative Deepening.
- **Quiescence Search**: Explores captures at leaf nodes to avoid the horizon effect.
- **Transposition Table**: Zobrist hashing is used to cache positions and scores.
- **Move Ordering**:
  1. Hash Move (from TT)
  2. Captures (MVV-LVA)
  3. Promotions
  4. Quiet Moves
- **Evaluation**: Material balance + Piece-Square Tables (PST) for positional understanding.

## Development Status

| Feature       | Status     | Notes                                     |
|:--------------|:-----------|:------------------------------------------|
| **Bitboards** | ✅ Complete | Fully implemented for board and move gen. |
| **Move Gen**  | ✅ Complete | Efficient bitboard-based generation.      |
| **Search**    | ✅ Complete | Alpha-Beta, ID, Quiescence, TT.           |
| **Eval**      | ✅ Complete | Material + PST.                           |
| **UCI**       | ✅ Complete | Supports standard commands + Hash option. |
| **Lichess**   | ✅ Ready    | Can be used with `lichess-bot`.           |

## Next Steps (Phase 6: Expert Features)

The engine is functional and strong (estimated ~1800+ Elo). The next phase focuses on advanced optimizations:

1. **Tapered Evaluation**: Interpolate between Middlegame and Endgame PSTs based on game phase.
2. **Killer Moves**: Store quiet moves that caused cutoffs at specific ply depths to prioritize them.
3. **History Heuristic**: Score quiet moves based on their historical success to improve ordering.
4. **Endgame Tablebases**: Integrate Syzygy tablebases for perfect endgame play.

## Testing

- **Unit Tests**: Run `go test ./...` to verify all components.
- **Perft**: Use `internal/engine/perft_test.go` to verify move generation counts against known values.
- **CLI**: Use `bin/arminia-cli` for manual interaction and debugging.

### Best Practices for New Tests

When adding new tests, prefer using human-readable strings for squares and moves to improve readability and maintainability.

- **Setup**: Use `game.Board.SetPieceAt("e4", WhitePawn)` instead of raw indices.
- **Moves**: Use `game.ParseMove("e2e4")` to create moves from algebraic notation.
- **Helpers**: Use `Sq("e4")` if you need the integer index of a square.

Example:

```go
func TestMyFeature(t *testing.T) {
    game := NewEmptyGame()
    game.Board.SetPieceAt("e1", WhiteKing)
    
    move, _ := game.ParseMove("e1g1")
    game.ExecuteMove(move)
    
    assert.Equal(t, WhiteKing, game.Board.GetPieceAt("g1"))
}
```

## Code Style

- Use `go fmt`.
- Prefer explicit variable names (`rank`, `file`, `sq`) over generic ones (`r`, `c`, `i`).
- Use `slog` for logging (only in UCI mode or CLI, never in search loop).
