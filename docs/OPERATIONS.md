# Operations and recovery runbook

README.md owns setup, commands, environment variables and deployment instructions.
This document covers exceptional operations, recovery and release evidence.

## Roles and environment boundaries

Use an admin/owner role only for reviewed schema changes and recovery imports.
Adapt migrations/roles.sql.example to the actual database name (Neon commonly uses
neondb), and set passwords through the provider UI or psql's password prompt.
The web role gets SELECT on assets, observations, daily_samples, analytics and
ingestion_runs. The ingest role gets SELECT/INSERT/UPDATE and sequence usage on
its pipeline tables. No browser credential or public administrative endpoint exists.

Production, Preview, development and test datasets need separate environment scopes.
Do not infer a secret's destination from its name: verify the host/database/role
privately in the provider UI. GitHub exposes secret names, not their values.

## Backups and restoration

Monitor Neon storage/compute and retain an independent logical backup according to
your recovery requirements. With an authorized admin connection configured locally:

```sh
# Writes a local backup file, not database records. Keep it outside Git.
pg_dump --dbname="$DATABASE_DIRECT_URL" --format=custom --file=/safe/path/crypto.dump
# Recovery example: restore only into a NEW, explicitly reviewed database.
# pg_restore --dbname="$RECOVERY_DATABASE_URL" /safe/path/crypto.dump
```

Verify backups by restoring into a separate database and reconciling counts,
time ranges and representative values. Provider restore windows depend on plan;
do not treat an untested provider snapshot as the only recovery mechanism.

The immutable archive and historical importer remain because no independently
verified recoverable backup has been demonstrated. They are recovery tools, never
runtime persistence. Original baseline: 2,716 source records, 2,256 canonical
observations, 16 assets. SQLite has 150 exact duplicate groups; JSON contributes
eleven extra observations. Raw timestamp text preserves original nanoseconds.

```sh
shasum -a 256 -c archive/SHA256SUMS
go run ./cmd/import-history                    # dry-run, no database
# Use a new/reviewed recovery database before the following:
go run ./cmd/import-history -mode import       # additive transaction
go run ./cmd/import-history -mode verify       # checks each source record
```

Import rejects invalid/conflicting values and verifies source identity, checksums,
asset, raw/normalized timestamps and price against canonical observations. Reruns
are idempotent. Do not use a different source snapshot with the same provenance IDs
without explicit reconciliation. archive/prelaunch preserves the last legacy main
snapshot separately; it is not imported automatically and does not replace the
verified original archive. Schema down-migration deliberately refuses to drop
history. Prefer restoring into a new database over destructive rollback.

## Safe release gates

Before merge: green CI on the current SHA, Ready Preview, successful health and
representative market/detail/chart/watchlist/comparison checks. User verification
of the working Preview is valid evidence; independently inspect configuration
where access permits. Confirm Vercel root web, Next.js, Node 24, main production
branch, intended initialized Production DB and separated Preview credentials.
Confirm both Actions secrets and the ingest connection destination/role.

Do not merge while a required gate fails or cannot be confirmed. Do not bypass
branch protection. A merge to main triggers Vercel production deployment and makes
the scheduled ingestion workflow eligible; do not dispatch additional ingestion
merely to smoke-test the release.

After merge: check PR merge state/main SHA, Vercel production Ready, health and
representative pages, workflow presence and absence of legacy deployment jobs.
Do not claim live success from a deployment build status alone. If release fails,
inspect logs and roll back website code through Vercel; retain the additive schema.
Database restoration or external migrations require separately reviewed intent.
