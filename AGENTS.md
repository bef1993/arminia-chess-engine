# Arminia Chess Engine - Developer & Agent Guide

This document provides technical guidance for developers and agents working on the Arminia chess engine codebase.

## Project Layout

See [README.md](README.md) for the current project file structure.

## Key Files & Responsibilities

### Core Engine (internal/engine/)

| File       | Purpose                   | Key Types/Functions                                                    |
| :--------- | :------------------------ | :--------------------------------------------------------------------- |
| `piece.go` | Chess piece definitions   | `PieceType` (enum), `Color` (enum), `Piece` struct, `GetSymbol()`      |
| `board.go` | Board state & operations  | `Board` struct, `NewBoard()`, `MovePiece()`, `GetLegalMoves()`         |
| `game.go`  | Game management           | `Game` struct, `NewGame()`, `PrintBoard()`, `SwitchTurn()`             |
| `moves.go` | Move generation algorithm | `MoveGenerator`, `GenerateMovesForPiece()`, 6 piece-specific functions |

### UCI Protocol (internal/uci/)

| File          | Purpose             | Key Functions                                                     |
| :------------ | :------------------ | :---------------------------------------------------------------- |
| `protocol.go` | UCI command handler | `Protocol` struct, `Run()`, `handleCommand()`, 7 command handlers |

### Search (internal/search/)

| File        | Purpose             | Key Functions           |
| :---------- | :------------------ | :---------------------- |
| `search.go` | Search algorithm    | `Search()`, `negamax()` |
| `eval.go`   | Evaluation function | `Evaluate()`            |

### Coordinate System

**Critical:** The board uses **(col, row)** indexing following chess notation:

- **col**: 0-7 representing files a-h (left to right)
- **row**: 0-7 representing ranks 8-1 (top to bottom in code, but inverted)

```text
Chess notation:  a8 b8 c8 ... h8
Code mapping:    (0,0) (1,0) (2,0) ... (7,0)
                 ...
Code mapping:    (0,7) (1,7) (2,7) ... (7,7)
Chess notation:  a1 b1 c1 ... h1
```

Conversion formulas:

- From algebraic (e.g., "e4"): `col = e - 'a' = 4`, `row = 8 - 4 = 4`
- To algebraic: `file = 'a' + col`, `rank = 8 - row`

### Global Constants

Use these constants instead of raw integers to ensure readability and prevent off-by-one errors.

| Category     | Constants                                                            | Value/Note                     |
| :----------- | :------------------------------------------------------------------- | :----------------------------- |
| **Files**    | `FileA` ... `FileH`                                                  | 0 ... 7                        |
| **Ranks**    | `Rank8` ... `Rank1`                                                  | 0 ... 7 (Note: Rank8 is row 0) |
| **Colors**   | `White`, `Black`                                                     |                                |
| **Pieces**   | `Pawn`, `Knight`, `Bishop`, `Rook`, `Queen`, `King`                  |                                |
| **Empty**    | `NoPiece`                                                            | Represents empty square        |
| **Castling** | `WhiteKingside`, `WhiteQueenside`, `BlackKingside`, `BlackQueenside` | Bitmasks                       |
| **Castling** | `AllCastling`, `NoCastling`                                          | Helpers                        |

## Development Tasks

### Building

```bash
# Build both executables
go build -o bin/arminia-uci.exe ./cmd/uci
go build -o bin/arminia-cli.exe ./cmd/cli
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/engine -v
go test ./internal/uci -v

# Run with coverage
go test ./... -cover

# Run specific test
go test ./internal/engine -run TestGeneratePawnMoves -v
```

### Common Development Pattern

1. **Modify core logic** → `internal/engine/*.go`
2. **Add tests** → `*_test.go` in same package
3. **Test changes** → `go test ./internal/engine -v`
4. **Rebuild binaries** → `go build -o bin/arminia-*.exe ./cmd/...`
5. **Manual testing** → `echo "commands" | .\bin\arminia-uci.exe`

## Known Limitations & TODOs

### Move Generation

- ✅ Generates all pseudo-legal moves
- ✅ Check detection (IsKingInCheck implemented)
- ✅ Checkmate detection
- ✅ Stalemate detection
- ✅ En passant support
- ✅ Castling support
- ✅ Pawn promotion logic

### Move Selection

- ✅ Search algorithm (Negamax with Alpha-Beta)
- ✅ Basic Evaluation function (Material)
- ❌ Fixed depth only (Depth 4)
- ❌ No time management or iterative deepening
- ❌ No move ordering
- ❌ No Quiescence search

### UCI Protocol

- ✅ Basic command parsing
- ✅ Position setup from startpos
- ✅ Move execution via UCI
- ✅ FEN support
- ✅ Move validation (rejects illegal moves)
- ❌ No time management

### Board State

- ✅ Piece placement
- ✅ Turn management
- ✅ Half-move clock (50-move rule)
- ✅ Full-move counter
- ✅ Castling rights tracking
- ✅ En passant target tracking

## Code Patterns

### Move Struct

```go
type Move struct {
    FromCol        int
    FromRow        int
    ToCol          int
    ToRow          int
    PromotionPiece Piece // 0 if no promotion
}
```

### Accessing Board

```go
// Helper methods using algebraic notation (preferred for tests)
board.SetPieceAt("e4", WhitePawn)
board.RemovePieceAt("e4")
piece := board.GetPieceAt("e4")

// Get piece at square
piece := board.GetPiece(col, row)

// Check if square is empty
isEmpty := board.IsEmpty(col, row)

// Check if square has player's piece
isOwnPiece := board.IsOccupiedByColor(col, row, White)

// Move piece
success := board.MovePiece(fromCol, fromRow, toCol, toRow)

// Generate all legal moves for a color
moves := board.GetLegalMoves(White)
```

### Adding New Piece Move Generation

All piece moves go in `internal/engine/moves.go`:

```go
func (mg *MoveGenerator) generateXxxMoves(col, row int, color Color) []Move {
    var moves []Move
    
    // Generate moves based on piece logic
    // Check bounds: 0 <= col < 8 && 0 <= row < 8
    // Check piece collision: mg.Board.IsEmpty() or !mg.Board.IsOccupiedByColor()
    
    return moves
}
```

### Adding Tests

Tests follow standard Go pattern in `*_test.go` files. Use `github.com/stretchr/testify/assert` for assertions.
Tests follow standard Go pattern in `*_test.go` files:

```go
func TestFeature(t *testing.T) {
    board := NewBoard()  // or NewEmptyBoard()
    
    // Setup using algebraic notation
    board.SetPieceAt("e4", WhitePawn)
    
    // Execute
    result := someFunction()
    
    // Assert using testify
    assert.Equal(t, expected, result)
}
```

## Testing Strategy

### Test Organization

- **piece_test.go**: Piece creation, symbol generation
- **board_test.go**: Board operations, piece access, movement
- **game_test.go**: Game initialization, turn switching
- **moves_test.go**: Move generation for all pieces (biggest test file)
- **protocol_test.go**: UCI command handling

### Test Coverage by Area

- Move generation: 25+ tests (board state, piece-specific, captures, blocks)
- Board operations: 15+ tests (placement, access, bounds)
- Game state: 5+ tests (initialization, turn switching)
- UCI protocol: 10+ tests (command parsing, position, moves)

### NewEmptyBoard() vs NewBoard()

- `NewBoard()`: Standard chess starting position
- `NewEmptyBoard()`: Empty 8x8 board (useful for isolated move tests)

## Phase Dependencies

### Phase 1: Core Functionality ✅ COMPLETE

- [x] Board representation
- [x] Piece definitions
- [x] Move generation (all pieces)
- [x] Unit tests (60+)
- [x] Modular package structure

### Phase 2: UCI Protocol & Move Legality ✅ COMPLETE

**Completed items:**

- [x] Check detection (needed for legal move validation)
- [x] Checkmate detection (needed to prevent illegal moves)
- [x] Move validation (reject moves that leave king in check)
- [x] FEN support (Moved to Phase 3)
- [x] Basic UCI working with full move validation

### Phase 3: Search & Evaluation ✅ COMPLETE

**Completed items:**

- [x] Legal move validation (from Phase 2)
- [x] Evaluation function (Material counting)
- [x] Minimax with alpha-beta pruning

### Phase 4: Advanced Features ⏳ IN PROGRESS

- Quiescence search
- Iterative Deepening
- Time management
- Move ordering
- Opening book
- Endgame tables

### Phase 5: Lichess Integration 🎯 HANDLED BY dolegi/lichess-bot

**Status:** Not required for core development

Instead of building custom Lichess integration, use the existing [dolegi/lichess-bot](https://github.com/dolegi/lichess-bot) project:

- Handles all Lichess API communication
- Manages bot account upgrades
- Streams challenges and games
- Spawns UCI engines as subprocesses

**Integration**: Once your engine reaches Phase 2 (legal moves), it can be used via:

```bash
./lichess-bot config.toml
```

Where `config.toml` points to your Arminia UCI binary

## Next Steps for Agents

### High Priority: Advanced Search & Time Management (Phase 4)

The basic search is working, but it plays at a fixed depth and doesn't manage time.

#### 1. Implement Iterative Deepening (`internal/search/search.go`)

- **Logic:**
  - Instead of calling `negamax(depth)`, call it in a loop: depth 1, 2, 3...
  - This allows the engine to always have a "best move" ready if time runs out.
  - It also helps with move ordering (best move from depth d-1 is searched first at depth d).

#### 2. Implement Time Management (`internal/uci/protocol.go` & `internal/search/`)

- **Logic:**
  - Calculate allocated time for the move based on `wtime`, `btime`, `winc`, `binc`.
  - Standard formula: `Time = (TimeLeft / MovesToGo) + Increment`. Default MovesToGo ~ 40.
  - Pass a `context` or `stop` channel to the search function.
  - In `negamax`, check periodically (every 2048 nodes) if time is up.

#### 3. Implement Quiescence Search (`internal/search/search.go`)

- **Problem:** The horizon effect (engine stops searching in the middle of a capture sequence).
- **Logic:**

  - At depth 0, instead of calling `Evaluate()`, call `Quiescence()`.
  - `Quiescence` only searches captures (and maybe checks).
  - It uses a "standing pat" score (evaluation of current position) as a lower bound.

### Testing Requirements

- Every new feature must have corresponding unit tests
- All 60+ existing tests must pass
- New tests should follow existing patterns in `*_test.go` files
- Aim for >80% code coverage

## Debugging Tips

### Binary Testing

```powershell
# Test UCI directly
"uci`nquit" | .\bin\arminia-uci.exe

# Test position command
"position startpos`ngo`nquit" | .\bin\arminia-uci.exe

# Test specific moves
"position startpos moves e2e4 e7e5`ngo`nquit" | .\bin\arminia-uci.exe
```

### Unit Test Debugging

```bash
# Run single test with output
go test ./internal/engine -run TestGenerateBishopMovesFromMiddle -v

# Run with print statements
go test ./internal/engine -v -count=1
```

### Common Issues

**Issue:** Tests fail after code changes

- **Fix:** Verify coordinate system hasn't been confused (col vs row)
- Check: Are you using (col, row) consistently?

**Issue:** Move generation counts are wrong

- **Fix:** Use `NewEmptyBoard()` for isolated tests
- **Fix:** Verify board bounds checking in loops

**Issue:** UCI binary doesn't respond

- **Fix:** Check if command ends with newline: `"command\nquit\n"`
- **Fix:** Verify binary was rebuilt with `go build`

## Implementation Checklist Template

When working on a feature:

- [ ] Read related code and tests
- [ ] Write test cases first (TDD)
- [ ] Implement feature
- [ ] Run `go test ./...` - all must pass
- [ ] Run `go build -o bin/arminia-*.exe ./cmd/...` - must succeed
- [ ] Manual test with binaries
- [ ] Verify no new lint/format issues
- [ ] Document changes in code comments
