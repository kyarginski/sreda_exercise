// Service for Sreda Talents.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"sreda/internal/config"
	"sreda/internal/lib/logger/sl"
	"sreda/internal/processes"
)

func main() {
	cfg := config.MustLoad()
	log := sl.SetupLogger(cfg.Env)
	log.Info(
		"starting sender service",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
		slog.String("URL", cfg.URL),
	)

	if err := run(log, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}
}

func run(log *slog.Logger, cfg *config.Config) error {
	startTime := time.Now()
	log.Info("started sender service")
	sender := processes.NewSender(log, cfg)
	ctx := context.Background()
	err := sender.Run(ctx)
	finishTime := time.Since(startTime).String()
	log.Info("finished sender service",
		"time_elapsed", finishTime,
	)

	return err
}
