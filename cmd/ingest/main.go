package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/platform"
)

func main() {
	config := flag.String("config", "coins.yaml", "coin configuration")
	flag.Parse()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()
	c, err := platform.Connect(ctx)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer c.Close(ctx)
	base := os.Getenv("COINGECKO_API_URL")
	if base == "" {
		base = "https://api.coingecko.com/api/v3"
	}
	api := &platform.Client{Base: base, Key: os.Getenv("COINGECKO_API_KEY"), HTTP: &http.Client{Timeout: 30 * time.Second}, Backoff: time.Second}
	if err = platform.Ingest(ctx, c, api, *config); err != nil {
		slog.Error("ingestion incomplete", "error", err)
		os.Exit(1)
	}
	slog.Info("ingestion completed")
}
