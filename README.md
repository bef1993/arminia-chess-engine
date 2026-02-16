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

## Building, Testing and Running

### Prerequisites

- Go 1.26 or later
- UCI-compatible tool or GUI
- To use the bot on Lichess <https://github.com/lichess-bot-devs/lichess-bot> has to be used as a bridge (use config.yml)

Lichess integration reference: <https://lichess.org/api#tag/Bot>

### Build All

```bash
go build -o bin/arminia-engine ./cmd/uci &&
go build -o bin/arminia-engine-cli ./cmd/cli
```

### Run Tests

```bash
# Run all tests
go test ./...
go test -cover ./... # with coverage
```

### Run Interactive CLI

```bash
# For basic manual testing
./bin/arminia-cli
```

See [AGENTS.md](AGENTS.md) for development tasks, code patterns, and tips.

## UCI Protocol Support

Arminia implements the **UCI (Universal Chess Interface)** protocol for engine competition. Currently supported:

- ✅ `uci` - Engine identification
- ✅ `isready` - Readiness check
- ✅ `position startpos` - Board setup
- ✅ `position fen` - FEN board setup
- ✅ `go` - Move generation
- ✅ `setoption` - Option configuration
  - `Hash` - set the transposition table to given size in MB
- ✅ `ucinewgame` - Game reset
- ✅ `stop` - stop the current search
- ✅ `quit` - Exit

Move notation: long algebraic (e.g., `e2e4`, `e7e8q`)

## Development Status

| Phase | Status        | Description                                           |
|:------|:--------------|:------------------------------------------------------|
| **1** | ✅ Complete    | Board, pieces, move generation                        |
| **2** | ✅ Complete    | UCI protocol, move validation, special moves          |
| **3** | ✅ Complete    | FEN support, Search algorithm, evaluation function    |
| **4** | ✅ Complete    | Advanced features (quiescence, time management, etc.) |
| **5** | ✅ Complete    | Lichess integration (via lichess-bot)                 |
| **6** | ⏳ In Progress | Expert features (Killer moves, Bitboards, SMP)        |

## Implemented Optimizations

- **Zobrist Hashing**: Efficient board state representation for table lookups.
- **Transposition Tables**: Caching search results to avoid redundant calculations.
- **Alpha-Beta Pruning**: Significantly reducing the search space in the Negamax algorithm.
- **Iterative Deepening**: Searching to increasing depths to ensure the best move is always available within time limits.
- **Quiescence Search**: Extending search at leaf nodes to avoid the horizon effect in volatile positions.
- **Move Ordering (Captures)**: MVV-LVA (Most Valuable Victim - Least Valuable Aggressor) heuristic to prioritize captures. Quiet move ordering is still pending.
- **Bitboards**: Full bitboard implementation using Magic Bitboards for sliding pieces and pre-calculated attack tables.
- **Evaluation**: Piece-Square Tables (PST) to encourage positional play (center control, king safety).

## Next Steps

**Future Improvements (Phase 6):**

- [ ] Move ordering for quiet moves
- [ ] Advanced Parallel Search (Lazy SMP)
- [ ] Tapered Evaluation (Variable King PST based on game phase)
- [ ] Opening Books
- [ ] Endgame Tablebases
- [ ] Killer Moves
- [ ] History Heuristic
- [ ] Check Extensions

## License

MIT License
