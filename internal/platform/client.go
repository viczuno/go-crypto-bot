package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Market struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Symbol  string    `json:"symbol"`
	Price   *float64  `json:"current_price"`
	Cap     *float64  `json:"market_cap"`
	Volume  *float64  `json:"total_volume"`
	Change  *float64  `json:"price_change_percentage_24h"`
	Updated time.Time `json:"last_updated"`
}
type Client struct {
	Base, Key string
	HTTP      *http.Client
	Backoff   time.Duration
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
	var last error
	for attempt := 0; attempt < 4; attempt++ {
		req, e := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(c.Base, "/")+path, nil)
		if e != nil {
			return e
		}
		req.Header.Set("Accept", "application/json")
		if c.Key != "" {
			req.Header.Set("x-cg-demo-api-key", c.Key)
		}
		resp, e := c.HTTP.Do(req)
		delay := c.Backoff * time.Duration(1<<attempt)
		if e != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			last = errors.New("market request failed")
		} else {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
			resp.Body.Close()
			if readErr != nil {
				return fmt.Errorf("read market response: %w", readErr)
			}
			if resp.StatusCode == 200 {
				if e = json.Unmarshal(body, out); e != nil {
					return fmt.Errorf("invalid market response: %w", e)
				}
				return nil
			}
			last = fmt.Errorf("market provider HTTP %d", resp.StatusCode)
			if resp.StatusCode != 429 && resp.StatusCode < 500 {
				return last
			}
			if v := resp.Header.Get("Retry-After"); v != "" {
				if n, e := strconv.Atoi(v); e == nil && n >= 0 {
					delay = time.Duration(n) * time.Second
				} else if t, e := http.ParseTime(v); e == nil {
					delay = time.Until(t)
				}
			}
		}
		if attempt == 3 {
			break
		}
		if delay > time.Minute {
			delay = time.Minute
		}
		if delay < 0 {
			delay = 0
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return last
}
func (c *Client) Markets(ctx context.Context, ids []string, top int) ([]Market, error) {
	q := url.Values{"vs_currency": {"usd"}, "sparkline": {"false"}, "per_page": {"100"}, "page": {"1"}, "order": {"market_cap_desc"}}
	if len(ids) > 0 {
		q.Set("ids", strings.Join(ids, ","))
	}
	if top > 0 {
		q.Set("per_page", strconv.Itoa(top))
	}
	var m []Market
	err := c.Get(ctx, "/coins/markets?"+q.Encode(), &m)
	return m, err
}
func valid(m Market, now time.Time) bool {
	if m.ID == "" || m.Price == nil || *m.Price <= 0 || math.IsNaN(*m.Price) || math.IsInf(*m.Price, 0) || m.Updated.IsZero() || m.Updated.After(now.Add(5*time.Minute)) {
		return false
	}
	for _, v := range []*float64{m.Cap, m.Volume, m.Change} {
		if v != nil && (math.IsNaN(*v) || math.IsInf(*v, 0)) {
			return false
		}
	}
	return (m.Cap == nil || *m.Cap >= 0) && (m.Volume == nil || *m.Volume >= 0)
}
