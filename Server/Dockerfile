# syntax=docker/dockerfile:1

# ---------- Build stage ----------
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache dependencies first for faster rebuilds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# modernc.org/sqlite is pure Go, so no cgo/musl toolchain is required.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app endpoints/monolith.go

# ---------- Runtime stage ----------
FROM alpine:3.20
WORKDIR /

# ca-certificates: outbound TLS (S3, SMTP). wget comes from busybox and is used
# by the Coolify healthcheck. tzdata: correct local time in logs and scheduling.
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/app /app

EXPOSE 8080

# Persistent volume for core.db and per-user .db files. On Coolify this is the
# host-mounted /storage volume.
ENV STORAGE_PATH=/storage
VOLUME /storage

# MySQL is only the read-only source during the SQLite drain. Once every user is
# migrated, these may be left empty and the drain/bootstrap paths are skipped.
ENV DATABASE_HOST=
ENV DATABASE_NAME=
ENV DATABASE_USER=
ENV DATABASE_PASSWORD=
ENV ALLOWED_ORIGIN=
ENV GO_API_HOST=
ENV REGION=
ENV BUCKETNAME=

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:8080/health || exit 1

CMD ["/app"]
