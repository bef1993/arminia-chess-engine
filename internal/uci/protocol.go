package uci

import (
	"arminia-chess-engine/internal/engine"
	"arminia-chess-engine/internal/search"
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// DefaultDepth is the fixed search depth used by the engine
const DefaultDepth = 5

// Protocol represents the UCI protocol handler
type Protocol struct {
	input  io.Reader
	output io.Writer
	game   *engine.Game
}

// NewProtocol creates a new UCI protocol handler
func NewProtocol(input io.Reader, output io.Writer) *Protocol {
	return &Protocol{
		input:  input,
		output: output,
		game:   engine.NewGame(),
	}
}

// Run starts the UCI protocol loop
func (u *Protocol) Run() error {
	scanner := bufio.NewScanner(u.input)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		slog.Info("UCI Input", "command", line)

		if err := u.handleCommand(line); err != nil {
			if err == io.EOF {
				return nil
			}
			slog.Error("UCI Error", "error", err)
			return err
		}
	}

	return scanner.Err()
}

// handleCommand processes a single UCI command
func (u *Protocol) handleCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "uci":
		return u.handleUCI()
	case "isready":
		return u.handleIsReady()
	case "setoption":
		return u.handleSetOption(parts[1:])
	case "ucinewgame":
		return u.handleUCINewGame()
	case "position":
		return u.handlePosition(parts[1:])
	case "go":
		return u.handleGo(parts[1:])
	case "quit":
		return io.EOF // Signal to exit
	default:
		slog.Warn("Unknown command", "command", command)
		return u.writeLine(fmt.Sprintf("info string Unknown command: %s", command))
	}
}

// handleUCI sends engine identification
func (u *Protocol) handleUCI() error {
	if err := u.writeLine("id name Arminia"); err != nil {
		return err
	}
	if err := u.writeLine("id author Stefan Wilfinger"); err != nil {
		return err
	}
	if err := u.writeLine(fmt.Sprintf("option name Hash type spin default %d min 1 max 1024", search.DefaultTTSizeMB)); err != nil {
		return err
	}
	if err := u.writeLine("option name Threads type spin default 1 min 1 max 32"); err != nil {
		return err
	}
	if err := u.writeLine("option name Move Overhead type spin default 10 min 0 max 5000"); err != nil {
		return err
	}
	if err := u.writeLine("option name SyzygyPath type string default <empty>"); err != nil {
		return err
	}
	if err := u.writeLine("option name UCI_ShowWDL type check default false"); err != nil {
		return err
	}
	return u.writeLine("uciok")
}

// handleIsReady responds when ready
func (u *Protocol) handleIsReady() error {
	return u.writeLine("readyok")
}

// handleSetOption processes option setting (currently no-op)
func (u *Protocol) handleSetOption(args []string) error {
	// Parse: setoption name <id> value <x>
	slog.Info("SetOption", "args", args)

	if len(args) >= 4 && args[0] == "name" && args[1] == "Hash" && args[2] == "value" {
		sizeMB, err := strconv.Atoi(args[3])
		if err == nil {
			slog.Info("Resizing Hash", "sizeMB", sizeMB)
			search.GlobalTT.Resize(sizeMB)
		}
	}

	return nil
}

// handleUCINewGame resets for a new game
func (u *Protocol) handleUCINewGame() error {
	slog.Info("New Game")
	u.game = engine.NewGame()
	return nil
}

// handlePosition sets the board position
func (u *Protocol) handlePosition(args []string) error {
	if len(args) == 0 {
		return u.writeLine("info string position command requires arguments")
	}

	if args[0] == "startpos" {
		u.game = engine.NewGame()
		// Parse moves if provided
		if len(args) > 1 && args[1] == "moves" {
			return u.applyMoves(args[2:])
		}
		return nil
	} else if args[0] == "fen" {
		// Find where "moves" starts, if any
		movesIndex := -1
		for i, arg := range args {
			if arg == "moves" {
				movesIndex = i
				break
			}
		}

		var fenString string
		var moveArgs []string

		if movesIndex != -1 {
			fenString = strings.Join(args[1:movesIndex], " ")
			moveArgs = args[movesIndex+1:]
		} else {
			fenString = strings.Join(args[1:], " ")
		}

		if err := u.game.LoadFEN(fenString); err != nil {
			slog.Error("Invalid FEN", "fen", fenString, "error", err)
			return u.writeLine(fmt.Sprintf("info string Invalid FEN: %v", err))
		}

		if len(moveArgs) > 0 {
			return u.applyMoves(moveArgs)
		}

		return nil
	}

	return u.writeLine("info string position command error")
}

// applyMoves applies a sequence of moves to the current position
func (u *Protocol) applyMoves(moveStrings []string) error {
	for _, moveStr := range moveStrings {
		move, err := engine.ParseMove(moveStr, u.game)
		if err != nil {
			slog.Error("Illegal move", "move", moveStr, "error", err)
			return u.writeLine(fmt.Sprintf("info string Illegal move: %v", err))
		}
		u.game.ExecuteMove(move)
	}
	return nil
}

// SearchLimits holds the parsed parameters from the "go" command
type SearchLimits struct {
	WhiteTime      int // wtime
	BlackTime      int // btime
	WhiteIncrement int // winc
	BlackIncrement int // binc
	MovesToGo      int // movestogo
	Depth          int // depth
	Nodes          int // nodes
	Mate           int // mate
	MoveTime       int // movetime
	Infinite       bool
}

// handleGo starts search for best move
func (u *Protocol) handleGo(args []string) error {
	limits := SearchLimits{}

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "wtime":
			if i+1 < len(args) {
				limits.WhiteTime, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "btime":
			if i+1 < len(args) {
				limits.BlackTime, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "winc":
			if i+1 < len(args) {
				limits.WhiteIncrement, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "binc":
			if i+1 < len(args) {
				limits.BlackIncrement, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "movestogo":
			if i+1 < len(args) {
				limits.MovesToGo, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "depth":
			if i+1 < len(args) {
				limits.Depth, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "nodes":
			if i+1 < len(args) {
				limits.Nodes, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "mate":
			if i+1 < len(args) {
				limits.Mate, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "movetime":
			if i+1 < len(args) {
				limits.MoveTime, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "infinite":
			limits.Infinite = true
		}
	}

	slog.Info("Search started", "limits", limits)

	start := time.Now()

	// Call the search package
	// TODO: Pass limits to search
	move, score, nodes := search.Search(u.game, DefaultDepth)

	duration := time.Since(start)
	ms := duration.Milliseconds()
	nps := 0
	if duration.Seconds() > 0 {
		nps = int(float64(nodes) / duration.Seconds())
	}

	if (move == engine.Move{}) {
		// No legal moves available (Checkmate or Stalemate)
		return u.writeLine("bestmove (none)")
	}

	// Format score (cp or mate)
	scoreStr := fmt.Sprintf("cp %d", score)
	if score > search.MateBound {
		movesToMate := (search.EvalMate - score + 1) / 2
		scoreStr = fmt.Sprintf("mate %d", movesToMate)
	} else if score < -search.MateBound {
		movesToMate := -(search.EvalMate + score + 1) / 2
		scoreStr = fmt.Sprintf("mate %d", movesToMate)
	}

	// Report search statistics
	infoStr := fmt.Sprintf("info depth %d score %s nodes %d nps %d time %d pv %s", DefaultDepth, scoreStr, nodes, nps, ms, move.String())
	slog.Info("Search stats", "depth", DefaultDepth, "score", scoreStr, "nodes", nodes, "nps", nps, "time_ms", ms, "pv", move.String())
	u.writeLine(infoStr)

	slog.Info("Best move found", "move", move.String())
	return u.writeLine(fmt.Sprintf("bestmove %s", move.String()))
}

// writeLine writes a line to output
func (u *Protocol) writeLine(text string) error {
	slog.Info("UCI Output", "response", text)
	_, err := fmt.Fprintln(u.output, text)
	return err
}
