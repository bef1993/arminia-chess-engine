package main

import (
	"arminia-chess-engine/internal/engine"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	game := engine.NewGame()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Arminia Chess Engine - Interactive CLI")
	fmt.Println("Commands: move <from> <to> (e.g., e2 e4), board, quit")
	fmt.Println()

	for {
		game.PrintBoard()

		if game.IsCheckmate() {
			fmt.Println("Checkmate! Game Over.")
			return
		}
		if game.IsStalemate() {
			fmt.Println("Stalemate! Game Over.")
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
			if len(parts) < 3 {
				fmt.Println("Usage: move <from> <to> (e.g., e2 e4)")
				continue
			}
			if err := handleMove(game, parts[1], parts[2]); err != nil {
				fmt.Println("Error:", err)
			}
		default:
			fmt.Println("Unknown command. Try 'move', 'board', or 'quit'.")
		}
	}
}

func handleMove(game *engine.Game, fromStr, toStr string) error {
	if len(fromStr) != 2 || len(toStr) != 2 {
		return fmt.Errorf("invalid square format")
	}

	fromCol := int(fromStr[0] - 'a')
	fromRow := 8 - int(fromStr[1]-'0')
	toCol := int(toStr[0] - 'a')
	toRow := 8 - int(toStr[1]-'0')

	if fromCol < 0 || fromCol > 7 || fromRow < 0 || fromRow > 7 ||
		toCol < 0 || toCol > 7 || toRow < 0 || toRow > 7 {
		return fmt.Errorf("square out of bounds")
	}

	moves := game.GetLegalMoves()
	for _, move := range moves {
		if move.FromCol == fromCol && move.FromRow == fromRow &&
			move.ToCol == toCol && move.ToRow == toRow {
			game.ExecuteMove(move)
			return nil
		}
	}

	return fmt.Errorf("illegal move")
}
