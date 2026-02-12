# Arminia Chess Engine

A chess engine written in Go with the goal of implementing the UCI protocol to compete against other engines on Lichess and other platforms.

## Project Structure

```
arminia-chess-engine/
├── go.mod                          # Go module definition
├── go.sum                          # Go dependencies
├── README.md                       # This file
│
├── cmd/                            # Application entry points
│   ├── uci/
│   │   └── main.go                 # UCI mode for engine competition
│   └── cli/
│       └── main.go                 # Interactive CLI for manual testing
│
├── internal/                       # Private packages (Go visibility rules)
│   ├── engine/                     # Core chess engine logic
│   │   ├── piece.go                # Piece definitions (types, colors)
│   │   ├── board.go                # Board representation & operations
│   │   ├── game.go                 # Game state management
│   │   ├── moves.go                # Move generation for all pieces
│   │   └── *_test.go               # Comprehensive unit tests (60+)
│   │
│   └── uci/                        # UCI protocol implementation
│       ├── protocol.go             # UCI command handler
│       └── protocol_test.go        # UCI protocol tests
│
└── bin/                            # Built executables
    ├── arminia-uci.exe             # UCI engine executable
    └── arminia-cli.exe             # Interactive CLI executable
```

## Multiple Entry Points

The modular structure enables different use cases:

- **`arminia-uci.exe`** - UCI mode for competing on Lichess/engines
- **`arminia-cli.exe`** - Interactive command-line interface for testing

## Features

- **Board Representation**: Standard 8x8 chess board with piece placement
- **Piece Types**: Pawn, Knight, Bishop, Rook, Queen, King
- **Move Generation**: Complete legal move generation for all piece types
- **Special Moves**: Castling, En Passant, Pawn Promotion
- **Game State**: Tracks current turn and move history
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

### Run UCI Mode
```bash
# For Lichess/engine competition
.\bin\arminia-uci.exe
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
- ✅ `go` - Move generation
- ✅ `setoption` - Option configuration
- ✅ `ucinewgame` - Game reset
- ✅ `quit` - Exit
- ❌ FEN positions (planned)
- ❌ Move validation (planned)

Move notation: long algebraic (e.g., `e2e4`, `e7e8q`)

## Development Status

| Phase | Status | Description |
|-------|--------|-------------|
| **1** | ✅ Complete | Board, pieces, move generation |
| **2** | ✅ Complete | UCI protocol, move validation, special moves |
| **3** | ⏳ In Progress | FEN support, Search algorithm, evaluation function |
| **4** | 🚫 Blocked | Advanced features (quiescence, opening book, etc.) |
| **5** | 🚫 Blocked | Lichess integration |

For detailed roadmap and implementation guidance, see [AGENTS.md](AGENTS.md).

## Next Steps

**High Priority (Phase 3):**
- [x] Check detection
- [x] Checkmate detection
- [x] Move validation (prevent illegal moves)
- [ ] FEN position support
- [ ] Evaluation function
- [ ] Minimax search

See [AGENTS.md](AGENTS.md) for development tasks, code patterns, and debugging tips.

Reference: https://lichess.org/api#tag/Bot

## Similar Engines

- **Stockfish** - Leading open-source chess engine
- **Leela Chess Zero (Lc0)** - Neural network-based engine
- **Komodo** - Commercial engine
- **Rybka** - Specialized UCI engine

## Performance Goals

- **Elo Rating**: Target 1600+ after Phase 3
- **Search Speed**: 100,000+ nodes per second
- **Opening Strength**: Standard openings library
- **Endgame**: Tablebase integration for perfect play

## Next Steps

See the Development Roadmap above for the implementation sequence.

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
