# Database and historical migration

PostgreSQL 17 is the durable source of truth. Go writes; the website has SELECT
permissions only. Use a Neon project near the Vercel function region. Production
connections require TLS. Never use a NEXT_PUBLIC_ variable for credentials.

## Local development
Install Go (version in go.mod), Node 24 and Docker. From the repository root:

```sh
cp .env.example .env
cp web/.env.example web/.env.local
set -a
. ./.env
set +a
make db-up
make db-migrate
make import-history
make verify-history
make dev
```

Alternatively start PostgreSQL 17 with your OS tools and create a `crypto`
database. Set both URLs to that database. Go uses DATABASE_DIRECT_URL; Next.js
uses DATABASE_URL. The Go commands intentionally do not implicitly load env files.
Docker's published port is bound to localhost. Its example password is local-only.

## Schema
- `assets`: CoinGecko ID, name, symbol and active discovery membership.
- `observations`: unique asset/provider timestamp, positive numeric price,
  optional market cap, volume and 24h percentage, ingestion time and source.
- `daily_samples`: last observation per completed UTC day, with original time.
- `analytics`: per-day Go-calculated indicator result JSON, algorithm version,
  required/available consecutive samples and unavailability reasons.
- `ingestion_runs`, `ingestion_assets`: completion, counts and partial failures.
- `import_records`: immutable source-row provenance and original timestamp text.

Index keys support bounded asset/time reads. Missing market metadata remains NULL.
No automatic history retention deletion is enabled. Monitor database storage and
back up before considering an explicit retention migration.

## Import protocol
The archive preserves the exact original SQLite bytes and all legacy JSON files.
These are migration inputs only, never production website data sources.

```sh
go run ./cmd/import-history                 # dry-run; no database required
go run ./cmd/import-history -mode import    # additive transaction
go run ./cmd/import-history -mode verify    # validate every source row
```

Expected baseline: 2,716 source records, 2,256 canonical observations, 16 assets.
SQLite contributes 2,395 records, including 150 exact duplicates. JSON adds eleven
unique observations plus overlapping exports. Every source row keeps its identity,
file checksum, raw timestamp and value. Canonical timestamps use PostgreSQL
microseconds; original nanoseconds remain in provenance. Conflicting values at
one canonical timestamp abort the transaction. Invalid timestamps or prices abort
before writes. Repeated imports validate existing rows rather than replace them.

Verification checks source identity, checksum, coin, raw and normalized timestamps,
and price against both provenance and observations. All source dates and values
are therefore reconciled, not just aggregate counts. Additional live observations
are allowed. Archive removal requires a verified external import and backup.

## Roles and operational safety
Run schema migrations/import as owner. Adapt `migrations/roles.sql.example` to
your database name, create strong passwords out of band, and grant the web only
SELECT. Give the scheduler the ingest role and a DIRECT connection: session
advisory locks must not use transaction pooling. The web can use Neon pooling.
The web also requests read-only transactions and five-second statement timeouts.

The importer takes a transaction advisory lock shared with ingestion. Ingestion
holds a session lock until completion; interrupted workers release the lock and
the next worker marks abandoned run records failed. Missing discovery/market
responses cause partial status, retained successful data and a failed job signal.
Freshness uses provider timestamps, never page render time.

Migrations run explicitly, never during a Vercel build. Goose rollback intentionally
refuses to drop historical tables. Back up using `pg_dump` and restore to a new
database for recovery. No production migration has been run by this setup.
