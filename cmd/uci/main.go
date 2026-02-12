package main

import (
	"arminia-chess-engine/internal/uci"
	"os"
)

func main() {
	protocol := uci.NewProtocol(os.Stdin, os.Stdout)
	if err := protocol.Run(); err != nil && err.Error() != "EOF" {
		os.Exit(1)
	}
}
