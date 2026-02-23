package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/api"
	"github.com/viczuno/go-crypto-bot/internal/db"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"github.com/viczuno/go-crypto-bot/internal/markdown"
	"github.com/viczuno/go-crypto-bot/internal/service"
)

const (
	dbPath       = "./crypto_history.db"
	readmePath   = "./README.md"
	hugoDataPath = "./data/crypto.json"
	fileMode     = 0644
	dirMode      = 0755
	timeout      = 2 * time.Minute
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

	log.Println("Execution successful! README, Database, and JSON data updated.")
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
	content, stats, err := svc.UpdateAndGenerateReport(ctx, coins)
	if err != nil {
		return err
	}

	// Write README
	if err := os.WriteFile(readmePath, []byte(content), fileMode); err != nil {
		return err
	}

	// Export JSON data for Hugo
	if err := exportJSONData(stats, coins); err != nil {
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

// CryptoData represents the JSON structure for Hugo data templates
type CryptoData struct {
	UpdatedAt string           `json:"updated_at"`
	Coins     []CryptoDataItem `json:"coins"`
}

// CryptoDataItem represents a single coin entry in the JSON
type CryptoDataItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Price       float64 `json:"price"`
	Change24h   float64 `json:"change_24h"`
	Change7d    float64 `json:"change_7d"`
	Change7dOk  bool    `json:"change_7d_ok"`
	Change30d   float64 `json:"change_30d"`
	Change30dOk bool    `json:"change_30d_ok"`
}

// exportJSONData writes the crypto stats to a JSON file for Hugo
func exportJSONData(stats []domain.CoinStats, coins []domain.CoinMetadata) error {
	log.Println("Exporting JSON data for Hugo...")

	// Create a map for quick metadata lookup
	coinMeta := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMeta[c.ID] = c
	}

	// Build JSON data structure
	data := CryptoData{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Coins:     make([]CryptoDataItem, 0, len(stats)),
	}

	for _, stat := range stats {
		meta := coinMeta[stat.Name]
		item := CryptoDataItem{
			ID:          stat.Name,
			Name:        meta.Name,
			Symbol:      stat.Symbol,
			Price:       stat.Price,
			Change24h:   stat.Change24h,
			Change7d:    stat.Change7d.PctChange,
			Change7dOk:  stat.Change7d.HasData,
			Change30d:   stat.Change30d.PctChange,
			Change30dOk: stat.Change30d.HasData,
		}
		data.Coins = append(data.Coins, item)
	}

	// Ensure data directory exists
	dir := filepath.Dir(hugoDataPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return err
	}

	// Marshal to JSON with pretty printing
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(hugoDataPath, jsonBytes, fileMode); err != nil {
		return err
	}

	log.Printf("JSON data exported to %s", hugoDataPath)
	return nil
}
