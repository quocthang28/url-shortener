# Run Instructions

This file contains the detailed steps to run, test, and optionally containerize
the URL shortener assignment.

## Requirements

- Go 1.26
- `make`
- `curl` for the smoke-test commands

No external database is required. The service stores mappings in SQLite using
the pure-Go `modernc.org/sqlite` driver.

Docker is optional.

## Run Locally

Start the server:

```bash
make run
```

By default, the server listens on `http://localhost:8080` and stores SQLite data
in `./urls.db`.

Run with explicit settings:

```bash
PORT=8080 DB_PATH=./urls.db BASE_URL=http://localhost:8080 make run
```

## Smoke Test

Encode a URL:

```bash
curl -s -X POST http://localhost:8080/encode \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://codesubmit.io/library/react"}'
```

Example response:

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

Decode the short URL:

```bash
curl -s -X POST http://localhost:8080/decode \
  -H 'Content-Type: application/json' \
  -d '{"short_url":"http://localhost:8080/GeAi9K"}'
```

Follow the redirect:

```bash
curl -i http://localhost:8080/GeAi9K
```

Replace `GeAi9K` with the code returned by your local `/encode` response.

## Build a Local Binary

```bash
make build
./bin/url-shortener
```

The binary uses the same environment variables as `make run`.

## Run Tests

Run the full test suite:

```bash
make test
```

Useful development checks:

```bash
make fmt
make vet
make tidy
```

## Docker

Build the Docker image:

```bash
make docker
```

Run it with a persistent SQLite volume:

```bash
docker run --rm -p 8080:8080 \
  -e BASE_URL=http://localhost:8080 \
  -v url-shortener-data:/data \
  url-shortener
```

The container stores SQLite data at `/data/urls.db`.

## Docker Compose With Caddy

The included Compose file is for deployment behind Caddy at
`url-shortener.thang-dev.xyz`. Before using it for another domain, update both:

- `BASE_URL` in `docker-compose.yml`
- the site name in `Caddyfile`

`docker-compose.yml` uses `Dockerfile.run`, which copies a prebuilt Linux binary
named `url-shortener`. Build that binary first. Use `GOARCH=amd64` for a
typical x86_64 Linux server, or `GOARCH=arm64` for an ARM64 Linux host:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o url-shortener ./cmd/server
```

Then start the deployment:

```bash
docker compose up --build -d
```

Check logs:

```bash
docker compose logs -f
```

Stop it:

```bash
docker compose down
```

The Compose setup uses a named Docker volume called `data` to persist
`/data/urls.db` across restarts.

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./urls.db` locally, `/data/urls.db` in Docker | SQLite database file path |
| `BASE_URL` | `http://localhost:8080` | Prefix used in `/encode` responses |
| `GIN_MODE` | unset locally, `release` in Docker | Gin runtime mode |

Set `BASE_URL` to the public domain in deployment. If it is left as
`localhost`, generated short URLs will point at localhost.
