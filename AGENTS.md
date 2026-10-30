# Repository engineering instructions

## Scope and repository context

These instructions apply throughout the repository. Read any more specific
`AGENTS.md` before changing files in its directory. Explicit user instructions
take precedence. `CLAUDE.md` contains supplementary architecture and domain notes;
keep shared workflow guidance consistent with this file.

The current application is a Go bot with SQLite storage, CoinGecko integration,
technical indicators, generated market data, and a Hugo website. Go's required
version is declared in `go.mod`. There is currently no TypeScript package or
Node lockfile; apply the TypeScript standards below when one is introduced.

## Git workflow and authorization

- Check `git status --short --branch` and the remote before starting. Preserve
  unrelated user changes and never commit credentials or tokens.
- Work on feature branches, never commit development changes directly to `main`.
  Use `<type>/<short-kebab-case-description>`, such as `feat/vercel-migration`,
  `fix/api-timeout`, `docs/contributing`, or `chore/dependency-update`.
- For a new branch, fetch and branch from the latest `origin/main`. Continue an
  existing task branch when appropriate; do not reset or rebase user work without
  authorization.
- Use meaningful Conventional Commits: `<type>(<optional-scope>): <summary>`.
  Types include `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`, `ci`,
  and `chore`. Use an imperative, specific summary, for example
  `docs(dev): establish repository engineering standards`. Mark breaking changes
  with `!` and explain them in a `BREAKING CHANGE:` footer.
- Keep commits focused. Review the diff and stage explicit paths, rather than
  including unrelated changes with `git add .`.
- Feature-branch pushes and draft PR creation are authorized unless the current
  task restricts them. Never merge PRs automatically, enable auto-merge, push
  directly to `main`, or force-push without explicit user authorization.
- Initial repository setup is local only: do not push, create a PR, or change
  GitHub metadata/settings as part of that setup. Report findings and permission
  gaps instead.

## Go standards

- Follow idiomatic Go: format changed Go files with `gofmt`, use clear package
  names and small cohesive functions, and document exported APIs where useful.
- Preserve the existing boundaries: `cmd` for composition, `internal/domain` for
  shared contracts, and focused packages for API, storage, indicators, and output.
  Prefer small interfaces at dependency boundaries and concrete types internally;
  avoid speculative abstractions and unnecessary dependencies.
- Return errors with operation context and wrap causes with `%w`. Use
  `errors.Is`/`errors.As` where appropriate; do not silently discard errors or
  panic for expected failures.
- Pass `context.Context` first for I/O, propagate cancellation and deadlines,
  bound HTTP calls and retries, close resources, and use parameterized SQL.
  Handle SQLite writes and transactions deliberately.
- Keep business logic deterministic and testable. Use UTC for stored timestamps.
  Test edge cases in indicator calculations and preserve historical data.
- Use table-driven tests where they clarify cases, existing function-based mocks,
  `httptest` for HTTP, and temporary databases/files. Tests must not require live
  CoinGecko access or mutate tracked data.

## TypeScript and website standards

- When adding TypeScript, enable strict checking. Prefer explicit domain types,
  discriminated unions, and `unknown` with narrowing over `any` or unchecked
  assertions. Validate external JSON at runtime.
- Keep modules cohesive and dependencies minimal. Follow the selected framework's
  conventions; separate data access from presentation and handle loading, empty,
  and error states deliberately.
- Use semantic HTML, accessible labels and keyboard interactions, responsive
  layouts, and safe output escaping. Keep secrets in server-side environment
  variables; never expose them through browser bundles or public-prefixed values.
- Use one package manager and commit its lockfile. Define reproducible formatting,
  lint, typecheck, test, and production-build scripts when the package is added.
  Use a frozen/immutable lockfile install in CI.
- Test meaningful behavior and regressions, including accessibility and critical
  user flows when relevant. Avoid tests that merely duplicate implementation.

## Validation before commits

- Review the full diff, run `git diff --check`, and verify the staged diff before
  committing. Confirm the branch and ensure only intended files are staged.
- For Go changes, format changed files and run `go test -race ./...`,
  `go vet ./...`, and `go build ./...`. `make test`, `make vet`, and `make build`
  provide related repository targets. Run `golangci-lint run` directly when
  available: the current `make lint` target masks lint failures, so its exit
  status alone is not evidence of a passing lint check.
- For TypeScript changes, run the package's formatting check, lint, typecheck,
  tests, and production build using its committed lockfile and scripts. Do not
  claim these passed before a TypeScript toolchain exists.
- For Hugo/template/config changes, also build the site with `hugo --minify`
  and review affected pages. For documentation-only changes, check accuracy,
  consistency, and whitespace; code tests are needed only when behavior changes.
- Report exact checks and outcomes. Investigate failures; if tooling, permissions,
  or pre-existing problems block validation, report the limitation before
  committing and in the PR. Never describe skipped checks as passing.
- Do not run `make run` or the bot merely to validate code: it fetches live data
  and rewrites `README.md`, SQLite, and exported JSON. Do not hand-edit generated
  market data or include incidental generated changes.

## GitHub pull requests

- Use the installed `gh` skill when working with GitHub. Verify `gh auth status`
  and target `viczuno/go-crypto-bot` explicitly for repository operations. Prefer
  structured JSON/API output and paginate lists as needed.
- Open draft PRs against `main` after validation and an authorized feature-branch
  push. Use a clear Conventional Commit title describing the resulting behavior,
  such as `feat(web): deploy the dashboard on Vercel`.
- Write the description for a reviewer without conversation context: explain the
  concrete problem, resulting behavior, key design choices, validation performed
  and results, and material risks or remaining work. Link relevant issues and
  include screenshots or preview URLs for visible changes when available.
- Scale detail to the change. Avoid vague titles, unverified test claims, and
  conversational logs. Use `--body-file` for multiline PR descriptions.
- Leave review and merging to the user. Do not modify repository description,
  topics, merge settings, or branch protection without task authorization.

## Deployment policy

- Vercel is the exclusive website deployment platform for previews and production.
  Do not add or expand GitHub Pages or another website host.
- GitHub Actions may continue collecting data, testing, and building artifacts;
  website releases must target Vercel. Keep deployment credentials in managed
  secrets and environment variables.
- Migration debt at setup: `.github/workflows/bot.yaml` still deploys to GitHub
  Pages, and `hugo.toml` and the generated README still use the Pages URL. These
  are legacy configuration, not the approved deployment architecture. Replace
  them during the Vercel migration, updating generated links at their source.
- This setup establishes policy only. Do not trigger a deployment or claim the
  Vercel migration is complete without configuration and successful validation.
