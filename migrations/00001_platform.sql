-- +goose Up
CREATE TABLE assets (
 id text PRIMARY KEY, name text NOT NULL, symbol text NOT NULL,
 active boolean NOT NULL DEFAULT false
);
CREATE TABLE ingestion_runs (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 started_at timestamptz NOT NULL DEFAULT now(), finished_at timestamptz,
 status text NOT NULL CHECK(status IN ('running','success','partial','failed')),
 expected integer NOT NULL DEFAULT 0, received integer NOT NULL DEFAULT 0,
 message text NOT NULL DEFAULT ''
);
CREATE TABLE observations (
 asset_id text NOT NULL REFERENCES assets(id), observed_at timestamptz NOT NULL,
 price numeric NOT NULL CHECK(price > 0 AND price <> 'NaN'::numeric),
 market_cap numeric, volume numeric, change_24h double precision,
 ingested_at timestamptz NOT NULL DEFAULT now(), source text NOT NULL,
 PRIMARY KEY(asset_id, observed_at)
);
CREATE INDEX observations_time ON observations(observed_at DESC);
CREATE TABLE daily_samples (
 asset_id text NOT NULL REFERENCES assets(id), day date NOT NULL,
 observed_at timestamptz NOT NULL, price numeric NOT NULL,
 PRIMARY KEY(asset_id,day)
);
CREATE TABLE analytics (
 asset_id text NOT NULL REFERENCES assets(id), day date NOT NULL,
 algorithm text NOT NULL, result jsonb NOT NULL,
 PRIMARY KEY(asset_id,day)
);
CREATE TABLE ingestion_assets (
 run_id bigint NOT NULL REFERENCES ingestion_runs(id), asset_id text NOT NULL,
 status text NOT NULL, message text NOT NULL DEFAULT '', PRIMARY KEY(run_id,asset_id)
);
CREATE TABLE import_records (
 source text NOT NULL, row_id text NOT NULL, checksum text NOT NULL,
 asset_id text NOT NULL, raw_timestamp text NOT NULL, price double precision NOT NULL,
 observed_at timestamptz NOT NULL, PRIMARY KEY(source,row_id)
);
-- +goose Down
-- Deliberately refuse automatic history deletion.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Restore a verified backup instead of dropping historical data'; END $$;
-- +goose StatementEnd
