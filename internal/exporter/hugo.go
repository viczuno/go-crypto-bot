// Package exporter provides data export functionality for the crypto bot.
package exporter

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

const (
	fileMode = 0644
	dirMode  = 0755
)

// CryptoData represents the JSON structure for Hugo data templates.
type CryptoData struct {
	UpdatedAt string           `json:"updated_at"`
	Coins     []CryptoDataItem `json:"coins"`
}

// CryptoDataItem represents a single coin entry in the JSON.
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

// CoinHistory represents the JSON structure for individual coin history.
type CoinHistory struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Symbol    string           `json:"symbol"`
	UpdatedAt string           `json:"updated_at"`
	Current   CryptoDataItem   `json:"current"`
	History   []PriceDataPoint `json:"history"`
}

// PriceDataPoint represents a single price point in history.
type PriceDataPoint struct {
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
}

// HugoExporter exports data for Hugo static site generation.
type HugoExporter struct {
	dataPath    string
	historyPath string
}

// NewHugoExporter creates a new Hugo exporter.
func NewHugoExporter(dataPath, historyPath string) *HugoExporter {
	return &HugoExporter{
		dataPath:    dataPath,
		historyPath: historyPath,
	}
}

// ExportAll exports crypto.json and all coin history files.
func (e *HugoExporter) ExportAll(ctx context.Context, stats []domain.CoinStats, coins []domain.CoinMetadata, repo domain.PriceRepository, days int) error {
	if err := e.ExportCryptoData(stats, coins); err != nil {
		return err
	}

	for _, coin := range coins {
		history, err := repo.GetPriceHistory(ctx, coin.ID, days)
		if err != nil {
			log.Printf("Warning: failed to get history for %s: %v", coin.ID, err)
			continue
		}
		if err := e.ExportCoinHistory(coin, history); err != nil {
			log.Printf("Warning: failed to export history for %s: %v", coin.ID, err)
		}
	}

	return nil
}

// ExportCryptoData exports the main crypto.json file.
func (e *HugoExporter) ExportCryptoData(stats []domain.CoinStats, coins []domain.CoinMetadata) error {
	coinMeta := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMeta[c.ID] = c
	}

	data := CryptoData{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Coins:     make([]CryptoDataItem, 0, len(stats)),
	}

	for _, stat := range stats {
		meta := coinMeta[stat.ID]
		data.Coins = append(data.Coins, CryptoDataItem{
			ID:          stat.ID,
			Name:        meta.Name,
			Symbol:      stat.Symbol,
			Price:       stat.Price,
			Change24h:   stat.Change24h,
			Change7d:    stat.Change7d.PctChange,
			Change7dOk:  stat.Change7d.HasData,
			Change30d:   stat.Change30d.PctChange,
			Change30dOk: stat.Change30d.HasData,
		})
	}

	if err := os.MkdirAll(filepath.Dir(e.dataPath), dirMode); err != nil {
		return err
	}

	if err := writeJSON(e.dataPath, data); err != nil {
		return err
	}

	log.Printf("Exported crypto data to %s", e.dataPath)
	return nil
}

// ExportCoinHistory exports individual history file for a coin.
func (e *HugoExporter) ExportCoinHistory(coin domain.CoinMetadata, history []domain.CryptoPrice) error {
	if err := os.MkdirAll(e.historyPath, dirMode); err != nil {
		return err
	}

	changes := calculateChanges(history)

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
			Price:       changes.price,
			Change24h:   changes.change24h,
			Change7d:    changes.change7d.PctChange,
			Change7dOk:  changes.change7d.HasData,
			Change30d:   changes.change30d.PctChange,
			Change30dOk: changes.change30d.HasData,
		},
		History: historyPoints,
	}

	filePath := filepath.Join(e.historyPath, coin.ID+".json")
	if err := writeJSON(filePath, coinHistory); err != nil {
		return err
	}

	log.Printf("Exported %s history (%d points)", coin.Name, len(historyPoints))
	return nil
}

type priceChanges struct {
	price     float64
	change24h float64
	change7d  domain.PriceChange
	change30d domain.PriceChange
}

func calculateChanges(history []domain.CryptoPrice) priceChanges {
	var changes priceChanges

	if len(history) == 0 {
		return changes
	}

	changes.price = history[len(history)-1].PriceUSD

	if len(history) < 2 {
		return changes
	}

	now := time.Now().UTC()
	var found24h bool

	for i := len(history) - 2; i >= 0; i-- {
		hoursDiff := now.Sub(history[i].FetchedAt).Hours()
		daysDiff := hoursDiff / 24

		if !found24h && hoursDiff >= domain.Hours24Threshold {
			changes.change24h = calcPctChange(changes.price, history[i].PriceUSD)
			found24h = true
		}
		if !changes.change7d.HasData && daysDiff >= domain.Days7Threshold {
			changes.change7d = domain.CalculatePriceChange(changes.price, history[i].PriceUSD, domain.Days7)
		}
		if !changes.change30d.HasData && daysDiff >= domain.Days30Threshold {
			changes.change30d = domain.CalculatePriceChange(changes.price, history[i].PriceUSD, domain.Days30)
			break
		}
	}

	// Fallback for 30d if we have enough data but didn't hit the threshold
	if !changes.change30d.HasData {
		oldestDays := now.Sub(history[0].FetchedAt).Hours() / 24
		if oldestDays >= domain.Days30Fallback {
			changes.change30d = domain.CalculatePriceChange(changes.price, history[0].PriceUSD, domain.Days30)
		}
	}

	return changes
}

func calcPctChange(current, past float64) float64 {
	if past == 0 {
		return 0
	}
	return ((current - past) / past) * 100
}

func writeJSON(path string, data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonBytes, fileMode)
}
