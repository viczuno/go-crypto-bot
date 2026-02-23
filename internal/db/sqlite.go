package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteRepository implements domain.PriceRepository using SQLite
type SQLiteRepository struct {
	conn *sql.DB
}

// NewSQLiteRepository creates and initializes a new SQLite repository
func NewSQLiteRepository(filepath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &SQLiteRepository{conn: db}
	if err := repo.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates the required database tables
func (r *SQLiteRepository) initSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS prices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			coin TEXT NOT NULL,
			price REAL NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_prices_coin_timestamp ON prices(coin, timestamp);
	`
	_, err := r.conn.Exec(query)
	return err
}

// SavePrices stores the current prices in the database
func (r *SQLiteRepository) SavePrices(prices map[string]domain.CryptoPrice) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO prices (coin, price, timestamp) VALUES (?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, data := range prices {
		if _, err := stmt.Exec(data.Coin, data.PriceUSD, data.FetchedAt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to insert price for %s: %w", data.Coin, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetHistoricalPrice retrieves the price from a specified number of days ago
func (r *SQLiteRepository) GetHistoricalPrice(coinID string, daysAgo int) (float64, bool, error) {
	query := `
		SELECT price 
		FROM prices 
		WHERE coin = ? AND substr(timestamp, 1, 19) <= datetime('now', ?) 
		ORDER BY timestamp DESC 
		LIMIT 1
	`
	timeModifier := fmt.Sprintf("-%d days", daysAgo)

	var price float64
	err := r.conn.QueryRow(query, coinID, timeModifier).Scan(&price)

	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to query historical price: %w", err)
	}

	return price, true, nil
}

// GetPriceHistory retrieves price history for a coin for the last N days
func (r *SQLiteRepository) GetPriceHistory(coinID string, days int) ([]domain.CryptoPrice, error) {
	query := `
		SELECT coin, price, timestamp 
		FROM prices 
		WHERE coin = ? AND substr(timestamp, 1, 19) >= datetime('now', ?)
		ORDER BY timestamp ASC
	`
	timeModifier := fmt.Sprintf("-%d days", days)

	rows, err := r.conn.Query(query, coinID, timeModifier)
	if err != nil {
		return nil, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	var prices []domain.CryptoPrice
	for rows.Next() {
		var p domain.CryptoPrice
		var timestamp string
		if err := rows.Scan(&p.Coin, &p.PriceUSD, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		parsed, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			parsed, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", timestamp)
			if err != nil {
				parsed, _ = time.Parse("2006-01-02 15:04:05", timestamp[:min(19, len(timestamp))])
			}
		}
		p.FetchedAt = parsed
		prices = append(prices, p)
	}

	return prices, rows.Err()
}

// GetHistoryDaysCount returns the number of days of history available for a coin
func (r *SQLiteRepository) GetHistoryDaysCount(coinID string) (int, error) {
	query := `
		SELECT CAST(julianday('now') - julianday(MIN(substr(timestamp, 1, 19))) AS INTEGER)
		FROM prices 
		WHERE coin = ?
	`
	var days sql.NullInt64
	err := r.conn.QueryRow(query, coinID).Scan(&days)
	if err != nil {
		return 0, fmt.Errorf("failed to query history days: %w", err)
	}
	if !days.Valid {
		return 0, nil
	}
	return int(days.Int64), nil
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Ensure SQLiteRepository implements PriceRepository
var _ domain.PriceRepository = (*SQLiteRepository)(nil)
