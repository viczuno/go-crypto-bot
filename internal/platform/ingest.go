package platform

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources []struct {
		Type    string   `yaml:"type"`
		Count   int      `yaml:"count"`
		Exclude []string `yaml:"exclude"`
		Coins   []struct {
			ID string `yaml:"id"`
		} `yaml:"coins"`
	} `yaml:"coin_sources"`
}

func Ingest(ctx context.Context, c *pgx.Conn, api *Client, path string) (err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg Config
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return err
	}
	if len(cfg.Sources) == 0 {
		return fmt.Errorf("no configured coin sources")
	}
	var locked bool
	if err = c.QueryRow(ctx, "SELECT pg_try_advisory_lock(482941)").Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return fmt.Errorf("another ingestion is running")
	}
	defer c.Exec(context.Background(), "SELECT pg_advisory_unlock(482941)")
	// A lock acquired after a terminated process proves previous running entries are abandoned.
	if _, err = c.Exec(ctx, `UPDATE ingestion_runs SET status='failed',finished_at=now(),message='Previous worker interrupted' WHERE status='running'`); err != nil {
		return err
	}
	var run int64
	if err = c.QueryRow(ctx, `INSERT INTO ingestion_runs(status) VALUES('running') RETURNING id`).Scan(&run); err != nil {
		return err
	}
	status := "failed"
	expected, received := 0, 0
	message := "Ingestion failed"
	defer func() {
		finish, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, e := c.Exec(finish, `UPDATE ingestion_runs SET status=$2,finished_at=now(),expected=$3,received=$4,message=$5 WHERE id=$1`, run, status, expected, received, message)
		if err == nil && e != nil {
			err = e
		}
	}()
	ids := map[string]bool{}
	partial := false
	for _, s := range cfg.Sources {
		switch s.Type {
		case "static":
			for _, coin := range s.Coins {
				if coin.ID != "" {
					ids[coin.ID] = true
				}
			}
		case "top_by_market_cap":
			if s.Count < 1 || s.Count > 100 {
				return fmt.Errorf("top count must be 1..100")
			}
			ms, e := api.Markets(ctx, nil, s.Count)
			if e != nil {
				partial = true
				continue
			}
			for _, m := range ms {
				excluded := false
				for _, x := range s.Exclude {
					if x == m.ID {
						excluded = true
					}
				}
				if !excluded {
					ids[m.ID] = true
				}
			}
		case "trending":
			if s.Count < 1 || s.Count > 100 {
				return fmt.Errorf("trending count must be 1..100")
			}
			var data struct {
				Coins []struct {
					Item struct {
						ID string `json:"id"`
					} `json:"item"`
				} `json:"coins"`
			}
			if e := api.Get(ctx, "/search/trending", &data); e != nil {
				partial = true
				continue
			}
			for i, m := range data.Coins {
				if i >= s.Count {
					break
				}
				ids[m.Item.ID] = true
			}
		default:
			return fmt.Errorf("unknown coin source %q", s.Type)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("no assets discovered")
	}
	if len(ids) > 250 {
		return fmt.Errorf("tracked asset limit is 250")
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	sort.Strings(list)
	expected = len(list)
	tx, err := c.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if !partial {
		if _, err = tx.Exec(ctx, "UPDATE assets SET active=false"); err != nil {
			return err
		}
	}
	for start := 0; start < len(list); start += 100 {
		end := start + 100
		if end > len(list) {
			end = len(list)
		}
		batch := list[start:end]
		ms, e := api.Markets(ctx, batch, 0)
		if e != nil {
			partial = true
		}
		found := map[string]Market{}
		for _, m := range ms {
			if ids[m.ID] && valid(m, time.Now()) {
				found[m.ID] = m
			}
		}
		for _, id := range batch {
			m, ok := found[id]
			assetStatus, reason := "success", ""
			if !ok {
				partial = true
				assetStatus = "failed"
				reason = "Missing or invalid provider observation"
			} else {
				if _, err = tx.Exec(ctx, `INSERT INTO assets(id,name,symbol,active) VALUES($1,$2,$3,true) ON CONFLICT(id) DO UPDATE SET name=excluded.name,symbol=excluded.symbol,active=true`, id, m.Name, m.Symbol); err != nil {
					return err
				}
				if _, err = tx.Exec(ctx, `INSERT INTO observations(asset_id,observed_at,price,market_cap,volume,change_24h,source) VALUES($1,$2,$3,$4,$5,$6,'coingecko') ON CONFLICT(asset_id,observed_at) DO NOTHING`, id, m.Updated, *m.Price, m.Cap, m.Volume, m.Change); err != nil {
					return err
				}
				if err = Rebuild(ctx, tx, id); err != nil {
					return err
				}
				received++
			}
			if _, err = tx.Exec(ctx, `INSERT INTO ingestion_assets(run_id,asset_id,status,message) VALUES($1,$2,$3,$4)`, run, id, assetStatus, reason); err != nil {
				return err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	status = "success"
	message = "Completed"
	if partial {
		status = "partial"
		message = "Some discovery sources or assets failed"
		return fmt.Errorf("partial ingestion: %d/%d assets", received, expected)
	}
	return nil
}
