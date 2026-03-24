// Package db provides database implementations for the crypto bot.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
	apperrors "github.com/viczuno/go-crypto-bot/internal/errors"
	_ "modernc.org/sqlite"
)

// timestampFormat is the standard format for storing timestamps.
const timestampFormat = time.RFC3339

// SQLiteRepository implements domain.PriceRepository using SQLite.
type SQLiteRepository struct {
	conn *sql.DB
}

// NewSQLiteRepository creates and initializes a new SQLite repository.
func NewSQLiteRepository(filepath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

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

// initSchema creates the required database tables.
func (r *SQLiteRepository) initSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS prices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			coin TEXT NOT NULL,
			price REAL NOT NULL,
			timestamp TEXT DEFAULT (datetime('now'))
		);
		CREATE INDEX IF NOT EXISTS idx_prices_coin_timestamp ON prices(coin, timestamp);
	`
	_, err := r.conn.Exec(query)
	return err
}

// SavePrices stores the current prices in the database.
func (r *SQLiteRepository) SavePrices(ctx context.Context, prices map[string]domain.CryptoPrice) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.NewDatabaseError("begin transaction", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO prices (coin, price, timestamp) VALUES (?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return apperrors.NewDatabaseError("prepare statement", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, data := range prices {
		timestamp := data.FetchedAt.UTC().Format(timestampFormat)
		if _, err := stmt.ExecContext(ctx, data.Coin, data.PriceUSD, timestamp); err != nil {
			_ = tx.Rollback()
			return apperrors.NewDatabaseError(fmt.Sprintf("insert price for %s", data.Coin), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.NewDatabaseError("commit transaction", err)
	}

	return nil
}

// GetHistoricalPrice retrieves the price from a specified number of days ago.
func (r *SQLiteRepository) GetHistoricalPrice(ctx context.Context, coinID string, daysAgo int) (domain.PriceResult, error) {
	query := `
		SELECT price
		FROM prices
		WHERE coin = ? AND timestamp <= datetime('now', ?)
		ORDER BY timestamp DESC
		LIMIT 1
	`
	timeModifier := fmt.Sprintf("-%d days", daysAgo)

	var price float64
	err := r.conn.QueryRowContext(ctx, query, coinID, timeModifier).Scan(&price)

	if err == sql.ErrNoRows {
		return domain.PriceResult{Found: false}, nil
	}
	if err != nil {
		return domain.PriceResult{}, apperrors.NewDatabaseError("query historical price", err)
	}

	return domain.PriceResult{Price: price, Found: true}, nil
}

// GetPriceHistory retrieves price history for a coin for the last N days.
func (r *SQLiteRepository) GetPriceHistory(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error) {
	query := `
		SELECT coin, price, timestamp
		FROM prices
		WHERE coin = ? AND timestamp >= datetime('now', ?)
		ORDER BY timestamp ASC
	`
	timeModifier := fmt.Sprintf("-%d days", days)

	rows, err := r.conn.QueryContext(ctx, query, coinID, timeModifier)
	if err != nil {
		return nil, apperrors.NewDatabaseError("query price history", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("warning: failed to close rows: %v", err)
		}
	}()

	var prices []domain.CryptoPrice
	for rows.Next() {
		var p domain.CryptoPrice
		var timestamp string
		if err := rows.Scan(&p.Coin, &p.PriceUSD, &timestamp); err != nil {
			return nil, apperrors.NewDatabaseError("scan row", err)
		}

		p.FetchedAt = parseTimestamp(timestamp)
		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("iterate rows", err)
	}

	return prices, nil
}

// parseTimestamp attempts to parse a timestamp string with multiple format fallbacks.
// This handles legacy data that may have been stored in different formats.
func parseTimestamp(timestamp string) time.Time {
	// Try RFC3339 first (preferred format)
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t
	}

	// Try Go's default format
	if t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", timestamp); err == nil {
		return t
	}

	// Try SQLite's default datetime format
	if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
		return t
	}

	// Last resort: truncate and try again
	if len(timestamp) >= 19 {
		if t, err := time.Parse("2006-01-02 15:04:05", timestamp[:19]); err == nil {
			return t
		}
	}

	return time.Time{}
}

// GetHistoryDaysCount returns the number of days of history available for a coin.
func (r *SQLiteRepository) GetHistoryDaysCount(ctx context.Context, coinID string) (int, error) {
	query := `
		SELECT CAST(julianday('now') - julianday(MIN(timestamp)) AS INTEGER)
		FROM prices
		WHERE coin = ?
	`
	var days sql.NullInt64
	err := r.conn.QueryRowContext(ctx, query, coinID).Scan(&days)
	if err != nil {
		return 0, apperrors.NewDatabaseError("query history days", err)
	}
	if !days.Valid {
		return 0, nil
	}
	return int(days.Int64), nil
}

// Close closes the database connection.
func (r *SQLiteRepository) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Ensure SQLiteRepository implements PriceRepository.
var _ domain.PriceRepository = (*SQLiteRepository)(nil)
