# Architecture

The platform separates scheduled analytics from request rendering. A Go worker
reads CoinGecko, deduplicates configured asset discovery, persists observations,
and builds daily technical analysis. Next.js reads PostgreSQL using a server-only
query layer. Public GET handlers share those queries. Vercel hosts the website;
GitHub Actions runs CI and twice-daily ingestion, with no deployment permissions.

## Why this architecture
Vercel Services can host Next.js plus Go, but is beta and adds a request runtime
without improving this workload. Go computations run only when observations
change. Next.js never duplicates RSI, MACD, averages, bands or consensus logic.
Neon provides scale-to-zero PostgreSQL; Supabase is compatible but its additional
products are not needed. GitHub scheduling preserves two updates daily without
requiring Vercel Pro. Quotas and provider availability remain operational limits.

## Contracts and numerical semantics
Go owns the relational schema and persisted analytics JSON. Zod validates the
TypeScript read contract. Unknown values are null, not zero. SQL uses parameters
and explicit limits. Market lists carry only 31 daily samples per asset; detail
and comparisons retrieve at most 2,000 per selected asset. Comparisons cap at five.
Public reads cache for five minutes; no page view calls CoinGecko.

Analytics use ascending consecutive completed UTC-day samples. A sample is the
last price observed that day, not an exchange closing price or an OHLC candle.
Gaps reset the indicator calculation window. Current incomplete days are excluded.
RSI uses Wilder smoothing, MACD 12/26/9 EMA, averages 50/200 SMA, Bollinger 20/2
population standard deviations. Crossovers need a previous complete window.
Consensus renormalizes existing weights over available valid indicators. Heuristic
strength is not a calibrated probability. All raw historical data remains intact.

Normalized comparisons start at the first shared date; missing dates are whitespace
in charts. Presentation arithmetic for percentage normalization is not a second
implementation of technical analysis. All figures describe the tracked universe,
not the entire crypto market. Freshness becomes stale after 18 hours.

## Frontend
Server-rendered route shells provide crawlable headings and initial data. Client
islands implement charts, filters, watchlists and themes. Lightweight Charts loads
only for chart views; Recharts supplies the distribution visualization. Chart
attribution is visible. Radix Dialog supplies the mobile navigation focus trap.
The design uses graphite/teal dark and neutral light tokens, keyboard focus states,
reduced motion, semantic tables and a chart-data alternative. Browser storage is
versioned and holds only preferences and coin IDs; authentication is unnecessary.

## Failure modes
No configured database: unavailable state and health 503. Empty database: empty
state. Missing asset: 404. Missing history: unavailable indicator with required
sample count. Partial provider responses: durable partial run with successful
observations retained. Interrupted job: lock release and abandoned-run recovery.
GitHub schedules can be delayed or disabled after prolonged repository inactivity.
Observe workflow failures and the dashboard's ingestion/freshness status.
