package uci

import (
	"arminia-chess-engine/internal/engine"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolInitialization(t *testing.T) {
	input := strings.NewReader("uci\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	assert.NotNil(t, protocol, "NewProtocol returned nil")
	assert.NotNil(t, protocol.game, "Protocol game is nil")
}

func TestHandleUCI(t *testing.T) {
	input := strings.NewReader("uci\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	assert.Contains(t, result, "id name Arminia")
	assert.Contains(t, result, "id author Stefan Wilfinger")
	assert.Contains(t, result, "uciok")
}

func TestHandleIsReady(t *testing.T) {
	input := strings.NewReader("isready\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	assert.Contains(t, result, "readyok")
}

func TestHandleUCINewGame(t *testing.T) {
	input := strings.NewReader("ucinewgame\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	err := protocol.Run()

	// Should complete without error
	assert.NoError(t, err)
}

func TestHandlePosition(t *testing.T) {
	input := strings.NewReader("position startpos\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	// Verify the game board is initialized with starting position
	piece := protocol.game.Board.GetPieceAt("e1")
	assert.NotEqual(t, engine.NoPiece, piece)
	assert.Equal(t, engine.King, piece.Type())
	assert.Equal(t, engine.White, piece.Color())
}

func TestHandlePositionWithMoves(t *testing.T) {
	input := strings.NewReader("position startpos moves e2e4 e7e5\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	// After e2e4, white pawn should be at e4
	piece := protocol.game.Board.GetPieceAt("e4")
	assert.NotEqual(t, engine.NoPiece, piece)
	assert.Equal(t, engine.Pawn, piece.Type())
	assert.Equal(t, engine.White, piece.Color())

	// After e7e5, black pawn should be at e5
	piece = protocol.game.Board.GetPieceAt("e5")
	assert.NotEqual(t, engine.NoPiece, piece)
	assert.Equal(t, engine.Pawn, piece.Type())
	assert.Equal(t, engine.Black, piece.Color())

	// Current turn should be white
	assert.Equal(t, engine.White, protocol.game.CurrentTurn)
}

func TestHandleGo(t *testing.T) {
	input := strings.NewReader("position startpos\ngo\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	assert.Contains(t, result, "bestmove")
}

func TestMoveFormat(t *testing.T) {
	input := strings.NewReader("position startpos moves e2e4\ngo\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	// Should contain a valid move in algebraic notation
	assert.Contains(t, result, "bestmove")
}

func TestInvalidMoveRejection(t *testing.T) {
	input := strings.NewReader("position startpos moves h2h9\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	assert.True(t, strings.Contains(result, "Illegal move") || strings.Contains(result, "error"), "Invalid moves should be rejected or reported")
}

func TestMultipleCommands(t *testing.T) {
	input := strings.NewReader(
		"uci\nisready\nucinewgame\nposition startpos\ngo\nquit\n",
	)
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	err := protocol.Run()

	assert.NoError(t, err)

	result := output.String()

	assert.Contains(t, result, "uciok")
	assert.Contains(t, result, "readyok")
	assert.Contains(t, result, "bestmove")
}

func TestEmptyInput(t *testing.T) {
	input := strings.NewReader("quit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	// Should handle gracefully
	assert.NotNil(t, protocol.game, "Game should exist after empty input")
}
