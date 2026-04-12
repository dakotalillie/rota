# Stage 1: Build UI
FROM node:lts AS ui
WORKDIR /src
COPY ui/package.json ui/package-lock.json ui/
RUN cd ui && npm ci
COPY ui/ ui/
RUN cd ui && npm run build
# Vite outputs to /src/internal/ui/dist (outDir: ../internal/ui/dist)

# Stage 2: Build Go binary
FROM golang:1.26-bookworm AS builder
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /src/internal/ui/dist ./internal/ui/dist/
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o /rota cmd/server/main.go

# Stage 3: Runtime
FROM ubuntu:24.04
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /rota /app/rota
EXPOSE 8080
ENTRYPOINT ["/app/rota"]
