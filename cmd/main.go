package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/viczuno/go-crypto-bot/internal/api"
	"github.com/viczuno/go-crypto-bot/internal/config"
	"github.com/viczuno/go-crypto-bot/internal/db"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"github.com/viczuno/go-crypto-bot/internal/exporter"
	"github.com/viczuno/go-crypto-bot/internal/markdown"
	"github.com/viczuno/go-crypto-bot/internal/service"
)

func main() {
	log.Println("Starting Go-Crypto-Bot...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	go handleShutdown(cancel)

	if err := run(ctx, cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Successfully completed all tasks")
}

func run(ctx context.Context, cfg *config.Config) (err error) {
	fetcher := api.NewCoinGeckoClient(
		api.WithBaseURL(cfg.APIBaseURL),
		api.WithTimeout(cfg.APITimeout),
		api.WithRetry(cfg.MaxRetries, cfg.RetryBackoff),
	)

	repo, err := db.NewSQLiteRepository(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close repository: %w", closeErr)
			} else {
				log.Printf("Warning: failed to close repository: %v", closeErr)
			}
		}
	}()

	coins := domain.DefaultCoins()

	svc := service.NewCryptoService(fetcher, repo, markdown.NewReadmeBuilder())
	content, stats, err := svc.UpdateAndGenerateReport(ctx, coins)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfg.ReadmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	hugo := exporter.NewHugoExporter(cfg.HugoDataPath, cfg.HugoHistoryPath)
	return hugo.ExportAll(ctx, stats, coins, repo, cfg.HistoryDays)
}

func handleShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Shutting down...")
	cancel()
}
