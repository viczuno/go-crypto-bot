package platform

import (
	"context"
	"encoding/json"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRetryClassification(t *testing.T) {
	for _, tc := range []struct{ status, want int }{{400, 1}, {401, 1}, {404, 1}, {429, 4}, {500, 4}, {200, 1}} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(tc.status)
				w.Write([]byte(`[]`))
			}))
			defer s.Close()
			c := Client{Base: s.URL, HTTP: s.Client()}
			var out []Market
			err := c.Get(context.Background(), "/", &out)
			if calls != tc.want {
				t.Fatalf("calls=%d want %d", calls, tc.want)
			}
			if (err == nil) != (tc.status == 200) {
				t.Fatal("incorrect error")
			}
		})
	}
}
func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := Client{Base: "http://127.0.0.1:1", HTTP: &http.Client{Timeout: time.Second}}
	if c.Get(ctx, "/", &[]Market{}) == nil {
		t.Fatal("cancellation ignored")
	}
}
func TestAnalyzeFlatAndPartial(t *testing.T) {
	for _, scale := range []float64{0.000001, 100, 1000000} {
		var prices []domain.CryptoPrice
		for i := 0; i < 201; i++ {
			prices = append(prices, domain.CryptoPrice{PriceUSD: scale, FetchedAt: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC)})
		}
		a := Analyze("test", prices)
		if a.Coverage != 4 || a.Signal != "Hold" {
			t.Fatalf("flat series: %+v", a)
		}
		for _, i := range a.Indicators {
			if i.Value == nil || math.IsNaN(*i.Value) || math.IsNaN(i.Strength) {
				t.Fatal("invalid value")
			}
			if i.Name == "RSI" && *i.Value != 50 {
				t.Fatal("flat RSI must be 50")
			}
			if i.Name == "BollingerBands" && math.Abs(*i.Value-0.5) > 1e-6 {
				t.Fatal("flat bands must be centered")
			}
		}
		if _, e := json.Marshal(a); e != nil {
			t.Fatal(e)
		}
		a = Analyze("test", prices[:20])
		if a.Coverage != 2 {
			t.Fatalf("partial coverage=%d", a.Coverage)
		}
	}
}
func TestArchiveReconciliation(t *testing.T) {
	r, e := ReadArchive("../../archive/crypto_history.db", "../../archive/history")
	if e != nil {
		t.Fatal(e)
	}
	report, e := Import(context.Background(), nil, r, false)
	if e != nil {
		t.Fatal(e)
	}
	if report.SourceRows != 2716 || report.Canonical != 2256 || report.Assets != 16 {
		t.Fatalf("unexpected archive: %+v", report)
	}
	r = append(r, r[0])
	r[len(r)-1].Price++
	if _, e = Import(context.Background(), nil, r, false); e == nil {
		t.Fatal("conflict not rejected")
	}
}
func TestTimestampFormats(t *testing.T) {
	for _, s := range []string{"2026-01-20 00:00:00 +0000 UTC", "2026-02-14 22:05:52.644832581 +0000 UTC", "2026-10-30T17:40:47Z", "2026-01-20 00:00:00"} {
		if _, e := ParseTime(s); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := ParseTime("bad"); e == nil {
		t.Fatal("invalid timestamp accepted")
	}
}
func TestDatabaseIntegration(t *testing.T) {
	if os.Getenv("DATABASE_DIRECT_URL") == "" || os.Getenv("CRYPTO_TEST_DATABASE") != "1" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	ctx := context.Background()
	c, e := Connect(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close(ctx)
	r, e := ReadArchive("../../archive/crypto_history.db", "../../archive/history")
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		if _, e = Import(ctx, c, r, false); e != nil {
			t.Fatal(e)
		}
	}
	if report, e := Import(ctx, c, r, true); e != nil || !report.Verified {
		t.Fatalf("verify: %+v %v", report, e)
	}
	var n int
	if e = c.QueryRow(ctx, "SELECT count(*) FROM observations").Scan(&n); e != nil || n != 2256 {
		t.Fatalf("observation count=%d %v", n, e)
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"test-asset","name":"Test","symbol":"TST","current_price":12,"last_updated":"2026-01-01T12:00:00Z"}]`))
	}))
	defer s.Close()
	p := filepath.Join(t.TempDir(), "coins.yaml")
	os.WriteFile(p, []byte("coin_sources:\n  - type: static\n    coins:\n      - id: test-asset\n      - id: missing-asset\n"), 0600)
	api := Client{Base: s.URL, HTTP: s.Client()}
	for i := 0; i < 2; i++ {
		if Ingest(ctx, c, &api, p) == nil {
			t.Fatal("partial run reported success")
		}
	}
	if e = c.QueryRow(ctx, "SELECT count(*) FROM observations WHERE asset_id='test-asset'").Scan(&n); e != nil || n != 1 {
		t.Fatalf("idempotency failed: %d %v", n, e)
	}
	c.Exec(ctx, "DELETE FROM analytics WHERE asset_id='test-asset'")
	c.Exec(ctx, "DELETE FROM daily_samples WHERE asset_id='test-asset'")
	c.Exec(ctx, "DELETE FROM observations WHERE asset_id='test-asset'")
	c.Exec(ctx, "DELETE FROM assets WHERE id='test-asset'")
}
