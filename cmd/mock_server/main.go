// Service for Sreda Talents.
//
// # Описание сервиса, реализующего функционал приема запросов.
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// Schemes: http, https
// Host: localhost:8091
// Version: 1.0.0
//
// swagger:meta
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sreda/internal/api"
	"sreda/internal/config"
	"sreda/internal/lib/logger/sl"

	"github.com/gorilla/mux"
)

func main() {
	port := config.GetMockPort()
	env := config.GetMockEnv()
	log := sl.SetupLogger(env)
	log.Info(
		"starting mock server",
		slog.String("env", env),
		slog.String("port", port),
	)

	if err := run(log, port); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}
}

func run(log *slog.Logger, port string) error {
	log.Debug("starting mock server")

	router := mux.NewRouter()

	router.HandleFunc("/api/request", api.ProcessRequest(log)).Methods("POST")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cfgAddress := fmt.Sprintf(":%s", port)

	srv := &http.Server{
		Addr:    cfgAddress,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start mock server", "error", err)
			}
		}
	}()

	log.Info("started mock server", "port", cfgAddress)

	<-done
	log.Info("stopping mock server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop mock server", "error", err)

		return err
	}

	log.Info("mock server stopped")

	return nil
}
