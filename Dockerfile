# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache deps first. modernc.org/sqlite is pure Go, so no cgo/gcc needed.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/url-shortener ./cmd/server

# ---- runtime stage ----
FROM alpine:3.22
WORKDIR /app

# Run as an unprivileged user and give it ownership of the data volume.
RUN adduser -D -u 10001 appuser \
    && mkdir -p /data \
    && chown appuser:appuser /data

# Persist the SQLite file across restarts: mount a volume at /data.
ENV PORT=8080 \
    DB_PATH=/data/urls.db \
    BASE_URL=http://localhost:8080 \
    GIN_MODE=release
VOLUME /data
EXPOSE 8080

COPY --from=build /out/url-shortener ./url-shortener

USER appuser
ENTRYPOINT ["/app/url-shortener"]
