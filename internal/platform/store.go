// Package platform implements durable market ingestion and daily analytics.
package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"github.com/viczuno/go-crypto-bot/internal/indicators"
	"math"
	"os"
	"time"
)

func Connect(ctx context.Context) (*pgx.Conn, error) {
	u := os.Getenv("DATABASE_DIRECT_URL")
	if u == "" {
		return nil, fmt.Errorf("DATABASE_DIRECT_URL is required")
	}
	c, err := pgx.Connect(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("database connection failed (check configuration)")
	}
	return c, nil
}

type Indicator struct {
	Name     string             `json:"name"`
	Value    *float64           `json:"value"`
	Signal   string             `json:"signal"`
	Strength float64            `json:"strength"`
	Values   map[string]float64 `json:"values"`
	Required int                `json:"required"`
	Reason   string             `json:"reason,omitempty"`
}
type Analysis struct {
	Day        string      `json:"day"`
	Samples    int         `json:"samples"`
	Signal     string      `json:"signal"`
	Strength   float64     `json:"strength"`
	Coverage   int         `json:"coverage"`
	Indicators []Indicator `json:"indicators"`
}

func Analyze(id string, prices []domain.CryptoPrice) Analysis {
	a := Analysis{Signal: "Unavailable", Samples: len(prices), Indicators: []Indicator{}}
	if len(prices) > 0 {
		a.Day = prices[len(prices)-1].FetchedAt.Format("2006-01-02")
	}
	s := indicators.NewSignalGenerator().GenerateSignals(id, prices)
	required := map[domain.Indicator]int{domain.IndicatorRSI: 15, domain.IndicatorMACD: 35, domain.IndicatorMovingAverage: 200, domain.IndicatorBollingerBands: 20}
	for _, r := range s.Indicators {
		i := Indicator{Name: string(r.Indicator), Signal: "Unavailable", Required: required[r.Indicator], Values: map[string]float64{}}
		if r.Error != nil {
			i.Reason = "Insufficient consecutive completed UTC days"
		} else if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) || math.IsNaN(r.Confidence) {
			i.Reason = "Invalid calculation"
		} else {
			v := r.Value
			i.Value = &v
			i.Signal = r.Signal.String()
			i.Strength = r.Confidence
			i.Values = r.Metadata
			a.Coverage++
		}
		a.Indicators = append(a.Indicators, i)
	}
	if a.Coverage > 0 {
		a.Signal = s.Consensus.String()
		a.Strength = s.Confidence
	}
	return a
}

// Rebuild derives completed-day samples and versioned analytics in the caller's transaction.
// Input history is never overwritten or interpolated. Gaps reset the indicator segment.
func Rebuild(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `INSERT INTO daily_samples(asset_id,day,observed_at,price)
 SELECT DISTINCT ON (asset_id,(observed_at AT TIME ZONE 'UTC')::date)
 asset_id,(observed_at AT TIME ZONE 'UTC')::date,observed_at,price FROM observations
 WHERE asset_id=$1 AND observed_at < date_trunc('day',now() AT TIME ZONE 'UTC') AT TIME ZONE 'UTC'
 ORDER BY asset_id,(observed_at AT TIME ZONE 'UTC')::date,observed_at DESC
 ON CONFLICT(asset_id,day) DO UPDATE SET price=excluded.price,observed_at=excluded.observed_at`, id)
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT day,price::float8 FROM daily_samples WHERE asset_id=$1 ORDER BY day`, id)
	if err != nil {
		return err
	}
	var all []domain.CryptoPrice
	for rows.Next() {
		var p domain.CryptoPrice
		p.Coin = id
		if err = rows.Scan(&p.FetchedAt, &p.PriceUSD); err != nil {
			rows.Close()
			return err
		}
		all = append(all, p)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}
	var segment []domain.CryptoPrice
	for _, p := range all {
		if len(segment) > 0 && p.FetchedAt.Sub(segment[len(segment)-1].FetchedAt) != 24*time.Hour {
			segment = nil
		}
		segment = append(segment, p)
		a := Analyze(id, segment)
		b, e := json.Marshal(a)
		if e != nil {
			return e
		}
		if _, err = tx.Exec(ctx, `INSERT INTO analytics(asset_id,day,algorithm,result) VALUES($1,$2,'daily-v1',$3) ON CONFLICT(asset_id,day) DO UPDATE SET result=excluded.result,algorithm=excluded.algorithm`, id, p.FetchedAt, b); err != nil {
			return err
		}
	}
	return nil
}
