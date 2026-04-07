# Repository Guidelines

## Project Structure & Module Organization
`cmd/sage-wiki/` contains the CLI entrypoint. Core Go packages live under `internal/`, grouped by concern: `compiler/`, `extract/`, `query/`, `mcp/`, `web/`, `wiki/`, and related helpers such as `config/`, `storage/`, and `llm/`. End-to-end coverage lives in root-level tests such as `integration_test.go`. The optional browser UI lives in `web/`, with Preact source in `web/src/` and production assets embedded from `internal/web/dist/`.

## Build, Test, and Development Commands
Use Go for the main binary and Node.js only for the web UI.

- `go test ./...` runs all Go unit and integration tests.
- `go build ./cmd/sage-wiki` builds the CLI without the embedded web UI.
- `go build -tags webui -o sage-wiki ./cmd/sage-wiki` builds the full binary with embedded frontend assets.
- `cd web && npm install` installs frontend dependencies.
- `cd web && npm run dev` starts the Vite dev server for UI work.
- `cd web && npm run build` type-checks and builds the frontend bundle consumed by the Go server.

## Coding Style & Naming Conventions
Follow standard Go formatting with `gofmt`; keep packages small and cohesive, and prefer table-driven tests where useful. Exported Go names use `CamelCase`; unexported helpers use `camelCase`. In `web/src/`, use TypeScript + Preact function components, `PascalCase` component filenames such as `ArticleView.tsx`, and `camelCase` for hooks and utilities. Keep comments sparse and only where logic is non-obvious.

## Testing Guidelines
Go tests use the standard `testing` package and live beside the code as `*_test.go`. Add tests in the same package you modify, and extend `integration_test.go` when behavior crosses package boundaries. For frontend changes, at minimum run `npm run build` to catch TypeScript and bundle regressions; there is no separate frontend test suite checked in yet.

## Commit & Pull Request Guidelines
Recent history favors short, imperative subjects such as `Add unified TUI dashboard...`, with occasional conventional prefixes like `fix:`. Keep the first line under roughly 72 characters when possible and describe the user-visible change. PRs should include a concise summary, linked issue if applicable, test/build notes, and screenshots or GIFs for `web/` or TUI changes.

## Security & Configuration Tips
Never commit API keys or generated wiki content with secrets. Keep credentials in environment variables referenced by `config.yaml` (for example `GEMINI_API_KEY`). When touching file-serving or ingestion code, preserve the repository’s path traversal and input validation safeguards.
