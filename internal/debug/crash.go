package debug

import (
	"log/slog"
	"os"
	"runtime/debug"
)

// LogFile is the path to the crash log file.
const LogFile = "arminia-crash.log"

// Recover is a helper to catch panics and log them to a file.
// It should be deferred in the main function and/or the search loop.
// contextData is optional information to log (e.g., game state).
func Recover(contextData ...any) {
	if r := recover(); r != nil {
		// Open the log file in append mode
		f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// If we can't open the file, try to print to stderr (though UCI might swallow it)
			slog.Error("CRITICAL: Panic caught but failed to open crash log file", "error", err, "panic", r)
			os.Exit(1)
		}
		defer f.Close()

		// Create a dedicated logger for the crash file
		logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
			AddSource: true,
		}))

		logger.Error("CRASH REPORT",
			"panic", r,
			"context", contextData,
			"stack", string(debug.Stack()),
		)

		// Exit with error code 1 to indicate failure
		os.Exit(1)
	}
}
