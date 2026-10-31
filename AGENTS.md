# Repository development rules

- Read README.md for setup/operations and docs/ARCHITECTURE.md for data semantics.
- Go owns ingestion and indicators; TypeScript reads PostgreSQL and renders UI.
  Never duplicate indicators or treat observations as daily candles.
- Preserve archive bytes and migration provenance. Recovery tools are not runtime
  data sources. Tests use a disposable local DB, never production credentials.
- Inspect git status/remote; preserve unrelated work. Use feature branches named
  `<type>/<kebab-case-summary>`. Stage explicit paths and review the full diff.
- Use Conventional Commits and specific PR titles. PR descriptions explain the
  behavior, architecture, checks actually run, risks and deployment needs.
- Push feature branches/open draft PRs within task authorization. Merge, production
  activation, metadata changes and force-push require explicit task authorization.
- Use idiomatic Go, gofmt, small interfaces, context cancellation, wrapped errors,
  parameterized SQL and bounded requests. Avoid speculative abstractions.
- Keep TypeScript strict; validate external data, preserve nulls, avoid any and
  unchecked assertions. Use accessible responsive UI and reduced-motion support.
- Keep secrets server-side and out of logs. SQLSTATE/allowlisted validation paths
  are safe; raw SQL, parameters, connection errors and row values may not be.
- Run make test, make lint and make build for code changes. PostgreSQL integration
  requires CRYPTO_TEST_DATABASE=1 and a dedicated migrated test database. Run
  Playwright for affected UI flows on desktop/mobile and dark/light themes.
- Record skipped/failed checks accurately. Documentation-only changes need accuracy,
  link and whitespace checks. Never claim unexecuted checks passed.
- Use the installed gh skill for GitHub, structured output and explicit repo scope.
- Vercel is the only website host. GitHub Actions performs CI and scheduled Go
  ingestion; it never deploys the website or commits market data.
