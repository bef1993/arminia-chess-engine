package uci

import (
	"arminia-chess-engine/internal/engine"
	"bufio"
	"fmt"
	"io"
	"strings"
)

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

		if err := u.handleCommand(line); err != nil {
			if err == io.EOF {
				return nil
			}
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
	if err := u.writeLine("option name Hash type spin default 16 min 1 max 512"); err != nil {
		return err
	}
	if err := u.writeLine("option name Threads type spin default 1 min 1 max 32"); err != nil {
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
	// For now, we ignore options but accept them gracefully
	return nil
}

// handleUCINewGame resets for a new game
func (u *Protocol) handleUCINewGame() error {
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
		if len(moveStr) < 4 {
			return u.writeLine(fmt.Sprintf("info string Invalid move format: %s", moveStr))
		}

		// Parse move: format is e2e4 or e7e8q (with promotion)
		fromCol := int(moveStr[0] - 'a')
		fromRow := 8 - int(moveStr[1]-'0')
		toCol := int(moveStr[2] - 'a')
		toRow := 8 - int(moveStr[3]-'0')

		var promotionPiece engine.Piece
		if len(moveStr) == 5 {
			switch moveStr[4] {
			case 'q':
				promotionPiece = engine.Queen.FromColor(u.game.CurrentTurn)
			case 'r':
				promotionPiece = engine.Rook.FromColor(u.game.CurrentTurn)
			case 'b':
				promotionPiece = engine.Bishop.FromColor(u.game.CurrentTurn)
			case 'n':
				promotionPiece = engine.Knight.FromColor(u.game.CurrentTurn)
			}

		}

		// Check if move exists in legal moves
		moves := u.game.GetLegalMoves()
		found := false

		for _, move := range moves {
			if move.FromCol == fromCol && move.FromRow == fromRow &&
				move.ToCol == toCol && move.ToRow == toRow &&
				move.PromotionPiece == promotionPiece {
				found = true
				u.game.ExecuteMove(move)
				break
			}
		}

		if !found {
			return u.writeLine(fmt.Sprintf("info string Illegal move: %s", moveStr))
		}
	}

	return nil
}

// handleGo starts search for best move
func (u *Protocol) handleGo(args []string) error {
	// For now, just return a random legal move
	moves := u.game.GetLegalMoves()

	if len(moves) == 0 {
		return u.writeLine("bestmove 0000")
	}

	// Select first legal move for now (TODO: implement proper search)
	move := moves[0]
	moveStr := fmt.Sprintf("%c%d%c%d",
		rune('a'+move.FromCol),
		8-move.FromRow,
		rune('a'+move.ToCol),
		8-move.ToRow)

	// Handle promotion in output
	if move.PromotionPiece != engine.NoPiece {
		switch move.PromotionPiece.Type() {
		case engine.Queen:
			moveStr += "q"
		case engine.Rook:
			moveStr += "r"
		case engine.Bishop:
			moveStr += "b"
		case engine.Knight:
			moveStr += "n"
		}
	}

	return u.writeLine(fmt.Sprintf("bestmove %s", moveStr))
}

// writeLine writes a line to output
func (u *Protocol) writeLine(text string) error {
	_, err := fmt.Fprintln(u.output, text)
	return err
}
