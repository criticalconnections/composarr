# Composarr

A Docker Compose stack lifecycle manager that fills the gap between clicking around in Portainer and writing Ansible playbooks. Every compose file is version-controlled in git, every deploy is diff-previewed, every update is health-verified with auto-rollback, and maintenance windows let you stage changes for off-hours.

## Features

- **Git-backed versioning** — every compose file change is auto-committed to a local per-stack git repository. Full history, one-click rollback to any version
- **Diff before deploy** — side-by-side YAML diff of changes between HEAD and working copy before you apply them
- **Health-check verification** — after `docker compose up -d`, Composarr polls container health (native `HEALTHCHECK` or running grace period). If anything is unhealthy within the timeout, the deploy auto-rolls back to the last successful version
- **Scheduled maintenance windows** — cron expression + duration. Queue compose updates to deploy automatically at the next window
- **Cross-stack dependencies** — declare that stack A requires stack B. Composarr verifies the dependency is running before deploying, and prevents cyclic dependencies
- **Real-time UI** — WebSocket-driven deploy timeline, live log streaming, toast notifications for every deploy lifecycle event
- **Dependency graph** — visualize your entire stack topology as a DAG

## Quick start (Docker)

```bash
docker compose up -d
```

This builds Composarr and starts it on port 8080 with its data directory as a Docker volume. Open http://localhost:8080.

To deploy against a real Docker daemon, Composarr mounts the host's Docker socket (`/var/run/docker.sock`). You can customize the config via environment variables in `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| `COMPOSARR_PORT` | `8080` | HTTP port |
| `COMPOSARR_DATA_DIR` | `/data` | Where the SQLite DB and per-stack git repos live |
| `COMPOSARR_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `COMPOSARR_HEALTH_TIMEOUT` | `120` | Seconds to wait for post-deploy health |
| `COMPOSARR_HEALTH_INTERVAL` | `5` | Seconds between health polls |
| `TZ` | `UTC` | Timezone for scheduler |

## Development setup

Prerequisites:
- **Go 1.23+** (tested on 1.26)
- **Node.js 22+**
- **Docker + Docker Compose** (for running stacks)

```bash
# Clone and install
git clone <repo> composarr
cd composarr

# Backend dependencies
go mod download

# Frontend dependencies
cd web && npm install && cd ..
```

### Run in dev

In two terminals:

```bash
# Terminal 1 — backend on :8080
go run ./cmd/composarr

# Terminal 2 — frontend dev server on :5173 (proxies /api to :8080)
cd web && npm run dev
```

Open http://localhost:5173.

### Build the production binary

```bash
# Builds frontend into internal/handler/static/, then embeds into the Go binary
make build
./bin/composarr
```

## Architecture

- **Backend** — Go + Gin HTTP, SQLite (modernc.org/sqlite for CGo-free builds), go-git for versioning, gorilla/websocket for push, robfig/cron for scheduling
- **Frontend** — React 19 + TypeScript + Vite + Tailwind, TanStack Query for server state, Monaco editor for YAML, `react-diff-viewer-continued` for diffs
- **Compose interaction** — shells out to the `docker compose` CLI (handles full spec natively), parses `ps --format json` for container state and health
- **Git strategy** — one repo per stack at `data/repos/<slug>/`, single `main` branch, forward-only commits. Rollback = new commit with old content. History is never rewritten

### Directory layout

```
composarr/
├── cmd/composarr/        # entrypoint
├── internal/
│   ├── config/           # env-based config
│   ├── database/         # SQLite + embedded migrations
│   ├── models/           # DB structs
│   ├── repository/       # CRUD per table
│   ├── service/          # business logic
│   │   ├── docker_service.go     # docker compose CLI wrapper
│   │   ├── git_service.go        # go-git versioning
│   │   ├── deploy_service.go     # deploy pipeline orchestrator
│   │   ├── health_service.go     # post-deploy health polling
│   │   ├── scheduler_service.go  # cron + maintenance windows
│   │   └── dependency_service.go # cross-stack DAG
│   ├── handler/          # Gin HTTP handlers
│   └── websocket/        # push hub + client pumps
├── web/                  # React SPA (Vite)
└── data/                 # runtime: SQLite DB + per-stack git repos
```

## API overview

All endpoints under `/api/v1`. See `internal/handler/router.go` for the full list.

**Stacks** — `GET/POST /stacks`, `GET/PUT/DELETE /stacks/:id`, `GET/PUT /stacks/:id/compose`, `POST /stacks/:id/start|stop|restart|deploy`, `GET /stacks/:id/status|logs`

**Versions** — `GET /stacks/:id/versions`, `GET /stacks/:id/versions/:hash/diff`, `POST /stacks/:id/versions/:hash/rollback`

**Deployments** — `GET /deployments`, `GET /deployments/:id`, `POST /deployments/:id/cancel`

**Schedules** — `GET/POST /stacks/:id/schedules`, `PUT/DELETE /schedules/:id`, `POST /stacks/:id/queue`, `GET /schedules/upcoming`

**Dependencies** — `GET /dependencies/graph`, `POST/DELETE /stacks/:id/dependencies`

**WebSocket** — `GET /api/v1/ws/events` (deploy lifecycle events, health updates)

## Makefile targets

| Target | Description |
|--------|-------------|
| `make build` | Build frontend + Go binary |
| `make dev` | Run Go backend only |
| `make dev-frontend` | Run Vite dev server |
| `make test` | Go tests |
| `make docker` | Build Docker image |
| `make up` / `make down` | Docker compose up/down (self-deploy) |
| `make clean` | Remove build artifacts |

## License

MIT
"# composarr" 
