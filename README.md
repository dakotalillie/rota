# Rota

A simple on call application.

## Requirements

- [Go](https://golang.org/) 1.26+
- [Node.js](https://nodejs.org/) (for the UI)
- [Task](https://taskfile.dev/)

## Getting started

Run both UI and API dev servers:

```sh
task dev
```

Open the UI at http://localhost:5173. Any requests to /api are proxied to the
API server.
