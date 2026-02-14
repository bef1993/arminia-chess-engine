package main

import (
	"arminia-chess-engine/internal/engine"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	game := engine.NewGame()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Arminia Chess Engine - Interactive CLI")
	fmt.Println("Commands: move <uci_move> (e.g., e2e4, e7e8q), board, quit")
	fmt.Println()

	for {
		PrintBoard(os.Stdout, game)

		status := game.GetGameStatus()
		if status != engine.StatusActive {
			switch status {
			case engine.StatusCheckmate:
				fmt.Println("Checkmate! Game Over.")
			case engine.StatusStalemate:
				fmt.Println("Stalemate! Game Over.")
			case engine.StatusDraw50Move:
				fmt.Println("Draw by 50-move rule! Game Over.")
			case engine.StatusDrawRepetition:
				fmt.Println("Draw by threefold repetition! Game Over.")
			case engine.StatusDrawInsufficientMaterial:
				fmt.Println("Draw by insufficient material! Game Over.")
			}
			return
		}

		color := "White"
		if game.CurrentTurn == engine.Black {
			color = "Black"
		}

		if game.Board.IsKingInCheck(game.CurrentTurn) {
			fmt.Printf("Check! ")
		}

		fmt.Printf("\n%s to move > ", color)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)

		switch parts[0] {
		case "quit":
			fmt.Println("Goodbye!")
			return
		case "board":
			continue
		case "move":
			if len(parts) < 2 {
				fmt.Println("Usage: move <uci_move> (e.g., e2e4 or e7e8q)")
				continue
			}

			if err := handleMove(game, parts[1]); err != nil {
				fmt.Println("Error:", err)
			}
		default:
			fmt.Println("Unknown command. Try 'move', 'board', or 'quit'.")
		}
	}
}

func handleMove(game *engine.Game, moveStr string) error {
	move, err := engine.ParseMove(moveStr, game)
	if err != nil {
		return err
	}

	game.ExecuteMove(move)
	return nil
}

// PrintBoard prints the current board state to the console
func PrintBoard(w io.Writer, g *engine.Game) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintln(w, "  a b c d e f g h")
	fmt.Fprintln(w, "  ╔═╦═╦═╦═╦═╦═╦═╦═╗")

	for row := 0; row < 8; row++ {
		fmt.Fprint(w, 8-row)
		fmt.Fprint(w, " ")
		fmt.Fprint(w, "║")

		for col := 0; col < 8; col++ {
			piece := g.Board.GetPiece(col, row)
			if piece != engine.NoPiece {
				fmt.Fprint(w, piece.GetSymbol())
			} else {
				fmt.Fprint(w, " ")
			}

			if col < 7 {
				fmt.Fprint(w, "║")
			} else {
				fmt.Fprint(w, "║")
			}
		}

		fmt.Fprint(w, " ")
		fmt.Fprint(w, 8-row)
		if row < 7 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "  ╠═╬═╬═╬═╬═╬═╬═╬═╣")
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "  ╚═╩═╩═╩═╩═╩═╩═╩═╝")
	fmt.Fprintln(w, "  a b c d e f g h")
	fmt.Fprintln(w)

	if g.CurrentTurn == engine.White {
		fmt.Fprintln(w, "Current turn: White")
	} else {
		fmt.Fprintln(w, "Current turn: Black")
	}
}
