# Go Crypto Bot · Market Observatory

A cryptocurrency research application with a Next.js dashboard, scheduled Go
analytics, and durable Neon/PostgreSQL history. Explore tracked markets, inspect
price trends and technical signals, maintain a browser-local watchlist, and compare
relative performance. The Go worker runs on a schedule; it is not a permanent API
server. Vercel serves the website.

![Dashboard using genuine local market data](docs/screenshots/dashboard-desktop.png)

## Features and boundaries

- Responsive dashboard, market explorer, coin research pages, dark/light themes.
- Search, sorting, signal/freshness filters, pagination and persistent row density.
- Interactive sampled-price charts, supported timeframes, overlays and RSI/MACD history.
- Browser-local watchlists and normalized comparisons of up to five assets; no account required.
- Go CoinGecko discovery/ingestion, PostgreSQL persistence and technical analysis.
- Explicit stale, partial, missing and insufficient-history states.

Stack: Go 1.26+, pgx, Goose, PostgreSQL 17, Next.js 16, React 19, strict TypeScript,
Tailwind CSS, Radix Dialog, Lightweight Charts, Recharts, Vitest and Playwright.
Exact dependencies are locked in go.mod/go.sum and web/package-lock.json.

Statistics describe the tracked universe, not the entire cryptocurrency market.
Twice-daily sampled prices are delayed observations, not live quotes or OHLC candles.
Heuristic signal strength is not a validated probability of profit. This is a
research tool; it does not execute trades. No public production URL is claimed here.

## Architecture and data flow

```mermaid
flowchart LR
  Schedule[GitHub Actions: twice daily] --> Go[Go ingestion worker]
  CoinGecko[CoinGecko API] --> Go
  Go --> Analysis[Go daily technical analysis]
  Go --> DB[(Neon PostgreSQL)]
  Analysis --> DB
  DB --> Next[Next.js server components / read APIs]
  Next --> UI[Interactive dashboard on Vercel]
  Archive[Immutable recovery archive] -. explicit recovery only .-> DB
```

CoinGecko supplies provider timestamps and market fields. Go discovers assets,
validates and stores observations, selects completed UTC-day samples, and calculates
indicators. PostgreSQL owns production data. Next.js performs bounded read-only
queries and caches successful snapshots for five minutes. Browsers store only
watchlists/preferences. Vercel's Git integration builds the website when code changes.

Market updates write PostgreSQL directly: they do **not** commit SQLite/JSON, rewrite
this README, or redeploy the website. GitHub Actions runs ingestion and CI separately.
See [Architecture](docs/ARCHITECTURE.md) for numerical semantics and failure handling.

## Repository structure

```text
cmd/                    Go entrypoints: ingest, db, import-history (recovery)
internal/domain/        Analytics types and contracts
internal/indicators/    RSI, MACD, moving averages, Bollinger bands, consensus
internal/platform/      CoinGecko client, PostgreSQL pipeline and recovery adapter
web/                    Next.js app, server reads, components and frontend tests
migrations/             Versioned SQL schema and least-privilege role example
.github/workflows/      CI and twice-daily ingestion; no website deployment job
archive/                Immutable verified history and prelaunch recovery snapshot
docs/                   Detailed architecture, operations runbook and screenshots
```

## macOS development: Neon first, no Docker required

Use a **separate development Neon branch/database**. Never point manual ingestion,
migrations or integration tests at production while developing. Schema/import uses
an owner/admin role; normal ingestion uses a writer role and the web uses a reader.

### 1. Install tools and dependencies

Use Go 1.26 or newer (go.mod), Node **24.x**, and npm bundled with Node 24 (npm 11+).
Git is required. PostgreSQL client tools are useful for SQL/backup work; GitHub CLI
is optional for PRs and Actions. On Apple Silicon with Homebrew:

```sh
brew install go node@24 libpq gh
export PATH="/opt/homebrew/opt/node@24/bin:/opt/homebrew/opt/libpq/bin:$PATH"
git clone https://github.com/viczuno/go-crypto-bot.git
cd go-crypto-bot
make install
```

Use `npm ci` through `make install` to reproduce the committed frontend lockfile.
Do not install a second package manager or regenerate the application.

### 2. Configure separate development connections

In Neon, create a development branch/database. Copy its **direct** URL for Go
and **pooled** URL for Next.js. Both must identify the same development dataset.
Require TLS; `sslmode=verify-full` makes certificate verification explicit.

From the repository root (copy only if these files do not already exist):

```sh
cp -n .env.example .env
cp -n web/.env.example web/.env.local
```

Edit the ignored files with your own values. Quote shell values, especially URLs
containing `&`; the following are placeholders, not credentials:

```dotenv
# .env: Go commands; source explicitly in your shell
DATABASE_DIRECT_URL='postgresql://ROLE:PASSWORD@DIRECT_HOST/DEV_DB?sslmode=verify-full'
COINGECKO_API_KEY='YOUR_DEMO_KEY'
COINGECKO_API_URL='https://api.coingecko.com/api/v3'
```

```dotenv
# web/.env.local: automatically loaded by Next.js, server-side only
DATABASE_URL='postgresql://READER:PASSWORD@POOLED_HOST/DEV_DB?sslmode=verify-full'
```

`DATABASE_DIRECT_URL` must use a direct/session connection because ingestion holds
a session advisory lock. `DATABASE_URL` uses pooling for serverless web reads. Do
not use `NEXT_PUBLIC_` for either. Go does **not** load `.env` automatically:

```sh
set -a
. ./.env
set +a
```

Environment variables already exported in the shell override Next.js env files.
Avoid exporting an unrelated DATABASE_URL while starting the web app.

### 3. Initialize only the development database

```sh
make db-migrate
# Optional: seed a new development database with the preserved real history
make import-history
make verify-history
make dev
```

Open http://localhost:3000. The archive has prices but no historical volume/market
cap; these remain unavailable until a successful ingestion provides them. After
schema initialization, grant the web/ingest roles using the reviewed example in
migrations/roles.sql.example, adapting its database name. Never grant the web owner
credentials merely to make a query succeed.

### 4. Ingest, test and build

```sh
make ingest             # Writes only the configured development database
make test               # Go race tests + frontend unit tests
make lint               # Go vet + frontend lint/types/format checks
make build              # Go packages + production Next.js build
```

Ingestion is an explicit operation, not a formatting/test prerequisite. Database
integration tests are opt-in; follow the isolated-database setup below.

### Optional local PostgreSQL

Homebrew alternative (not Docker):

```sh
brew install postgresql@17
export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"
brew services start postgresql@17
createdb crypto_dev
```

Use a local direct URL such as `'postgresql://YOUR_MAC_USER@localhost/crypto_dev?sslmode=disable'`
for development. Local PostgreSQL needs no external pooler; the web can use the same
server with a separate reader role. Keep production TLS settings unchanged.

Docker alternative: `make db-up` starts the optional Compose PostgreSQL service on
localhost:5432. Its example password is local-only. For Go use
`'postgresql://crypto:local-development-only@localhost:5432/crypto?sslmode=disable'`,
then migrate/seed that disposable database. Docker is not required by normal development.

## Make command reference

Run every target from the repository root. Export Go environment variables first.

| Command | Purpose | Prerequisites | Database effect |
| --- | --- | --- | --- |
| `make help` | List supported commands (also default `make`) | make | None |
| `make install` | Download Go modules; npm ci | Go, Node/npm, network | None |
| `make dev` | Next.js development server | Installed web dependencies, web env | Reads |
| `make db-up` | Optional local PostgreSQL | Docker Compose | Creates/starts local service |
| `make db-migrate` | Apply pending Goose schema versions | Direct admin URL | Schema writes |
| `make import-history` | Explicit archive recovery/seed | Direct admin URL | Additive writes |
| `make verify-history` | Reconcile archive provenance/values | Direct URL, imported dataset | Reads, advisory lock |
| `make ingest` | Collect observations and rebuild analytics | Direct ingest URL, API access | Writes |
| `make test` | Go race + frontend unit tests | Go, installed web dependencies | Opt-in integration writes only |
| `make test-go` | Go race tests | Go | Same integration guard |
| `make test-web` | Vitest | Node/npm dependencies | Optional read-only DB tests |
| `make build` | Go build and Next.js production build | Go, Node/npm dependencies | None required |
| `make lint` | vet, ESLint, TypeScript, Prettier check | Go, web dependencies | None |
| `make fmt` | Format Go and frontend | Go, web dependencies | None; rewrites source |

## PostgreSQL model and safe inspection

| Table | Responsibility |
| --- | --- |
| `assets` | CoinGecko IDs, display metadata and current discovery membership |
| `observations` | Unique asset/provider timestamp, price, nullable market fields and provenance |
| `daily_samples` | Last observation per completed UTC day, preserving actual sample time |
| `analytics` | Go-calculated daily indicators, algorithm version, coverage and unavailable reasons |
| `ingestion_runs` | Overall status, timing and expected/received counts |
| `ingestion_assets` | Per-asset ingestion outcome |
| `import_records` | Immutable source identity, checksums, raw timestamps and values |

Raw observations are never fabricated into candles. Daily samples exclude the
current UTC day. Missing days reset calculation windows: MA200 requires 200
consecutive days and crossover detection requires 201. Retention never silently
deletes historical rows. Run schema migrations explicitly before code requiring a
new schema; web builds and the scheduler do not apply them automatically.

Safe read-only SQL (run through your SQL editor or `psql`, without sharing URLs):

```sql
SELECT current_database(), now();
SELECT count(*) FROM assets;
SELECT count(*) FROM observations;
SELECT count(*) FROM daily_samples;
SELECT count(*) FROM analytics;
SELECT count(*) FROM import_records;
SELECT asset_id, max(observed_at) AS latest, count(*) AS observations
FROM observations GROUP BY asset_id ORDER BY asset_id;
SELECT id, started_at, finished_at, status, expected, received
FROM ingestion_runs ORDER BY id DESC LIMIT 10;
```

The verified original archive represents 2,716 source records, 2,256 canonical
observations and 16 assets. Preserve it and its importer until independently
verified recoverable backups exist. See [Recovery and roles](docs/OPERATIONS.md).

## CoinGecko integration

Create a free Demo key in the [CoinGecko Developer Dashboard](https://www.coingecko.com/en/developers/dashboard)
using **+ Add New Key**. Put it in root `.env` for development and the GitHub Actions
secret `COINGECKO_API_KEY` for scheduling. It is not needed by Vercel web reads.

Test authentication without putting the key in a URL or verbose request log:

```sh
python3 - <<'PY'
import os, urllib.request
request = urllib.request.Request('https://api.coingecko.com/api/v3/ping',
    headers={'x-cg-demo-api-key': os.environ['COINGECKO_API_KEY']})
with urllib.request.urlopen(request, timeout=15) as response:
    print('CoinGecko HTTP', response.status)
PY
```

`coins.yaml` defaults to the top 25 by market cap plus the original five IDs, with
deduplication. Static, top-market-cap and optional trending sources share the client.
Discovery is capped at 250 assets and price requests are batched at 100 IDs. Up to
four attempts handle network failures, HTTP 429 and server errors; non-retryable
client errors stop immediately. Backoff and Retry-After waits are bounded by one
minute and the worker deadline. Partial results are recorded, not presented as a
complete update. Quotas depend on the API plan; monitor usage rather than assume
an unlimited free feed. Data attribution: [CoinGecko](https://www.coingecko.com/).

## Automation: CI versus market updates

`.github/workflows/ci.yaml` checks PRs/code pushes using a disposable PostgreSQL
service, no production secrets. It validates Go, the schema/import, frontend and
browser flows. Browser screenshots/traces are uploaded as CI artifacts.

`.github/workflows/ingest.yaml` runs at **00:17 and 12:17 UTC** on `main`, with a
12-minute job timeout, workflow concurrency control and a database advisory lock.
GitHub timing is best effort; schedules can be delayed or disabled after extended
inactivity. This avoids Vercel Hobby's once-daily cron restriction.

Required Actions repository secrets:

- `DATABASE_DIRECT_URL`: intended production database, restricted ingest role, direct connection.
- `COINGECKO_API_KEY`: CoinGecko Demo key.

After the workflow exists on `main`, trigger it from **Actions → Ingest market
observations → Run workflow → main**, or `gh workflow run ingest.yaml --ref main`.
Manual execution writes data; do not run it as a release smoke test without intent.
Inspect logs in the same Actions page. Success requires a green job and a completed
`ingestion_runs.status = 'success'`; partial runs retain valid observations, record
outcomes and exit nonzero. Interrupted workers release locks; the next job marks
abandoned run records failed. No commits or Vercel deployments result from ingestion.

## Vercel deployment

Connect the repository through Vercel's GitHub integration:

| Setting | Value |
| --- | --- |
| Framework | Next.js |
| Root directory | `web` |
| Node.js | 24.x |
| Install command | `npm ci` |
| Build command | `npm run build` |
| Output directory | Framework default |
| Production branch | `main` |
| Server environment | `DATABASE_URL` using intended pooled reader URL and TLS |

Use separate Production and Preview credentials/databases. Changes to a PR create
Preview deployments; merging code to `main` triggers production through Vercel.
Do not add a GitHub website-deploy workflow. Configure region near Neon. Builds do
not require database credentials or run migrations. Webpack is selected explicitly.

`/api/health` checks connectivity; HTTP 200 alone does not prove table permissions
or valid data. Verify the dashboard, an asset page and comparisons as well. Inspect
Vercel build and runtime logs. Errors log safe stage, SQLSTATE/category or redacted
validation paths, never raw queries/parameters/credentials. Failed market reads
are retryable and are not cached as successful unavailable snapshots.

## Testing

Quick validation is `make test && make lint && make build`. Individual commands:

```sh
go test -race ./cmd/... ./internal/...
go vet ./cmd/... ./internal/...
test -z "$(gofmt -l cmd internal)"
npm --prefix web run typecheck
npm --prefix web run lint
npm --prefix web run format:check
npm --prefix web test
npm --prefix web run build
```

PostgreSQL integration tests **write** and require a disposable local database.
Use a new test database; never source production `.env` for these commands:

```sh
createdb crypto_test
export DATABASE_DIRECT_URL='postgresql://YOUR_MAC_USER@localhost/crypto_test?sslmode=disable'
make db-migrate
CRYPTO_TEST_DATABASE=1 make test-go
# Go integration tests populate the isolated test DB with the archive.
MARKET_TEST_DATABASE_URL="$DATABASE_DIRECT_URL" make test-web
```

Browser tests require a built app and a populated isolated database:

```sh
export DATABASE_URL="$DATABASE_DIRECT_URL"  # isolated local test DB only
npm --prefix web run build
(cd web && npx playwright install chromium && npm run test:e2e)
```

Playwright covers desktop/mobile views, chart controls, theme, watchlist persistence,
comparison, explorer accessibility, API bounds and 404s. It is not a complete manual
accessibility/security audit. Use the dedicated test database for repeatable results.

## Troubleshooting and operations

| Symptom | Action |
| --- | --- |
| Connection refused | Start optional local PostgreSQL or check Neon endpoint/network; verify port |
| Missing env | Source root `.env` for Go; configure `web/.env.local` for Next.js; restart server |
| Lock/pool problems | Use direct URL for Go; pooled URL only for request reads |
| Missing tables/columns | Check database/branch and run reviewed pending migrations on the intended dev DB |
| SQLSTATE 42501 | Grant required SELECT to web or write/sequence access to ingest; do not use owner as workaround |
| Health OK, market unavailable | Inspect sanitized `market_read_failed` stage/SQLSTATE/Zod paths; health only executes SELECT 1 |
| Empty market | Confirm imported/ingested rows and environment scopes; do not invent fallback prices |
| Stale data | Check latest provider timestamps, Actions failures and ingestion_runs |
| HTTP 429 | Check CoinGecko plan quota; allow backoff; avoid repeated manual ingestion |
| Partial/failed ingestion | Inspect Actions plus ingestion_assets; correct credentials/provider availability before retrying |
| Vercel build failure | Check Node 24, root web, npm lockfile and build logs; reproduce make build |
| TLS-mode warning | Use sslmode=verify-full explicitly; do not disable certificate validation in production |

Freshness becomes stale after 18 hours; successful snapshots cache for five minutes.
Inspect Neon storage/compute and API usage, GitHub job logs, and Vercel runtime/build
logs. Free tiers have quotas and no assumed production SLA; paid upgrades are an
operational choice, not required by the architecture. Keep backups independent of
Git and test restores. Browser preferences are device-local and are not backed up.

Never commit `.env`, web env files, owner passwords or API keys. Rotate exposed
credentials. Runtime web credentials should be SELECT-only; ingestion needs writes
but not schema ownership. See [Operations runbook](docs/OPERATIONS.md) for roles,
backup/recovery and release verification. Financial charts use
[TradingView Lightweight Charts](https://www.tradingview.com/).
