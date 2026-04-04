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
task hooks:install
```

Run both UI and API dev servers:

```sh
task dev
```

Open the UI at http://localhost:5173. Any requests to /api are proxied to the
API server.
