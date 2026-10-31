# Development instructions

Read AGENTS.md for engineering and authorization rules. The current architecture
is documented in docs/ARCHITECTURE.md and operational instructions in docs/DATABASE.md
and docs/VERCEL_DEPLOYMENT.md. Go owns ingestion and technical analysis; Next.js
reads PostgreSQL and supplies the interface. Never duplicate indicators in TS.

Run `make test`, `make lint`, `make build`, and relevant Playwright tests before
committing application changes. Database tests require a dedicated migrated test
DB. Do not run live ingestion as a substitute for deterministic tests. Preserve
archive bytes. Never deploy production or merge a PR automatically.
