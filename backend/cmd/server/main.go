// Command server is the composition root: it loads config, wires concrete
// implementations together, and runs the HTTP server.
package main

import (
	"log/slog"
	"os"

	"github.com/parse-companies/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}

	logger.Info("starting", "addr", cfg.HTTPAddr)
	// HTTP server wiring lands in Phase 4 (Task 13).
}
