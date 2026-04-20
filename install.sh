#!/usr/bin/env bash
set -euo pipefail

# Composarr installer — supports a local (build-from-source) install or a
# Docker-based install. Run non-interactively with --mode=local|docker or
# interactively (the script will prompt).

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_DIR"

MODE=""
PORT="8080"
START_AFTER_INSTALL="ask"

usage() {
    cat <<'EOF'
Usage: ./install.sh [options]

Options:
  --mode=<local|docker>   Install mode (skip the interactive prompt)
  --port=<port>           Host port for Composarr (default: 8080)
  --start                 Start Composarr after installing
  --no-start              Do not start Composarr after installing
  -h, --help              Show this help

Modes:
  local   Build the Go binary + frontend on this machine. Requires Go 1.23+,
          Node.js 22+, and npm. Produces ./bin/composarr.
  docker  Build a Docker image and run via docker compose. Requires Docker
          and the Docker Compose plugin.
EOF
}

for arg in "$@"; do
    case "$arg" in
        --mode=*) MODE="${arg#*=}" ;;
        --port=*) PORT="${arg#*=}" ;;
        --start) START_AFTER_INSTALL="yes" ;;
        --no-start) START_AFTER_INSTALL="no" ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $arg" >&2; usage; exit 2 ;;
    esac
done

log()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m!!\033[0m  %s\n' "$*" >&2; }
err()  { printf '\033[1;31mXX\033[0m  %s\n' "$*" >&2; }

have() { command -v "$1" >/dev/null 2>&1; }

prompt_mode() {
    echo "How would you like to install Composarr?"
    echo "  1) Docker       — build an image and run via docker compose (recommended)"
    echo "  2) Local build  — compile the Go binary and frontend on this machine"
    while true; do
        read -r -p "Select [1/2]: " choice
        case "$choice" in
            1|docker|Docker) MODE="docker"; return ;;
            2|local|Local)   MODE="local";  return ;;
            *) echo "Please enter 1 or 2." ;;
        esac
    done
}

confirm_start() {
    if [ "$START_AFTER_INSTALL" != "ask" ]; then return; fi
    read -r -p "Start Composarr now? [Y/n]: " ans
    case "${ans:-Y}" in
        y|Y|yes|Yes|YES) START_AFTER_INSTALL="yes" ;;
        *)               START_AFTER_INSTALL="no" ;;
    esac
}

check_local_prereqs() {
    local missing=0
    if ! have go; then
        err "Go is not installed. Install Go 1.23+ from https://go.dev/dl/"
        missing=1
    else
        local gover
        gover="$(go version | awk '{print $3}' | sed 's/^go//')"
        log "Found Go $gover"
    fi
    if ! have node; then
        err "Node.js is not installed. Install Node.js 22+ from https://nodejs.org/"
        missing=1
    else
        log "Found Node.js $(node --version)"
    fi
    if ! have npm; then
        err "npm is not installed (usually bundled with Node.js)."
        missing=1
    fi
    if ! have make; then
        warn "make is not installed; the script will run build steps directly."
    fi
    [ "$missing" -eq 0 ] || exit 1
}

check_docker_prereqs() {
    if ! have docker; then
        err "Docker is not installed. See https://docs.docker.com/engine/install/"
        exit 1
    fi
    if ! docker compose version >/dev/null 2>&1; then
        err "Docker Compose plugin not found. Install the 'docker compose' plugin."
        exit 1
    fi
    if ! docker info >/dev/null 2>&1; then
        err "Cannot talk to the Docker daemon. Is it running? Do you have permission?"
        exit 1
    fi
    log "Found Docker $(docker --version | awk '{print $3}' | tr -d ',')"
}

install_local() {
    check_local_prereqs
    log "Downloading Go dependencies..."
    go mod download
    log "Installing frontend dependencies and building..."
    (cd web && npm ci && npm run build)
    log "Building Go binary..."
    mkdir -p bin
    CGO_ENABLED=0 go build -o bin/composarr ./cmd/composarr
    log "Built ./bin/composarr"

    confirm_start
    if [ "$START_AFTER_INSTALL" = "yes" ]; then
        log "Starting Composarr on port $PORT..."
        COMPOSARR_PORT="$PORT" ./bin/composarr
    else
        cat <<EOF

Installation complete.
Run Composarr with:

  COMPOSARR_PORT=$PORT ./bin/composarr

Then open http://localhost:$PORT
EOF
    fi
}

install_docker() {
    check_docker_prereqs
    log "Building Docker image (composarr:latest)..."
    docker build -t composarr:latest .

    confirm_start
    if [ "$START_AFTER_INSTALL" = "yes" ]; then
        log "Starting Composarr via docker compose on port $PORT..."
        COMPOSARR_HOST_PORT="$PORT" docker compose up -d
        log "Composarr is up. Open http://localhost:$PORT"
    else
        cat <<EOF

Installation complete.
Start Composarr with:

  docker compose up -d

Then open http://localhost:8080 (or edit docker-compose.yml to change the port).
EOF
    fi
}

log "Composarr installer"

if [ -z "$MODE" ]; then
    if [ -t 0 ] && [ -t 1 ]; then
        prompt_mode
    else
        err "No --mode specified and not running in an interactive terminal."
        usage
        exit 2
    fi
fi

case "$MODE" in
    local)  install_local ;;
    docker) install_docker ;;
    *) err "Invalid mode: $MODE"; usage; exit 2 ;;
esac
