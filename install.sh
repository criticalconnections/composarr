#!/usr/bin/env bash
set -Eeuo pipefail

# Composarr installer — supports a local (build-from-source) install or a
# Docker-based install. Run non-interactively with --mode=local|docker or
# interactively (the script will prompt). Shows live progress for all
# downloads and surfaces detailed errors when any step fails.

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_DIR"

MODE=""
PORT="8080"
START_AFTER_INSTALL="ask"
AUTO_INSTALL_PREREQS="ask"         # ask|yes|no
LOG_FILE="${TMPDIR:-/tmp}/composarr-install-$$.log"
CURRENT_STEP="startup"

# Pinned prereq versions used by the auto-installer.
GO_VERSION="${GO_VERSION:-1.23.4}"
NODE_VERSION="${NODE_VERSION:-22.11.0}"
VENDOR_DIR="$REPO_DIR/.vendor"

# ---- formatting helpers ---------------------------------------------------

if [ -t 1 ]; then
    C_RESET=$'\033[0m'; C_BLUE=$'\033[1;34m'; C_YELLOW=$'\033[1;33m'
    C_RED=$'\033[1;31m'; C_GREEN=$'\033[1;32m'; C_DIM=$'\033[2m'
else
    C_RESET=""; C_BLUE=""; C_YELLOW=""; C_RED=""; C_GREEN=""; C_DIM=""
fi

log()     { printf '%s==>%s %s\n' "$C_BLUE"   "$C_RESET" "$*"; }
ok()      { printf '%s ok%s  %s\n' "$C_GREEN"  "$C_RESET" "$*"; }
warn()    { printf '%s !! %s  %s\n' "$C_YELLOW" "$C_RESET" "$*" >&2; }
err()     { printf '%s XX %s  %s\n' "$C_RED"    "$C_RESET" "$*" >&2; }
step()    { CURRENT_STEP="$1"; log "$1"; }

have() { command -v "$1" >/dev/null 2>&1; }

# ---- error reporting ------------------------------------------------------

on_error() {
    local exit_code=$1 line=$2 cmd=${3:-?}
    err "Install failed during: ${CURRENT_STEP}"
    err "Command exited ${exit_code} at install.sh:${line}"
    err "  -> ${cmd}"
    if [ -s "$LOG_FILE" ]; then
        err "Last 20 lines of log (${LOG_FILE}):"
        tail -n 20 "$LOG_FILE" | sed "s/^/${C_DIM}   | ${C_RESET}/" >&2 || true
        err "Full log: ${LOG_FILE}"
    fi
    case "$CURRENT_STEP" in
        *prereq*|*download*Go*|*download*Node*)
            err "Hint: check your network, proxy, or try installing the prereq manually." ;;
        *npm*)
            err "Hint: try 'rm -rf web/node_modules' and re-run, or check Node.js version." ;;
        *go*build*|*go*module*)
            err "Hint: try 'go clean -cache' and re-run, or verify Go >= 1.23." ;;
        *docker*build*)
            err "Hint: ensure the Docker daemon has disk space and network access." ;;
    esac
    exit "$exit_code"
}
trap 'on_error $? $LINENO "$BASH_COMMAND"' ERR
trap 'warn "Installer interrupted."; exit 130' INT TERM

# ---- arg parsing ----------------------------------------------------------

usage() {
    cat <<'EOF'
Usage: ./install.sh [options]

Options:
  --mode=<local|docker>   Install mode (skip the interactive prompt)
  --port=<port>           Host port for Composarr (default: 8080)
  --start                 Start Composarr after installing
  --no-start              Do not start Composarr after installing
  --install-prereqs       Auto-download missing Go/Node into .vendor/ (local mode)
  --no-install-prereqs    Never download prereqs; fail if missing
  -h, --help              Show this help

Modes:
  local   Build the Go binary + frontend on this machine. Requires Go 1.23+
          and Node.js 22+ (installer can download these into .vendor/ on
          Linux/macOS if you like).
  docker  Build a Docker image and run via docker compose. Requires Docker
          and the Docker Compose plugin.

Progress is shown for every download. On any failure the script prints the
failing step, the exact command, and the tail of the install log.
EOF
}

for arg in "$@"; do
    case "$arg" in
        --mode=*) MODE="${arg#*=}" ;;
        --port=*) PORT="${arg#*=}" ;;
        --start) START_AFTER_INSTALL="yes" ;;
        --no-start) START_AFTER_INSTALL="no" ;;
        --install-prereqs) AUTO_INSTALL_PREREQS="yes" ;;
        --no-install-prereqs) AUTO_INSTALL_PREREQS="no" ;;
        -h|--help) usage; exit 0 ;;
        *) err "Unknown option: $arg"; usage; exit 2 ;;
    esac
done

# ---- prompts --------------------------------------------------------------

is_interactive() { [ -t 0 ] && [ -t 1 ]; }

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
    if ! is_interactive; then START_AFTER_INSTALL="no"; return; fi
    read -r -p "Start Composarr now? [Y/n]: " ans
    case "${ans:-Y}" in
        y|Y|yes|Yes|YES) START_AFTER_INSTALL="yes" ;;
        *)               START_AFTER_INSTALL="no" ;;
    esac
}

confirm_prereq_install() {
    # Name is $1, purpose is $2. Honors --install-prereqs / --no-install-prereqs.
    if [ "$AUTO_INSTALL_PREREQS" = "yes" ]; then return 0; fi
    if [ "$AUTO_INSTALL_PREREQS" = "no" ];  then return 1; fi
    if ! is_interactive; then return 1; fi
    read -r -p "Download $1 into .vendor/ for this install? [Y/n]: " ans
    case "${ans:-Y}" in
        y|Y|yes|Yes|YES) return 0 ;;
        *)               return 1 ;;
    esac
}

# ---- platform detection ---------------------------------------------------

detect_os()   { case "$(uname -s)" in Linux) echo linux;; Darwin) echo darwin;; *) echo unsupported;; esac; }
detect_arch() { case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; *) echo unsupported;; esac; }

# ---- downloader with live progress ----------------------------------------

download() {
    # download <url> <dest>. Shows a live progress bar; writes stderr to LOG_FILE.
    local url="$1" dest="$2"
    log "Downloading $(basename "$dest")"
    printf '%s    %s%s\n' "$C_DIM" "$url" "$C_RESET"
    if have curl; then
        # -# gives a progress bar; --fail surfaces HTTP errors; -L follows redirects.
        curl -fL --retry 3 --retry-delay 2 -# -o "$dest" "$url" 2>>"$LOG_FILE"
    elif have wget; then
        wget --tries=3 --show-progress -O "$dest" "$url" 2>>"$LOG_FILE"
    else
        err "Neither curl nor wget is installed; cannot download $url"
        return 1
    fi
    ok "Downloaded $(basename "$dest") ($(du -h "$dest" | awk '{print $1}'))"
}

# ---- prereq auto-installers ----------------------------------------------

install_go_into_vendor() {
    CURRENT_STEP="download Go $GO_VERSION"
    local os arch tarball url dest
    os="$(detect_os)"; arch="$(detect_arch)"
    if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
        err "Auto-install of Go not supported on $(uname -s)/$(uname -m). Install Go 1.23+ manually: https://go.dev/dl/"
        return 1
    fi
    tarball="go${GO_VERSION}.${os}-${arch}.tar.gz"
    url="https://go.dev/dl/${tarball}"
    mkdir -p "$VENDOR_DIR"
    dest="$VENDOR_DIR/$tarball"
    download "$url" "$dest"
    step "Extracting Go into $VENDOR_DIR/go"
    rm -rf "$VENDOR_DIR/go"
    tar -C "$VENDOR_DIR" -xzf "$dest"
    rm -f "$dest"
    export PATH="$VENDOR_DIR/go/bin:$PATH"
    export GOROOT="$VENDOR_DIR/go"
    ok "Go installed: $(go version)"
}

install_node_into_vendor() {
    CURRENT_STEP="download Node.js $NODE_VERSION"
    local os arch tarball url dest dirname
    os="$(detect_os)"; arch="$(detect_arch)"
    if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
        err "Auto-install of Node.js not supported on $(uname -s)/$(uname -m). Install Node.js 22+ manually: https://nodejs.org/"
        return 1
    fi
    dirname="node-v${NODE_VERSION}-${os}-${arch}"
    tarball="${dirname}.tar.xz"
    url="https://nodejs.org/dist/v${NODE_VERSION}/${tarball}"
    mkdir -p "$VENDOR_DIR"
    dest="$VENDOR_DIR/$tarball"
    download "$url" "$dest"
    if ! have tar || ! tar --help 2>&1 | grep -q -- '--xz\|xz'; then
        if ! have xz; then
            err "xz (or a tar with .xz support) is required to extract Node.js. Install 'xz-utils'."
            return 1
        fi
    fi
    step "Extracting Node.js into $VENDOR_DIR/node"
    rm -rf "$VENDOR_DIR/node"
    tar -C "$VENDOR_DIR" -xJf "$dest"
    mv "$VENDOR_DIR/$dirname" "$VENDOR_DIR/node"
    rm -f "$dest"
    export PATH="$VENDOR_DIR/node/bin:$PATH"
    ok "Node.js installed: $(node --version), npm $(npm --version)"
}

# ---- prereq checks --------------------------------------------------------

check_local_prereqs() {
    step "Checking local build prerequisites"
    # Prefer any tool already on PATH from a previous run.
    [ -x "$VENDOR_DIR/go/bin/go" ]      && export PATH="$VENDOR_DIR/go/bin:$PATH"
    [ -x "$VENDOR_DIR/node/bin/node" ]  && export PATH="$VENDOR_DIR/node/bin:$PATH"

    local need_go=0 need_node=0
    if have go; then
        ok "Found Go $(go version | awk '{print $3}' | sed 's/^go//')"
    else
        warn "Go is not installed (need 1.23+)."
        need_go=1
    fi
    if have node && have npm; then
        ok "Found Node.js $(node --version), npm $(npm --version)"
    else
        warn "Node.js / npm not installed (need Node 22+)."
        need_node=1
    fi

    if [ "$need_go" -eq 1 ]; then
        if confirm_prereq_install "Go $GO_VERSION"; then
            install_go_into_vendor
        else
            err "Cannot continue without Go. Install from https://go.dev/dl/ and re-run."
            exit 1
        fi
    fi
    if [ "$need_node" -eq 1 ]; then
        if confirm_prereq_install "Node.js $NODE_VERSION"; then
            install_node_into_vendor
        else
            err "Cannot continue without Node.js / npm. Install from https://nodejs.org/ and re-run."
            exit 1
        fi
    fi
}

check_docker_prereqs() {
    step "Checking Docker prerequisites"
    if ! have docker; then
        err "Docker is not installed. See https://docs.docker.com/engine/install/"
        exit 1
    fi
    if ! docker compose version >/dev/null 2>&1; then
        err "Docker Compose plugin not found. Install the 'docker compose' plugin:"
        err "  https://docs.docker.com/compose/install/"
        exit 1
    fi
    if ! docker info >/dev/null 2>&1; then
        err "Cannot reach the Docker daemon. Is it running? Are you in the 'docker' group?"
        err "  Try: sudo systemctl start docker   (Linux)"
        err "  Or:  open -a Docker                (macOS)"
        exit 1
    fi
    ok "Docker $(docker --version | awk '{print $3}' | tr -d ',') reachable"
    ok "Compose $(docker compose version --short 2>/dev/null || echo ok)"
}

# ---- install paths --------------------------------------------------------

install_local() {
    check_local_prereqs

    step "Downloading Go module dependencies"
    # -x prints per-module download progress to stderr; tee so user sees it and we log it.
    GOFLAGS="${GOFLAGS:-}" go mod download -x 2> >(tee -a "$LOG_FILE" >&2)

    step "Installing frontend dependencies (npm ci)"
    # --progress=true forces the progress bar; --loglevel=http shows each fetch.
    (cd web && npm ci --progress=true --loglevel=http 2> >(tee -a "$LOG_FILE" >&2))

    step "Building frontend (vite)"
    (cd web && npm run build | tee -a "$LOG_FILE")

    step "Building Go binary"
    mkdir -p bin
    CGO_ENABLED=0 go build -v -o bin/composarr ./cmd/composarr 2> >(tee -a "$LOG_FILE" >&2)
    ok "Built $(pwd)/bin/composarr"

    confirm_start
    if [ "$START_AFTER_INSTALL" = "yes" ]; then
        step "Starting Composarr on port $PORT"
        COMPOSARR_PORT="$PORT" exec ./bin/composarr
    else
        cat <<EOF

${C_GREEN}Installation complete.${C_RESET}
Run Composarr with:

  COMPOSARR_PORT=$PORT ./bin/composarr

Then open http://localhost:$PORT
EOF
    fi
}

install_docker() {
    check_docker_prereqs

    step "Building Docker image composarr:latest"
    # BuildKit + progress=plain streams every layer's output in real time.
    DOCKER_BUILDKIT=1 docker build --progress=plain -t composarr:latest . 2> >(tee -a "$LOG_FILE" >&2)
    ok "Image composarr:latest built"

    confirm_start
    if [ "$START_AFTER_INSTALL" = "yes" ]; then
        step "Starting Composarr via docker compose"
        docker compose up -d | tee -a "$LOG_FILE"
        ok "Composarr is up. Open http://localhost:$PORT"
        log "Follow logs with: docker compose logs -f composarr"
    else
        cat <<EOF

${C_GREEN}Installation complete.${C_RESET}
Start Composarr with:

  docker compose up -d

Then open http://localhost:8080 (or edit docker-compose.yml to change the port).
EOF
    fi
}

# ---- main -----------------------------------------------------------------

: > "$LOG_FILE"
log "Composarr installer — logging to $LOG_FILE"

if [ -z "$MODE" ]; then
    if is_interactive; then
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
