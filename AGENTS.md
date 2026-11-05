# Repository Guidelines

## Project Structure & Module Organization

This workspace contains a Next.js frontend in `FE/` and a Go Modulith backend in `BE/`. Read each directory's `AGENTS.md` before editing it. The `.agents/`, `.codex/`, and `.aws/` directories are reserved for environment configuration; do not use them for application code.

When adding the initial implementation, group source files by feature or module and keep tests close to the code they verify, or in a clearly named `tests/` directory. Place static assets in an explicit directory such as `assets/`. Add a `README.md` describing the chosen structure and technology stack.

## Build, Test, and Development Commands

Run frontend commands inside `FE/`: `npm ci`, `npm run dev`, `npm run build`, `npm test`, `npm run lint`, `npm run typecheck`. Run backend commands inside `BE/`: `go run ./cmd/api`, `go build ./...`, `go test -race ./...`, `go vet ./...`. See the root and module READMEs for prerequisites and exact configuration.

When introducing tooling, document exact installation, development, build, and test commands in `README.md`. Commit the relevant manifests and lockfiles so contributors can reproduce the setup.

## Coding Style & Naming Conventions

Use the frontend's ESLint/Prettier configuration and `gofmt`/`go vet` for Go. Keep module ownership boundaries and avoid unrelated formatting changes.

Introduce formatter or linter configuration with the first implementation and document how to run it. Keep filenames consistent within each module.

## Testing Guidelines

Frontend tests use Node's built-in test runner; Go tests use `testing` and `httptest`, including race and module-boundary checks. Backend integration tests require a dedicated PostgreSQL `TEST_DATABASE_URL` (`make test`); without it they skip. No coverage threshold is defined. Add tests alongside behavior changes and document how to run them.

## Commit & Pull Request Guidelines

Git history is unavailable in this workspace, so existing commit conventions cannot be verified. Use concise, imperative commit subjects, for example `Add project setup documentation`, and keep each commit focused.

Pull requests should explain the change, link relevant issues, and list validation performed. Include screenshots for visible interface changes. Explicitly state when tests could not be run and why.

## Security & Configuration

Never commit credentials, access tokens, or local environment files. Provide placeholder configuration examples when needed and document required environment variables without real secret values.

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `FE/node_modules/next/dist/docs/` before changing Next.js code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
