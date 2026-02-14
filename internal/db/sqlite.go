package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/viczuno/go-crypto-bot.git/internal/api"
	_ "modernc.org/sqlite"
)

type CryptoDB struct {
	conn *sql.DB
}

type PriceChange struct {
	PastPrice float64
	AbsChange float64
	PctChange float64
	HasData   bool
}

func InitDB(filepath string) (*CryptoDB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	query := `CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		coin TEXT NOT NULL,
		price REAL NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = db.Exec(query); err != nil {
		return nil, err
	}

	return &CryptoDB{conn: db}, nil
}

func (db *CryptoDB) SavePrices(prices map[string]api.CryptoData) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO prices (coin, price, timestamp) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for coin, data := range prices {
		if _, err = stmt.Exec(coin, data.USD, now); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (db *CryptoDB) GetHistoricalChange(coin string, currentPrice float64, days int) PriceChange {
	var pastPrice float64

	query := `SELECT price FROM prices WHERE coin = ? AND timestamp <= datetime('now', ?) ORDER BY timestamp DESC LIMIT 1`
	timeModifier := fmt.Sprintf("-%d days", days)

	err := db.conn.QueryRow(query, coin, timeModifier).Scan(&pastPrice)
	if err != nil {
		return PriceChange{HasData: false}
	}

	absChange := currentPrice - pastPrice
	pctChange := (absChange / pastPrice) * 100.0

	return PriceChange{
		PastPrice: pastPrice,
		AbsChange: absChange,
		PctChange: pctChange,
		HasData:   true,
	}
}
