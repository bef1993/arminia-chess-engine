package main

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUCIApplication(t *testing.T) {
	// Capture stdout and stdin to verify main() behavior
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	// Create pipes for stdin/stdout
	rIn, wIn, err := os.Pipe()
	assert.NoError(t, err)

	rOut, wOut, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdin = rIn
	os.Stdout = wOut

	// Write commands to stdin in a goroutine
	go func() {
		// Send UCI command to verify output, then quit to exit main
		// We need "quit" to ensure main() returns and doesn't block forever
		_, _ = wIn.Write([]byte("uci\nquit\n"))
		_ = wIn.Close()
	}()

	// Run main() in a goroutine to prevent blocking the test runner
	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	// Wait for main to finish with a timeout
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for main() to finish")
	}

	// Close writer to read captured output
	_ = wOut.Close()

	outputBytes, err := io.ReadAll(rOut)
	assert.NoError(t, err)
	output := string(outputBytes)

	// Verify expected UCI output
	assert.Contains(t, output, "id name Arminia")
	assert.Contains(t, output, "uciok")

	// Cleanup the log file created by main
	_ = os.Remove("arminia.log")
}
