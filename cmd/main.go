package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/api"
	"github.com/viczuno/go-crypto-bot/internal/db"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"github.com/viczuno/go-crypto-bot/internal/markdown"
	"github.com/viczuno/go-crypto-bot/internal/service"
)

const (
	dbPath     = "./crypto_history.db"
	readmePath = "./README.md"
	fileMode   = 0644
	timeout    = 2 * time.Minute
)

func main() {
	log.Println("Starting Go-Crypto-Bot...")

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Handle graceful shutdown
	go handleSignals(cancel)

	if err := run(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Execution successful! README and Database updated.")
}

func run(ctx context.Context) error {
	// Initialize dependencies
	fetcher := api.NewCoinGeckoClient()

	repo, err := db.NewSQLiteRepository(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	generator := markdown.NewReadmeBuilder()

	// Create and run service
	svc := service.NewCryptoService(fetcher, repo, generator)

	coins := domain.DefaultCoins()
	content, err := svc.UpdateAndGenerateReport(ctx, coins)
	if err != nil {
		return err
	}

	// Write README
	if err := os.WriteFile(readmePath, []byte(content), fileMode); err != nil {
		return err
	}

	return nil
}

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received shutdown signal, cancelling...")
	cancel()
}
