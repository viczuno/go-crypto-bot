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
2. Expand Phase 5 tests for empty/unavailable datasets, lock contention, daily gaps,
   and accessibility on details/comparison. Do not rerun live CoinGecko ingestion
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
Record the final PR URL in this document after creation.
