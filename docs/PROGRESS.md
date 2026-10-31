# Resume checkpoint — 2026-10-31

## Branch and implementation
Continue `feat/vercel-migration`. Do not restart or regenerate the application.
Implementation is Next.js 16.4.3 / React 19.3.1 / strict TypeScript in web, with
Go pgx/Goose persistence, analytics and scheduled ingestion. Production target is
Vercel with Neon. All changes originated in this migration; no unrelated user
modifications were found during resume.

Completed repository-side work:
- Schema, local PostgreSQL tooling and typed read APIs.
- Historical import with source provenance, checksums and reconciliation.
- Corrected daily Go indicators and partial/idempotent ingestion.
- Dashboard/explorer/details, charts, watchlists, comparisons, dark/light themes.
- Legacy website pipeline removed; immutable inputs moved into archive.
- CI, scheduled ingestion, maintained README and deployment/database guides.

Validation evidence and gaps: see VALIDATION.md. All current Go checks, frontend
unit/static/build checks and four desktop/mobile browser tests passed. Historical
verification returns 2,716 source rows / 2,256 canonical observations / 16 assets.

## Next work, in order
1. Inspect remote CI results for this branch/draft PR; fix failures if present.
2. Expand Phase 5 tests for empty/unavailable datasets and accessibility on
   details/comparison. Lock contention and daily-gap regression tests now pass. Do not rerun live CoinGecko ingestion
   merely to test code.
3. User provisions Neon and Vercel; follow DATABASE.md and VERCEL_DEPLOYMENT.md.
   Use separate preview/prod databases and read-only web credentials.
4. Verify preview, external historical import, and role permissions. User controls
   merge and production activation. Verify the first scheduled run after merge.

## Local environment used
Node: /opt/homebrew/opt/node@24/bin; PostgreSQL: /opt/homebrew/opt/postgresql@17/bin.
Existing ignored cluster: .local/postgres on 127.0.0.1:55432, user crypto.
Database crypto contains archived data plus one successful live ingestion.
Database crypto_test is dedicated to integration tests. Use CRYPTO_TEST_DATABASE=1
and its DATABASE_DIRECT_URL explicitly. Neither database is production.
Go commands do not load .env automatically. Frontend reads web/.env.local or env.
Do not commit .local, credentials, node_modules, .next or test artifacts.

## External status
No production deployment, external migration or merge has been performed.
Feature push, draft PR and repository description/topics updates are authorized.
Draft PR: https://github.com/viczuno/go-crypto-bot/pull/10

Commits: 5da716d (data), 0ea7ffe (web), 99f3aa1 (CI/cleanup/docs).
Repository description/topics updated to implemented features. Remote CI run
41179337887 was in progress at this checkpoint; no remote green claim is made.
The follow-up test commit adds overlapping-worker rejection, gap-reset and
last-sample-per-day regression coverage, validated uncached against crypto_test.

## Preview market-read investigation
Sanitized diagnostics added for market/detail reads: stage, SQLSTATE category/code,
and allowlisted Zod issue paths only. No error messages, SQL, row values or URLs.
Existing local imported database and the configured Neon database both returned
ready with the actual query/schema; the packaged production server also returned
HTTP 200/ready for that Neon connection. SSL warning was reproduced without read
failure. Preview may use different credentials/settings; its exact failure has
not yet been classified. Do not claim a root cause without its sanitized event.
Next action: inspect market_read_failed from the new Preview after the diagnostic
commit. 42501=permissions, 42*=SQL/schema, 22*=database values, validation=Zod paths.
No database writes, reset, migrations or reimport were performed in this investigation.

## Preview follow-up: numeric types and failure recovery
Read-only inspection of the configured Neon connection confirmed change_24h is
DOUBLE PRECISION (OID 701), all 2,256 imported observations have NULL change_24h,
and SELECT permission is present on assets, observations, daily_samples, analytics,
and ingestion_runs. No schema/value mismatch was found: the actual snapshot SQL
and all imported analytics records pass the existing Zod schemas. This does not
prove Preview uses the same connection/role. Preview is access-protected (302),
and no authenticated Vercel log access is configured in this workspace.

Confirmed application defect fixed: unavailable fallbacks were cached as successful
snapshots for five minutes. Reads now throw a sanitized marker inside the cache;
the unavailable response is constructed outside, so the next request retries.
Successful snapshots still cache normally. The SQL now explicitly casts change_24h
to float8 and has shared real-PostgreSQL coverage for NUMERIC and DOUBLE PRECISION
values, NULL market metadata, and imported analytics. No schema coercion or history
rewrites were needed. Cache namespace v3 avoids prior cached unavailable results.

Validation: 20 tests passed against the local imported database and configured
Neon database (read-only), plus lint/typecheck and production build. Sanitized
logging from f728d85 remains active. The original Preview-triggering exception
still needs its full market_read_failed event (stage/category/sqlstate/issues);
do not claim the missing cast caused the imported-data outage.
