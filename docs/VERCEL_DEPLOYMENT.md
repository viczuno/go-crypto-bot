# Deploy on Vercel

The repository is prepared for deployment; no production deployment is performed
automatically by this migration. A Vercel account and a provisioned database are
required. Review current Hobby eligibility and quotas for your intended usage.

## Provision and import
1. Create a Neon PostgreSQL project. Choose a region close to the Vercel function
   region; use the same region for preview and production where practical.
2. Create separate production and preview databases/branches. Follow DATABASE.md
   to run schema migrations, import the archive, verify it and configure roles.
3. In Vercel choose **Add New → Project → Import Git Repository**, select
   `viczuno/go-crypto-bot`, and set **Root Directory** to `web`.
4. Select **Next.js**, **Node.js 24.x**, install `npm ci`, build `npm run build`.
   Leave output directory at the framework default. `web/vercel.json` defines the
   framework/build settings. Do not add Services, cron, Pages or GitHub deploy jobs.
5. Configure `DATABASE_URL` with the pooled read-only Neon URL (TLS required).
   Scope production to the production database and preview to the preview database.
   Configure `SITE_URL` to the intended canonical origin. Credentials stay server-side.
6. Set production branch to `main`; the migration branch produces previews after
   Git integration is connected. The user controls the initial production deploy.
7. Verify `/api/health` returns 200, the dashboard reflects imported data, and coin
   detail/watchlist/comparison views work on desktop and mobile. Check Vercel logs.

The Next.js build does not require database credentials or mutate the schema.
Webpack is explicitly selected for reproducible builds in restricted environments.

## Scheduled ingestion
In GitHub repository **Settings → Secrets and variables → Actions**, add:
- `DATABASE_DIRECT_URL`: direct Neon URL for the restricted ingest role.
- `COINGECKO_API_KEY`: CoinGecko Demo key (sent as a header).

After user review and merge, the workflow on `main` runs at 00:17/12:17 UTC and can
also be manually dispatched. It never deploys or commits data. The workflow has
read-only repository access, a 12-minute timeout and concurrency control. Database
locking also protects against manual overlap. Do not put production secrets in PR
or fork test jobs. CI uses an isolated PostgreSQL service.

A draft PR does not activate a new schedule on the default branch. Verify the first
scheduled run, asset counts, provider timestamps and partial/failure reporting.
GitHub timing is best effort; it is not an exact-time market-data SLA.

## Launch and rollback checklist
- Verify archive import and back up PostgreSQL.
- Check preview with its separate database and read-only credentials.
- Confirm environment scopes and production branch.
- User merges and approves production activation.
- Verify production URL, health, data freshness and first scheduled ingestion.
- If necessary, use Vercel deployment rollback; preserve the additive schema.
- Restore a database backup into a new database only after explicit review.

No custom domain or public production URL is assumed. Record it after deployment
is actually verified. No production migration, merge, or launch is part of the
agent's automatic authorization.
