package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httphandler "github.com/autoparts/backend-test/internal/http"
	"github.com/autoparts/backend-test/internal/repository/memory"
	"github.com/autoparts/backend-test/internal/repository/sqlite"
	"github.com/autoparts/backend-test/internal/usecase"
)

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "parts.db"
	}

	var partUseCase *usecase.PartUseCase

	if os.Getenv("IN_MEMORY") == "true" {
		slog.Info("using in-memory repository")
		partUseCase = usecase.NewPartUseCase(memory.New())
	} else {
		slog.Info("using sqlite repository", "path", dbPath)
		sqliteRepo, err := sqlite.New(dbPath)
		if err != nil {
			slog.Error("failed to create sqlite repo", "error", err)
			os.Exit(1)
		}
		partUseCase = usecase.NewPartUseCase(sqliteRepo)
	}

	handler := httphandler.NewHandler(partUseCase)
	router := httphandler.NewRouter(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	readTimeout := time.Duration(getEnvInt("READ_TIMEOUT", 15)) * time.Second
	writeTimeout := time.Duration(getEnvInt("WRITE_TIMEOUT", 15)) * time.Second
	idleTimeout := time.Duration(getEnvInt("IDLE_TIMEOUT", 60)) * time.Second

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
