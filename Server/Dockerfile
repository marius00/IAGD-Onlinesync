# Start by building the application.
FROM golang:1.25-alpine as build

WORKDIR /go/src/app

# Cache dependencies first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go, so no cgo/musl toolchain is required.
RUN env GOOS=linux go build -ldflags="-s -w" -o /go/bin/app endpoints/monolith.go

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian12
COPY --from=build /go/bin/app /
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
CMD ["/app"]
