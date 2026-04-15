# Composarr

Service/Stack Lifecycle Manager for Docker Compose stacks.

## Tech Stack
- **Backend**: Go 1.23 + Gin + SQLite (modernc.org/sqlite) + sqlx
- **Frontend**: React 19 + TypeScript + Vite + Tailwind CSS + TanStack Query
- **Docker**: Shells out to `docker compose` CLI; Docker Go SDK for health inspection
- **Git**: go-git (pure Go) for compose file versioning

## Project Structure
- `cmd/composarr/main.go` - Application entrypoint
- `internal/config/` - Environment-based configuration
- `internal/database/` - SQLite setup + embedded SQL migrations
- `internal/models/` - Go structs (Stack, Deployment, Schedule, etc.)
- `internal/repository/` - Database CRUD operations
- `internal/service/` - Business logic (stack, docker, git, deploy, health, scheduler)
- `internal/handler/` - Gin HTTP handlers + router
- `web/` - React SPA (Vite)

## Commands
- `make dev` - Run Go backend in dev mode
- `make dev-frontend` - Run Vite dev server (port 5173, proxies API to 8080)
- `make build` - Build frontend + Go binary
- `make docker` - Build Docker image
- `make test` - Run Go tests
- `make up` / `make down` - Docker compose up/down for self-deploy

## Architecture Notes
- OneDrive-synced directory: use PowerShell for git/filesystem mutations from bash
- Frontend embedded in Go binary via `//go:embed` for production
- Development: Vite dev server proxies `/api` to Go backend
- One git repo per stack at `data/repos/<slug>/`
- Forward-only git commits (rollback = new commit with old content)
