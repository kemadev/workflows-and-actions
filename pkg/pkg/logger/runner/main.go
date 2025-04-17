package runner

import (
	"log/slog"
	"os"
)

func init() {
	var logLevel slog.Level
	if os.Getenv("RUNNER_DEBUG") == "1" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}
