package uci

import (
	"arminia-chess-engine/internal/engine"
	"bytes"
	"strings"
	"testing"
	"time"

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
	input := strings.NewReader("position startpos\ngo depth 1\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)
	protocol.Run()

	result := output.String()

	assert.Contains(t, result, "bestmove")
}

func TestMoveFormat(t *testing.T) {
	input := strings.NewReader("position startpos moves e2e4\ngo depth 1\nquit\n")
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
		"uci\nisready\nucinewgame\nposition startpos\ngo depth 1\nquit\n",
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

func TestParseSearchLimits(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected SearchLimits
	}{
		{
			name: "Standard Time Control",
			args: []string{"wtime", "60000", "btime", "50000", "winc", "1000", "binc", "1000"},
			expected: SearchLimits{
				WhiteTime:      60000,
				BlackTime:      50000,
				WhiteIncrement: 1000,
				BlackIncrement: 1000,
			},
		},
		{
			name: "Move Time",
			args: []string{"movetime", "5000"},
			expected: SearchLimits{
				MoveTime: 5000,
			},
		},
		{
			name: "Depth and Nodes",
			args: []string{"depth", "10", "nodes", "100000"},
			expected: SearchLimits{
				Depth: 10,
				Nodes: 100000,
			},
		},
		{
			name: "Infinite",
			args: []string{"infinite"},
			expected: SearchLimits{
				Infinite: true,
			},
		},
		{
			name: "Mate",
			args: []string{"mate", "5"},
			expected: SearchLimits{
				Mate: 5,
			},
		},
		{
			name: "Moves To Go",
			args: []string{"movestogo", "40"},
			expected: SearchLimits{
				MovesToGo: 40,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSearchLimits(tt.args)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestHandleStop(t *testing.T) {
	// Simulate a search that is stopped manually
	input := strings.NewReader("position startpos\ngo infinite\nstop\nquit\n")
	output := &bytes.Buffer{}

	protocol := NewProtocol(input, output)

	// Run the protocol loop
	err := protocol.Run()
	assert.NoError(t, err)

	result := output.String()

	// Verify that a bestmove was sent (response to stop)
	assert.Contains(t, result, "bestmove", "Engine should output bestmove after stop command")
}

func TestCalculateTimeLimit(t *testing.T) {
	p := NewProtocol(nil, nil) // Input/Output not needed for this calculation

	tests := []struct {
		name     string
		turn     engine.Color
		limits   SearchLimits
		expected time.Duration
	}{
		{
			name: "Fixed Move Time",
			turn: engine.White,
			limits: SearchLimits{
				MoveTime: 5000,
			},
			expected: 5 * time.Second,
		},
		{
			name: "Infinite",
			turn: engine.White,
			limits: SearchLimits{
				Infinite: true,
			},
			expected: 0,
		},
		{
			name: "White Time Control (Default MovesToGo)",
			turn: engine.White,
			limits: SearchLimits{
				WhiteTime: 60000, // 60s
			},
			// 60000 / 20 = 3000ms
			expected: 3 * time.Second,
		},
		{
			name: "Black Time Control (Default MovesToGo)",
			turn: engine.Black,
			limits: SearchLimits{
				BlackTime: 40000, // 40s
			},
			// 40000 / 20 = 2000ms
			expected: 2 * time.Second,
		},
		{
			name: "Time Control with Increment",
			turn: engine.White,
			limits: SearchLimits{
				WhiteTime:      60000,
				WhiteIncrement: 2000,
			},
			// 60000/20 + 2000/2 = 3000 + 1000 = 4000ms
			expected: 4 * time.Second,
		},
		{
			name: "Time Control with Explicit MovesToGo",
			turn: engine.White,
			limits: SearchLimits{
				WhiteTime: 60000,
				MovesToGo: 30,
			},
			// 60000 / 30 = 2000ms
			expected: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.game.CurrentTurn = tt.turn
			got := p.calculateTimeLimit(tt.limits)
			assert.Equal(t, tt.expected, got)
		})
	}
}
