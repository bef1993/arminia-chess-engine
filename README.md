# Arminia Chess Engine

A chess engine written in Go with the goal of implementing the UCI protocol to compete against other engines on Lichess and other platforms.

**Play against Arminia on Lichess:** [https://lichess.org/@/Arminia-Bot](https://lichess.org/@/Arminia-Bot)

Powered by [lichess-bot](https://github.com/lichess-bot-devs/lichess-bot).

## Multiple Entry Points

The modular structure enables different use cases:

- **`arminia-uci.exe`** - UCI mode for competing on Lichess/engines
- **`arminia-cli.exe`** - Interactive command-line interface for testing

## Project Structure

The codebase is organized into modular packages:

- **`cmd/`**: Application entry points (`uci` and `cli`).
- **`internal/`**: Core application logic.
  - **`engine/`**: Board representation, move generation, and game state management.
  - **`search/`**: Search algorithms (Negamax, Quiescence), evaluation, and transposition tables.
  - **`uci/`**: UCI protocol handling for communication with chess GUIs.

## Features

- **Board Representation**: Standard 8x8 chess board with piece placement. **Bitboard representation** is partially implemented (Phase 6).
- **Piece Types**: Pawn, Knight, Bishop, Rook, Queen, King
- **Move Generation**: Complete legal move generation for all piece types
- **Move Validation**: Rejects illegal moves
- **Special Moves**: Castling, En Passant, Pawn Promotion
- **Game State**: Tracks current turn and move history, castling rights, en passant target
- **Draw Detection**: Stalemate, 50-move rule, Insufficient Material, Threefold Repetition
- **Display**: ASCII board visualization with Unicode chess symbols
- **Comprehensive Tests**: Full test coverage for board and move generation

## Building and Running

### Prerequisites

- Go 1.26 or later

### Build All

```bash
go build -o bin/arminia-uci.exe ./cmd/uci
go build -o bin/arminia-cli.exe ./cmd/cli
```

### Run Interactive CLI

```bash
# For manual testing
.\bin\arminia-cli.exe
```

### Run Tests

```bash
# Run all tests with coverage
go test ./...

# Run specific package tests
go test ./internal/engine -v
go test ./internal/uci -v
```

## UCI Protocol Support

Arminia implements the **UCI (Universal Chess Interface)** protocol for engine competition. Currently supported:

- ✅ `uci` - Engine identification
- ✅ `isready` - Readiness check
- ✅ `position startpos` - Board setup
- ✅ `position fen` - FEN board setup
- ✅ `go` - Move generation
- ✅ `setoption` - Option configuration
- ✅ `ucinewgame` - Game reset
- ✅ `quit` - Exit

Move notation: long algebraic (e.g., `e2e4`, `e7e8q`)

## Development Status

| Phase | Status         | Description                                           |
| :---- | :------------- | :---------------------------------------------------- |
| **1** | ✅ Complete    | Board, pieces, move generation                        |
| **2** | ✅ Complete    | UCI protocol, move validation, special moves          |
| **3** | ✅ Complete    | FEN support, Search algorithm, evaluation function    |
| **4** | ✅ Complete    | Advanced features (quiescence, time management, etc.) |
| **5** | ✅ Complete    | Lichess integration (via lichess-bot)                 |
| **6** | ⏳ In Progress | Expert features (Killer moves, Bitboards, SMP)        |

For detailed roadmap and implementation guidance, see [AGENTS.md](AGENTS.md).

## Implemented Optimizations

- **Zobrist Hashing**: Efficient board state representation for table lookups.
- **Transposition Tables**: Caching search results to avoid redundant calculations.
- **Alpha-Beta Pruning**: Significantly reducing the search space in the Negamax algorithm.
- **Iterative Deepening**: Searching to increasing depths to ensure a best move is always available within time limits.
- **Quiescence Search**: Extending search at leaf nodes to avoid the horizon effect in volatile positions.
- **Move Ordering**: MVV-LVA (Most Valuable Victim - Least Valuable Aggressor) heuristic to prioritize captures.
- **Bitboards (Partial)**: Basic bitboard infrastructure implemented; move generation migration pending.
- **Killer Moves (Planned)**: Heuristic to prioritize quiet moves that caused cutoffs at the same search depth.
- **History Heuristic (Planned)**: Global table to prioritize moves that frequently cause cutoffs.

## Next Steps

**Future Improvements (Phase 6):**

- [x] Bitboard Representation (First version implemented)
- [ ] Migrate Move Generation to Bitboards (Optimization)
- [ ] Advanced Parallel Search (Lazy SMP)
- [ ] More Sophisticated Evaluation (Piece-Square Tables, Variable Piece Values)
- [ ] Opening Books
- [ ] Endgame Tablebases
- [ ] Killer Moves
- [ ] History Heuristic
- [ ] Check Extensions

See [AGENTS.md](AGENTS.md) for development tasks, code patterns, and debugging tips.

Reference: <https://lichess.org/api#tag/Bot>

## Testing

The engine includes comprehensive unit tests:

```bash
# Run all tests
go test ./... -v

# Build binaries
go build -o ./bin/arminia-engine.exe ./cmd/uci
go build -o ./bin/arminia-cli.exe ./cmd/cli

```

## License

MIT License
