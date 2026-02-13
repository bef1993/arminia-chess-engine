# Arminia Chess Engine

A chess engine written in Go with the goal of implementing the UCI protocol to compete against other engines on Lichess and other platforms.

**Play against Arminia on Lichess:** [https://lichess.org/@/Arminia-Bot](https://lichess.org/@/Arminia-Bot)

Powered by [lichess-bot](https://github.com/lichess-bot-devs/lichess-bot).

## Multiple Entry Points

The modular structure enables different use cases:

- **`arminia-uci.exe`** - UCI mode for competing on Lichess/engines
- **`arminia-cli.exe`** - Interactive command-line interface for testing

## Features

- **Board Representation**: Standard 8x8 chess board with piece placement
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
| **4** | ⏳ In Progress | Advanced features (quiescence, time management, etc.) |
| **5** | ✅ Complete    | Lichess integration (via lichess-bot)                 |

For detailed roadmap and implementation guidance, see [AGENTS.md](AGENTS.md).

## Implemented Optimizations

- **Zobrist Hashing**: Efficient board state representation for table lookups.
- **Transposition Tables**: Caching search results to avoid redundant calculations.
- **Alpha-Beta Pruning**: Significantly reducing the search space in the Negamax algorithm.
- **Quiescence Search**: Extending search at leaf nodes to avoid the horizon effect in volatile positions.

## Next Steps

**High Priority (Phase 4 - Advanced Search):**

- [x] Iterative Deepening & Time Management
- [ ] Heuristic Move Ordering

**Future Improvements (Phase 5+):**

- [ ] Bitboard Representation
- [ ] Advanced Parallel Search (Lazy SMP)
- [ ] More Sophisticated Evaluation (Piece-Square Tables, Variable Piece Values)
- [ ] Opening Books
- [ ] Endgame Tablebases

See [AGENTS.md](AGENTS.md) for development tasks, code patterns, and debugging tips.

Reference: <https://lichess.org/api#tag/Bot>

## Performance Goals

- **Elo Rating**: Target 1600+ after Phase 3
- **Search Speed**: 100,000+ nodes per second
- **Opening Strength**: Standard openings library
- **Endgame**: Tablebase integration for perfect play

## Testing

The engine includes comprehensive unit tests:

```bash
# Run all tests with verbose output
go test ./engine -v

# Run specific test
go test ./engine -run TestGeneratePawnMoves -v

# Run with coverage
go test ./engine -cover
```

## License

MIT License
