package main

import (
	"log"
	"os"

	"github.com/viczuno/go-crypto-bot.git/internal/api"
	"github.com/viczuno/go-crypto-bot.git/internal/db"
	"github.com/viczuno/go-crypto-bot.git/internal/markdown"
)

func main() {
	log.Println("Starting Go-Crypto-Bot...")

	database, err := db.InitDB("./crypto_history.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	coins := []string{"bitcoin", "ethereum", "solana", "cardano", "polkadot"}
	liveData, err := api.FetchCryptoData(coins)
	if err != nil {
		log.Fatalf("Failed to fetch API data: %v", err)
	}

	if err := database.SavePrices(liveData); err != nil {
		log.Fatalf("Failed to save to database: %v", err)
	}

	var stats []markdown.CoinStats
	for _, coin := range coins {
		data := liveData[coin]
		stats = append(stats, markdown.CoinStats{
			Name:      coin,
			Price:     data.USD,
			Change24h: data.USD24hChange,
			Change7d:  database.GetHistoricalChange(coin, data.USD, 7),
			Change30d: database.GetHistoricalChange(coin, data.USD, 30),
		})
	}

	readmeContent := markdown.BuildReadme(stats)
	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		log.Fatalf("Failed to write README.md: %v", err)
	}

	log.Println("Execution successful! README and Database updated.")
}
