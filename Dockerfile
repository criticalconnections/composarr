# Stage 1: Build React frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy built frontend into the static directory for embedding
COPY --from=frontend /app/web/dist ./internal/handler/static
RUN CGO_ENABLED=0 GOOS=linux go build -o /composarr ./cmd/composarr

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache \
    docker-cli \
    docker-cli-compose \
    ca-certificates \
    tzdata

COPY --from=backend /composarr /usr/local/bin/composarr

EXPOSE 8080
VOLUME /data

ENV COMPOSARR_DATA_DIR=/data
ENV COMPOSARR_PORT=8080

ENTRYPOINT ["composarr"]
