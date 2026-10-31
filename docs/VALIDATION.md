# Validation checkpoint — 2026-10-31

## Executed successfully
- `go test -race ./cmd/... ./internal/...`
- Uncached database check: `CRYPTO_TEST_DATABASE=1 DATABASE_DIRECT_URL=<isolated-test-db> go test -count=1 -race ./internal/platform`
- `go vet ./cmd/... ./internal/...` and `go build ./cmd/... ./internal/...`
- `go run ./cmd/import-history -mode verify` against the existing local imported database:
  2,716 source records, 2,256 canonical observations, 16 assets, verified true.
- Frontend lint, strict TypeScript, Prettier checks, six Vitest tests and production
  Webpack build. No production credentials required for the build.
- Four Playwright checks: desktop/mobile dashboard, search, watchlist persistence,
  coin overlays/timeframes, comparison, theme switching, browser errors, automated
  explorer accessibility scan, invalid API bounds and unknown asset handling.
- Earlier local CoinGecko ingestion completed successfully; no production writes.

## Corrections during validation
Aligned indicator fixtures with ascending daily samples; corrected flat-series
floating point cancellation; fixed theme hydration; bounded summary chart reads;
kept detail history separate from the cached market snapshot; validated comparison
responses and cleared stale comparisons after request failure. Browser selectors
wait for the streamed coin view to settle before choosing an overlay.

## Limits and remaining verification
- CI is configured but remote results must be checked after the branch push.
- Empty/unavailable database states exist but lack dedicated end-to-end scenarios.
- Automated accessibility coverage currently scans the explorer only; expand to
  coin/compare and keyboard navigation before declaring a complete accessibility audit.
- Explicit advisory-lock contention, daily-gap reset and last-sample-per-day
  regressions now pass uncached in PostgreSQL, along with repeated import,
  ingestion idempotency and partial runs.
- Preview/prod Neon roles, external import, Vercel routing, and scheduled default-
  branch ingestion require provisioning and user-controlled activation.
- Vite emits a config-loader compatibility warning; tests pass. Turbopack process
  binding failed locally; the supported Webpack build is selected explicitly.

Screenshots in `docs/screenshots/` show actual local data, not a live production
website. The archive remains immutable migration input.
