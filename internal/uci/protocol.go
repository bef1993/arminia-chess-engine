package uci

import (
	"arminia-chess-engine/internal/engine"
	"arminia-chess-engine/internal/search"
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultThreads = 8

// Protocol represents the UCI protocol handler
type Protocol struct {
	input        io.Reader
	output       io.Writer
	game         *engine.Game
	cancelSearch context.CancelFunc
	searchWg     sync.WaitGroup
	threads      int
}

// NewProtocol creates a new UCI protocol handler
func NewProtocol(input io.Reader, output io.Writer) *Protocol {
	return &Protocol{
		input:   input,
		output:  output,
		game:    engine.NewGame(),
		threads: defaultThreads,
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
	case "stop":
		return u.handleStop()
	case "quit":
		if u.cancelSearch != nil {
			u.cancelSearch()
		}
		u.searchWg.Wait()
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
	if err := u.sendUCIOptions(); err != nil {
		return err
	}
	return u.writeLine("uciok")
}

// sendUCIOptions sends available UCI options
func (u *Protocol) sendUCIOptions() error {
	options := []string{
		fmt.Sprintf("option name Hash type spin default %d min 1 max 1024", search.DefaultTTSizeMB),
		"option name Threads type spin default 1 min 1 max 32",
		"option name Move Overhead type spin default 10 min 0 max 5000",
		"option name SyzygyPath type string default <empty>",
		"option name UCI_ShowWDL type check default false",
	}

	for _, opt := range options {
		if err := u.writeLine(opt); err != nil {
			return err
		}
	}
	return nil
}

// handleIsReady responds when ready
func (u *Protocol) handleIsReady() error {
	return u.writeLine("readyok")
}

// handleSetOption processes option setting
func (u *Protocol) handleSetOption(args []string) error {
	// Parse: setoption name <id> value <x>
	slog.Info("SetOption", "args", args)

	var name, value string
	for i := 0; i < len(args); i++ {
		if args[i] == "name" && i+1 < len(args) {
			name = args[i+1]
		}
		if args[i] == "value" && i+1 < len(args) {
			value = args[i+1]
		}
	}

	if strings.EqualFold(name, "Hash") {
		sizeMB, err := strconv.Atoi(value)
		if err == nil {
			u.ensureSearchStopped()
			slog.Info("Resizing Hash", "sizeMB", sizeMB)
			search.GlobalTT.Resize(sizeMB)
		}
	} else if strings.EqualFold(name, "Threads") {
		threads, err := strconv.Atoi(value)
		if err == nil {
			u.ensureSearchStopped()
			slog.Info("Setting Threads", "count", threads)
			u.threads = threads
		}
	}

	return nil
}

// handleUCINewGame resets for a new game
func (u *Protocol) handleUCINewGame() error {
	u.ensureSearchStopped()
	slog.Info("New Game")
	search.GlobalTT.Clear()
	u.game = engine.NewGame()
	return nil
}

// handlePosition sets the board position
func (u *Protocol) handlePosition(args []string) error {
	u.ensureSearchStopped()

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

// handleStop stops the current search
func (u *Protocol) handleStop() error {
	if u.cancelSearch != nil {
		u.cancelSearch()
	}
	return nil
}

// ensureSearchStopped cancels the current search and waits for it to finish
func (u *Protocol) ensureSearchStopped() {
	if u.cancelSearch != nil {
		u.cancelSearch()
	}
	u.searchWg.Wait()
}

// handleGo starts search for best move
func (u *Protocol) handleGo(args []string) error {
	limits := parseSearchLimits(args)

	slog.Info("Search started", "limits", limits)

	// Ensure previous search is stopped before starting a new one
	u.ensureSearchStopped()

	// Determine search options
	options := search.SearchOptions{
		MaxDepth: limits.Depth,
		Threads:  u.threads,
	}

	// If no depth specified, use a high default (effectively infinite with time limit)
	if options.MaxDepth == 0 {
		options.MaxDepth = 100
	}

	duration := u.calculateTimeLimit(limits)

	return u.runSearch(options, duration)
}

// parseSearchLimits parses the arguments for the go command
func parseSearchLimits(args []string) SearchLimits {
	limits := SearchLimits{}
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
	return limits
}

// calculateTimeLimit determines the time duration for the search
func (u *Protocol) calculateTimeLimit(limits SearchLimits) time.Duration {
	if limits.MoveTime > 0 {
		return time.Duration(limits.MoveTime) * time.Millisecond
	}
	if limits.Infinite {
		return 0
	}

	var timeAvailable, increment int
	if u.game.CurrentTurn == engine.White {
		timeAvailable = limits.WhiteTime
		increment = limits.WhiteIncrement
	} else {
		timeAvailable = limits.BlackTime
		increment = limits.BlackIncrement
	}

	if timeAvailable > 0 {
		// Strategy: Use 1/20th of remaining time + increment/2
		// This is a simple but effective strategy for blitz/rapid
		movesToGo := 20
		if limits.MovesToGo > 0 {
			movesToGo = limits.MovesToGo
		}
		return time.Duration(timeAvailable/movesToGo+increment/2) * time.Millisecond
	}

	return 0
}

// runSearch executes the search in a separate goroutine
func (u *Protocol) runSearch(options search.SearchOptions, duration time.Duration) error {
	var ctx context.Context
	var cancel context.CancelFunc
	if duration > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), duration)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	u.cancelSearch = cancel

	// Channel for search updates
	// Buffer slightly to prevent search from blocking on I/O
	infoCh := make(chan search.SearchInfo, 32)
	consumerDone := make(chan struct{})

	// Goroutine 1: Consumer (UCI Output)
	u.searchWg.Go(func() {
		defer close(consumerDone)
		start := time.Now()
		for info := range infoCh {
			u.sendSearchInfo(info, start)
		}
	})

	// Goroutine 2: Producer (Search Execution)
	u.searchWg.Go(func() {
		defer cancel() // Ensure resources are released

		move, _, _ := search.Search(ctx, u.game, options, infoCh)
		close(infoCh) // Close channel so consumer can finish

		// Wait for consumer to finish printing all info lines
		<-consumerDone

		if (move == engine.Move{}) {
			u.writeLine("bestmove (none)")
		} else {
			slog.Info("Best move found", "move", move.String())
			u.writeLine(fmt.Sprintf("bestmove %s", move.String()))
		}
	})

	return nil
}

// sendSearchInfo formats and sends search progress information
func (u *Protocol) sendSearchInfo(info search.SearchInfo, start time.Time) {
	elapsed := time.Since(start)
	ms := elapsed.Milliseconds()
	var nps int64
	if elapsed.Seconds() > 0 {
		nps = int64(float64(info.Nodes) / elapsed.Seconds())
	}

	// Format score (cp or mate)
	scoreStr := fmt.Sprintf("cp %d", info.Score)
	if info.Score > search.MateBound {
		movesToMate := (search.EvalMate - info.Score + 1) / 2
		scoreStr = fmt.Sprintf("mate %d", movesToMate)
	} else if info.Score < -search.MateBound {
		movesToMate := -(search.EvalMate + info.Score + 1) / 2
		scoreStr = fmt.Sprintf("mate %d", movesToMate)
	}

	// Format PV
	pvStr := ""
	for _, m := range info.PV {
		pvStr += m.String() + " "
	}

	infoStr := fmt.Sprintf("info depth %d seldepth %d score %s nodes %d nps %d time %d pv %s", info.Depth, info.SelDepth, scoreStr, info.Nodes, nps, ms, pvStr)
	u.writeLine(infoStr)
}

// writeLine writes a line to output
func (u *Protocol) writeLine(text string) error {
	slog.Info("UCI Output", "response", text)
	_, err := fmt.Fprintln(u.output, text)
	return err
}
