package main

import (
	"arminia-chess-engine/internal/debug"
	"arminia-chess-engine/internal/uci"
	"log/slog"
	"os"
)

func main() {
	defer debug.Recover("Main Loop")
	// Setup logging to file
	logFile, err := os.OpenFile("arminia.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger := slog.New(slog.NewTextHandler(logFile, nil))
		slog.SetDefault(logger)
	}

	slog.Info("Arminia UCI Engine started")

	protocol := uci.NewProtocol(os.Stdin, os.Stdout)
	if err := protocol.Run(); err != nil && err.Error() != "EOF" {
		slog.Error("UCI Protocol error", "error", err)
		os.Exit(1)
	}

	slog.Info("Arminia UCI Engine stopped")
}
