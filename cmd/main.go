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
	"github.com/viczuno/go-crypto-bot/internal/exporter"
	"github.com/viczuno/go-crypto-bot/internal/markdown"
	"github.com/viczuno/go-crypto-bot/internal/service"
)

const (
	dbPath          = "./crypto_history.db"
	readmePath      = "./README.md"
	hugoDataPath    = "./data/crypto.json"
	hugoHistoryPath = "./data/history"
	timeout         = 5 * time.Minute
	historyDays     = 30
)

func main() {
	log.Println("Starting Go-Crypto-Bot...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go handleShutdown(cancel)

	if err := run(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Successfully completed all tasks")
}

func run(ctx context.Context) error {
	fetcher := api.NewCoinGeckoClient()
	repo, err := db.NewSQLiteRepository(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = repo.Close() }()

	coins := domain.DefaultCoins()

	svc := service.NewCryptoService(fetcher, repo, markdown.NewReadmeBuilder())
	content, stats, err := svc.UpdateAndGenerateReport(ctx, coins)
	if err != nil {
		return err
	}

	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return err
	}

	hugo := exporter.NewHugoExporter(hugoDataPath, hugoHistoryPath)
	return hugo.ExportAll(stats, coins, repo, historyDays)
}

func handleShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Shutting down...")
	cancel()
}
