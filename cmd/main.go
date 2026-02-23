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
	dbPath          = "./crypto_history.db"
	readmePath      = "./README.md"
	hugoDataPath    = "./data/crypto.json"
	hugoHistoryPath = "./data/history"
	fileMode        = 0644
	dirMode         = 0755
	timeout         = 5 * time.Minute
	requiredDays    = 30
	rateLimitDelay  = 6 * time.Second
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
	fetcher := api.NewCoinGeckoClient()

	repo, err := db.NewSQLiteRepository(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	generator := markdown.NewReadmeBuilder()

	coins := domain.DefaultCoins()

	if err := ensureHistoricalData(ctx, fetcher, repo, coins); err != nil {
		log.Printf("Warning: failed to backfill some historical data: %v", err)
	}

	svc := service.NewCryptoService(fetcher, repo, generator)

	content, stats, err := svc.UpdateAndGenerateReport(ctx, coins)
	if err != nil {
		return err
	}

	if err := os.WriteFile(readmePath, []byte(content), fileMode); err != nil {
		return err
	}

	if err := exportJSONData(stats, coins); err != nil {
		return err
	}

	if err := exportCoinHistories(repo, coins); err != nil {
		return err
	}

	return nil
}

// ensureHistoricalData checks if we have enough history and backfills if needed
func ensureHistoricalData(ctx context.Context, fetcher domain.PriceFetcher, repo *db.SQLiteRepository, coins []domain.CoinMetadata) error {
	for _, coin := range coins {
		daysAvailable, err := repo.GetHistoryDaysCount(coin.ID)
		if err != nil {
			log.Printf("Error checking history for %s: %v", coin.ID, err)
			continue
		}

		if daysAvailable >= requiredDays {
			log.Printf("%s: %d days of history available (sufficient)", coin.Name, daysAvailable)
			continue
		}

		daysNeeded := requiredDays + 5 // Fetch extra days for buffer
		log.Printf("%s: only %d days available, fetching %d days of history...", coin.Name, daysAvailable, daysNeeded)

		// Rate limiting
		time.Sleep(rateLimitDelay)

		prices, err := fetcher.FetchHistoricalPrices(ctx, coin.ID, daysNeeded)
		if err != nil {
			log.Printf("Error fetching history for %s: %v", coin.ID, err)
			continue
		}

		// Save each price point
		for _, price := range prices {
			priceMap := map[string]domain.CryptoPrice{
				coin.ID: price,
			}
			if err := repo.SavePrices(priceMap); err != nil {
				log.Printf("Error saving historical price for %s: %v", coin.ID, err)
			}
		}

		log.Printf("%s: saved %d historical price points", coin.Name, len(prices))
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

// CoinHistory represents the JSON structure for individual coin history
type CoinHistory struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Symbol    string           `json:"symbol"`
	UpdatedAt string           `json:"updated_at"`
	Current   CryptoDataItem   `json:"current"`
	History   []PriceDataPoint `json:"history"`
}

// PriceDataPoint represents a single price point in history
type PriceDataPoint struct {
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
}

// exportCoinHistories exports individual history files for each coin
func exportCoinHistories(repo *db.SQLiteRepository, coins []domain.CoinMetadata) error {
	log.Println("Exporting individual coin history files...")

	// Ensure history directory exists
	if err := os.MkdirAll(hugoHistoryPath, dirMode); err != nil {
		return err
	}

	for _, coin := range coins {
		history, err := repo.GetPriceHistory(coin.ID, requiredDays)
		if err != nil {
			log.Printf("Error getting history for %s: %v", coin.ID, err)
			continue
		}

		// Get current price (most recent)
		var currentPrice float64
		var change24h, change7d, change30d float64
		var has7d, has30d bool
		if len(history) > 0 {
			currentPrice = history[len(history)-1].PriceUSD
		}

		// Calculate changes from history
		if len(history) > 1 {
			now := time.Now().UTC()

			// Find prices at different intervals by looking backwards
			for i := len(history) - 2; i >= 0; i-- {
				hoursDiff := now.Sub(history[i].FetchedAt).Hours()
				daysDiff := hoursDiff / 24

				// 24h change - find first price at least 20 hours ago
				if change24h == 0 && hoursDiff >= 20 {
					change24h = ((currentPrice - history[i].PriceUSD) / history[i].PriceUSD) * 100
				}
				// 7d change - find first price at least 6.5 days ago
				if !has7d && daysDiff >= 6.5 {
					change7d = ((currentPrice - history[i].PriceUSD) / history[i].PriceUSD) * 100
					has7d = true
				}
				// 30d change - find first price at least 28 days ago
				if !has30d && daysDiff >= 28 {
					change30d = ((currentPrice - history[i].PriceUSD) / history[i].PriceUSD) * 100
					has30d = true
					break
				}
			}

			// If we don't have 30d data but have old enough history, use oldest price
			if !has30d && len(history) > 0 {
				oldestDays := now.Sub(history[0].FetchedAt).Hours() / 24
				if oldestDays >= 25 { // Use oldest if it's at least 25 days old
					change30d = ((currentPrice - history[0].PriceUSD) / history[0].PriceUSD) * 100
					has30d = true
				}
			}
		}

		// Build history data points
		historyPoints := make([]PriceDataPoint, 0, len(history))
		for _, h := range history {
			historyPoints = append(historyPoints, PriceDataPoint{
				Timestamp: h.FetchedAt.Format(time.RFC3339),
				Price:     h.PriceUSD,
			})
		}

		coinHistory := CoinHistory{
			ID:        coin.ID,
			Name:      coin.Name,
			Symbol:    coin.Symbol,
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			Current: CryptoDataItem{
				ID:          coin.ID,
				Name:        coin.Name,
				Symbol:      coin.Symbol,
				Price:       currentPrice,
				Change24h:   change24h,
				Change7d:    change7d,
				Change7dOk:  has7d,
				Change30d:   change30d,
				Change30dOk: has30d,
			},
			History: historyPoints,
		}

		// Write to file
		filePath := filepath.Join(hugoHistoryPath, coin.ID+".json")
		jsonBytes, err := json.MarshalIndent(coinHistory, "", "  ")
		if err != nil {
			log.Printf("Error marshaling history for %s: %v", coin.ID, err)
			continue
		}

		if err := os.WriteFile(filePath, jsonBytes, fileMode); err != nil {
			log.Printf("Error writing history file for %s: %v", coin.ID, err)
			continue
		}

		log.Printf("Exported history for %s (%d points)", coin.Name, len(historyPoints))
	}

	return nil
}
