# CLAUDE.md

## Commands

This project uses [Task](https://taskfile.dev/) as the task runner.

- `task dev` — start both API (`:8080`) and UI (`:5173`) dev servers in parallel
- `task dev:server` — run Go API server only
- `task dev:ui` — run Vite UI dev server only
- `task test:ui` — run UI unit tests with Vitest
- `task lint:ui` — lint UI with ESLint
- `task format:check:ui` — check formatting with Prettier
- `task seed` — seed the database with development data
- `task hooks:install` — install Lefthook pre-commit hooks (one-time setup)

There are no Go test or lint tasks defined yet. Run Go tests directly with `go test ./...`.

## Architecture

Rota is an on-call rotation management app with a Go backend and a React/TypeScript frontend. The Vite dev server proxies `/api` requests to `http://localhost:8080`.

### IDs

Entity IDs use prefixed ULIDs: a short type prefix followed by an underscore and a 26-character ULID (e.g. `rot_01JQGF0000000000000000000` for a rotation).

### Backend (`internal/`)

Follows Domain Driven Design/clean architecture with four layers:

1. **Domain** (`internal/domain/`) — pure Go entities (`Rotation`, `RotationCadence`), the `RotationRepository` interface, and domain error types. No external dependencies.
2. **Application** (`internal/application/`) — use cases that orchestrate repository calls (e.g., `GetRotationUseCase`).
3. **Presentation** (`internal/presentation/httpapi/`) — HTTP handlers and JSON:API-formatted response DTOs. Handlers depend on application use cases.
4. **Infrastructure** (`internal/infrastructure/sqlite/`) — SQLite implementation of `RotationRepository`. Database auto-migrates via Goose on startup using embedded `.sql` files in `migrations/`.

### Frontend (`ui/src/`)

State is managed in `App.tsx` with `useState` and passed down as props. The timeline algorithm in `utils.ts` builds an 8-week forward schedule client-side by applying overrides over a base weekly rotation. Key types are in `types.ts`.

UI components use [Base UI](https://base-ui.com/) primitives with Tailwind CSS v4.
