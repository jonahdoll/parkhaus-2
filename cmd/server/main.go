package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"parkhaus-2/internal/config"
	"parkhaus-2/internal/db"
	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/server"
)

func main() {
	model.InitDecimalMarshal()

	cfg := config.Load()

	// Strukturiertes Logging via slog.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	// Datenbankverbindung (DDL via Init-Skripte, kein AutoMigrate).
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Start fehlgeschlagen", "fehler", err)
		os.Exit(1)
	}
	defer db.Close(database)

	router := server.New(database)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Server in Goroutine starten.
	go func() {
		slog.Info("Server gestartet", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server-Fehler", "fehler", err)
			os.Exit(1)
		}
	}()

	// Auf SIGINT/SIGTERM warten (Graceful Shutdown).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown wird eingeleitet ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Shutdown-Fehler", "fehler", err)
	}
	slog.Info("Server beendet")
}
