.PHONY: build dev test lint docker clean

# Build the Go binary
build: build-frontend
	go build -o bin/composarr ./cmd/composarr

# Run in development mode with hot-reload
dev:
	go run ./cmd/composarr

# Run frontend dev server
dev-frontend:
	cd web && npm run dev

# Build frontend
build-frontend:
	cd web && npm ci && npm run build

# Run tests
test:
	go test ./... -v

# Run linter
lint:
	golangci-lint run ./...

# Build Docker image
docker:
	docker build -t composarr:latest .

# Docker compose up
up:
	docker compose up -d

# Docker compose down
down:
	docker compose down

# Clean build artifacts
clean:
	rm -rf bin/ web/dist/ web/node_modules/
