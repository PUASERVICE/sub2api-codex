# syntax=docker/dockerfile:1.7

# =============================================================================
# Sub2API Multi-Stage Dockerfile
# =============================================================================
# Stage 1: Build frontend
# Stage 2: Build Go backend with embedded frontend
# Stage 3: Final runtime image
# =============================================================================

ARG NODE_IMAGE=node:24-alpine
ARG GOLANG_IMAGE=golang:1.26.2-alpine
ARG ALPINE_IMAGE=alpine:3.20
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

# -----------------------------------------------------------------------------
# Stage 1: Frontend Builder
# -----------------------------------------------------------------------------
FROM ${NODE_IMAGE} AS frontend-builder

WORKDIR /app/frontend

# Install pnpm
RUN corepack enable && corepack prepare pnpm@9.15.4 --activate

# Install dependencies first (better caching)
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --prefer-offline

# Copy frontend source and build
COPY frontend/ ./
RUN pnpm run build

# -----------------------------------------------------------------------------
# Stage 2: Backend Builder
# -----------------------------------------------------------------------------
FROM ${GOLANG_IMAGE} AS backend-builder

# Build arguments for version info (set by CI)
ARG VERSION=
ARG COMMIT=docker
ARG DATE
ARG GOPROXY
ARG GOSUMDB

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

# Install build dependencies
RUN apk add --update-cache --no-progress git ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

WORKDIR /app/backend

# Copy go mod files first (better caching)
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy backend source first
COPY backend/ ./

# Copy frontend dist from previous stage (must be after backend copy to avoid being overwritten)
COPY --from=frontend-builder /app/backend/internal/web/dist ./internal/web/dist

# Build the binary (BuildType=release for CI builds, embed frontend)
# Version precedence: build arg VERSION > cmd/server/VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    VERSION_VALUE="${VERSION}" && \
    if [ -z "${VERSION_VALUE}" ]; then VERSION_VALUE="$(tr -d '\r\n' < ./cmd/server/VERSION)"; fi && \
    DATE_VALUE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -tags embed \
    -ldflags="-s -w -X main.Version=${VERSION_VALUE} -X main.Commit=${COMMIT} -X main.Date=${DATE_VALUE} -X main.BuildType=release" \
    -trimpath \
    -o /app/sub2api \
    ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3: Final Runtime Image
# -----------------------------------------------------------------------------
FROM ${NODE_IMAGE}

# Labels
LABEL maintainer="Wei-Shaw <github.com/Wei-Shaw>"
LABEL description="Sub2API - AI API Gateway Platform"
LABEL org.opencontainers.image.source="https://github.com/Wei-Shaw/sub2api"

# Install runtime dependencies
RUN apk add --update-cache --no-progress \
    ca-certificates \
    tzdata \
    curl \
    su-exec \
    postgresql-client \
    && rm -rf /var/cache/apk/*

# Create non-root user (avoid fixed uid/gid conflicts in base image)
RUN addgroup -S sub2api && \
    adduser -S -G sub2api -s /bin/sh sub2api

# Set working directory
WORKDIR /app

# Copy binary/resources with ownership to avoid extra full-layer chown copy
COPY --from=backend-builder --chown=sub2api:sub2api /app/sub2api /app/sub2api
COPY --from=backend-builder --chown=sub2api:sub2api /app/backend/resources /app/resources

# Create data directory
RUN mkdir -p /app/data && chown sub2api:sub2api /app/data

# Copy entrypoint script (fixes volume permissions then drops to sub2api)
COPY deploy/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose port (can be overridden by SERVER_PORT env var)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q -T 5 -O /dev/null http://localhost:${SERVER_PORT:-8080}/health || exit 1

# Run the application (entrypoint fixes /app/data ownership then execs as sub2api)
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/sub2api"]
