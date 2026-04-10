# Rota

A simple on call application.

## Requirements

- [Go](https://golang.org/) 1.26+
- [Node.js](https://nodejs.org/) (for the UI)
- [Task](https://taskfile.dev/)
- [Lefthook](https://github.com/evilmartians/lefthook) (for git hooks)

## Getting started

Install git hooks (one-time, after cloning):

```sh
task install:hooks
```

Run both UI and API dev servers:

```sh
task dev
```

Run the tests:

```sh
task test:server  # Go tests
task test:ui      # UI tests
```

Format Go code:

```sh
task format:server
```

Run the linters:

```sh
task lint:server  # Go (golangci-lint)
task lint:ui      # UI (ESLint)
```

After intentionally changing an API response, regenerate the snapshot golden files:

```sh
task update-snapshots
```

Open the UI at http://localhost:5173. Any requests to /api are proxied to the
API server.
