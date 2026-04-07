Rota is an on-call rotation management app with a Go backend and a React/TypeScript frontend.

## Shared Instructions

- Follow Domain-Driven Design and clean architecture patterns in the backend.
- When making backend changes, add or update tests as part of the same change.
- Keep the presentation layer in `internal/presentation/httpapi/` compliant with JSON:API.
- Prefer commands defined in `Taskfile.yml` for formatting, linting, and testing.
- After intentionally changing an API response, regenerate snapshot golden files with `task test:server:update-snapshots`.

## Preferred Verification

- Use `task test:server` for Go tests.
- Use `task test:ui` for UI tests.
- Use `task format:server` for Go formatting.
- Use `task lint:server` and `task lint:ui` for linting.
