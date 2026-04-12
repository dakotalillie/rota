# Rota

Rota is an on-call rotation management application. Create rotations with configurable schedules, manage team members, see who is currently on call, and set temporary overrides. Built with a Go backend, React/TypeScript frontend, and SQLite database. The production build produces a single binary with the UI embedded.

## Running with Docker

```sh
docker run -p 8080:8080 -v rota-data:/data -e DATABASE_PATH=/data/rota.db dakotalillie/rota
```

Open http://localhost:8080.

## Running with Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  rota:
    image: dakotalillie/rota:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_PATH=/data/rota.db
    volumes:
      - rota-data:/data

volumes:
  rota-data:
```

Then start the application:

```sh
docker compose up -d
```

Open http://localhost:8080. To stop:

```sh
docker compose down
```

## Development

### Requirements

- [Go](https://golang.org/) 1.26+
- [Node.js](https://nodejs.org/)
- [Task](https://taskfile.dev/)
- [Lefthook](https://github.com/evilmartians/lefthook) (for git hooks)

### Getting started

Install git hooks (one-time, after cloning):

```sh
task install:hooks
```

Run both UI and API dev servers:

```sh
task dev
```

Open the UI at http://localhost:5173. Requests to `/api` are proxied to the API server.

### Testing

```sh
task test:server  # Go tests
task test:ui      # UI tests
task test:e2e     # E2E tests (requires Docker)
```

### Formatting

```sh
task format:server  # Go (goimports)
task format:ui      # UI (Prettier)
```

### Linting

```sh
task lint:server  # Go (golangci-lint)
task lint:ui      # UI (ESLint)
```

### Updating snapshots

After intentionally changing an API response, regenerate the snapshot golden files:

```sh
task update-snapshots
```

## Environment variables

| Variable | Description | Default |
|---|---|---|
| `PORT` | HTTP listen port | `8080` |
| `HOSTNAME` | Base URL used for JSON:API links | `http://localhost:<PORT>` |
| `DATABASE_PATH` | Path to the SQLite database file | `rota.db` |
| `LOG_LEVEL` | Minimum log level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Log output format (`text`, `json`) | `text` |

## API

The backend exposes a JSON:API-compliant REST API at `/api`.

```
POST   /api/rotations
GET    /api/rotations
GET    /api/rotations/{rotationID}
DELETE /api/rotations/{rotationID}
POST   /api/rotations/{rotationID}/members
PUT    /api/rotations/{rotationID}/members
DELETE /api/rotations/{rotationID}/members/{memberID}
GET    /api/rotations/{rotationID}/schedule
POST   /api/rotations/{rotationID}/overrides
DELETE /api/rotations/{rotationID}/overrides/{overrideID}
```
