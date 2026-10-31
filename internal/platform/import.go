package platform

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"math"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Record struct {
	Source, ID, Checksum, Coin, Raw string
	Name, Symbol                    string
	Price                           float64
	Time                            time.Time
}

func ParseTime(s string) (time.Time, error) {
	for _, f := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02 15:04:05"} {
		if t, e := time.Parse(f, s); e == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", s)
}
func ReadArchive(dbPath, jsonDir string) ([]Record, error) {
	b, e := os.ReadFile(dbPath)
	if e != nil {
		return nil, e
	}
	sum := sha256.Sum256(b)
	checksum := hex.EncodeToString(sum[:])
	abs, e := filepath.Abs(dbPath)
	if e != nil {
		return nil, e
	}
	u := url.URL{Scheme: "file", Path: abs, RawQuery: "mode=ro"}
	db, e := sql.Open("sqlite", u.String())
	if e != nil {
		return nil, e
	}
	defer db.Close()
	rows, e := db.Query("SELECT id,coin,price,timestamp FROM prices ORDER BY id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var records []Record
	metadata := map[string]domain.CoinMetadata{}
	for _, m := range domain.DefaultCoins() {
		metadata[m.ID] = m
	}
	for rows.Next() {
		var r Record
		r.Source = "sqlite"
		r.Checksum = checksum
		if e = rows.Scan(&r.ID, &r.Coin, &r.Price, &r.Raw); e != nil {
			return nil, e
		}
		if r.Time, e = ParseTime(r.Raw); e != nil {
			return nil, e
		}
		records = append(records, r)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	paths, e := filepath.Glob(filepath.Join(jsonDir, "*.json"))
	if e != nil {
		return nil, e
	}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			return nil, e
		}
		sum := sha256.Sum256(b)
		var d struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
			History []struct {
				Timestamp string  `json:"timestamp"`
				Price     float64 `json:"price"`
			} `json:"history"`
		}
		if e = json.Unmarshal(b, &d); e != nil {
			return nil, e
		}
		metadata[d.ID] = domain.CoinMetadata{ID: d.ID, Name: d.Name, Symbol: d.Symbol}
		for i, h := range d.History {
			t, e := ParseTime(h.Timestamp)
			if e != nil {
				return nil, e
			}
			records = append(records, Record{Source: filepath.Base(p), ID: strconv.Itoa(i), Checksum: hex.EncodeToString(sum[:]), Coin: d.ID, Raw: h.Timestamp, Price: h.Price, Time: t})
		}
	}
	for i := range records {
		r := records[i]
		m := metadata[r.Coin]
		records[i].Name = m.Name
		records[i].Symbol = m.Symbol
		if records[i].Name == "" {
			records[i].Name = r.Coin
		}
		if records[i].Symbol == "" {
			records[i].Symbol = r.Coin
		}
		if r.Coin == "" || r.Price <= 0 || math.IsNaN(r.Price) || math.IsInf(r.Price, 0) {
			return nil, fmt.Errorf("invalid source row %s/%s", r.Source, r.ID)
		}
	}
	return records, nil
}

type ImportReport struct {
	SourceRows int  `json:"sourceRows"`
	Canonical  int  `json:"canonicalObservations"`
	Assets     int  `json:"assets"`
	Verified   bool `json:"verified"`
}

func Import(ctx context.Context, c *pgx.Conn, records []Record, verify bool) (ImportReport, error) {
	report := ImportReport{SourceRows: len(records)}
	unique := map[string]Record{}
	assets := map[string]bool{}
	for _, r := range records {
		key := r.Coin + "/" + r.Time.Truncate(time.Microsecond).Format(time.RFC3339Nano)
		if old, ok := unique[key]; ok && old.Price != r.Price {
			return report, fmt.Errorf("conflicting values at %s", key)
		}
		unique[key] = r
		assets[r.Coin] = true
	}
	report.Canonical = len(unique)
	report.Assets = len(assets)
	if c == nil {
		return report, nil
	}
	tx, e := c.Begin(ctx)
	if e != nil {
		return report, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(482941)"); e != nil {
		return report, e
	}
	for _, r := range records {
		if !verify {
			if _, e = tx.Exec(ctx, `INSERT INTO assets(id,name,symbol) VALUES($1,$2,$3) ON CONFLICT(id) DO UPDATE SET name=CASE WHEN assets.name=assets.id THEN excluded.name ELSE assets.name END,symbol=CASE WHEN assets.symbol=assets.id THEN excluded.symbol ELSE assets.symbol END`, r.Coin, r.Name, r.Symbol); e != nil {
				return report, e
			}
			if _, e = tx.Exec(ctx, `INSERT INTO import_records(source,row_id,checksum,asset_id,raw_timestamp,price,observed_at) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, r.Source, r.ID, r.Checksum, r.Coin, r.Raw, r.Price, r.Time); e != nil {
				return report, e
			}
			if _, e = tx.Exec(ctx, `INSERT INTO observations(asset_id,observed_at,price,source) VALUES($1,$2,$3,'archive') ON CONFLICT DO NOTHING`, r.Coin, r.Time, r.Price); e != nil {
				return report, e
			}
		}
		var matches bool
		e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM import_records i JOIN observations o ON o.asset_id=i.asset_id AND o.observed_at=i.observed_at WHERE i.source=$1 AND i.row_id=$2 AND i.checksum=$3 AND i.asset_id=$4 AND i.raw_timestamp=$5 AND i.price=$6 AND o.price::float8=$6 AND i.observed_at=$7)`, r.Source, r.ID, r.Checksum, r.Coin, r.Raw, r.Price, r.Time).Scan(&matches)
		if e != nil {
			return report, e
		}
		if !matches {
			return report, fmt.Errorf("verification failed for %s/%s; import rolled back", r.Source, r.ID)
		}
	}
	if !verify {
		for id := range assets {
			if e = Rebuild(ctx, tx, id); e != nil {
				return report, e
			}
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return report, e
	}
	report.Verified = true
	return report, nil
}
