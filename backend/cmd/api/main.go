package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tianmuwu/short.ly/internal/httpapi"
	"github.com/tianmuwu/short.ly/internal/storage"
)

func main() {
	_ = godotenv.Load(".env", "../.env")

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpAddr := getenv("HTTP_ADDR", "0.0.0.0:8080")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Fatalf("DATABASE_URL is required")
	}

	repo, err := storage.OpenAndMigrate(databaseURL)
	if err != nil {
		logger.Fatalf("db: %v", err)
	}

	srv := httpapi.NewServer(httpapi.ServerConfig{
		HTTPAddr: httpAddr,
		Logger:   logger,
		Repo:     repo,
	})

	go func() {
		logger.Printf("listening on %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http serve: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("http shutdown: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
