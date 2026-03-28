package main

import (
	"context"
	"flag"
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

var (
	validateCoins = flag.Bool("validate-coins", false, "Validate coin configuration and exit")
	coinConfigPath = flag.String("coin-config", "coins.yaml", "Path to coin configuration file")
)

func main() {
	flag.Parse()

	log.Println("Starting Go-Crypto-Bot...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if *validateCoins {
		ctx := context.Background()
		if err := config.ValidateCoins(ctx, *coinConfigPath, cfg.APIBaseURL); err != nil {
			log.Fatalf("Coin validation failed: %v", err)
		}
		log.Println("✓ Coin configuration is valid")
		return
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

	coins := loadCoins(ctx, *coinConfigPath, cfg.APIBaseURL)

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

func loadCoins(ctx context.Context, configPath, apiBaseURL string) []domain.CoinMetadata {
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("Loading coins from configuration file: %s", configPath)
		coins, err := config.LoadCoinsFromConfig(ctx, configPath, apiBaseURL)
		if err != nil {
			log.Printf("Warning: failed to load coins from config: %v", err)
			log.Println("Falling back to default coins")
			return domain.DefaultCoins()
		}
		log.Printf("Successfully loaded %d coins from configuration", len(coins))
		return coins
	}

	log.Printf("Coin configuration file not found: %s", configPath)
	log.Println("Using default coins")
	return domain.DefaultCoins()
}

func handleShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Shutting down...")
	cancel()
}
