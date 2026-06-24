# URL Shortener

A small Go URL shortening service. It accepts a long `http` or `https` URL,
stores it in SQLite, returns a short URL, and can later decode or redirect that
short URL back to the original URL.

The service is intentionally simple and single-node, but it includes the core
production concerns for the assignment: persistent storage, URL validation,
idempotent encodes, collision handling, graceful shutdown, tests, and per-IP
rate limiting.

## Run Instructions

Detailed setup, local run, Docker, Docker Compose, and test instructions are in
[RUN.md](RUN.md).

## API

All JSON errors use this shape:

```json
{ "error": "descriptive message" }
```

Common status codes:

| Status | Meaning |
|--------|---------|
| `200` | Successful JSON response |
| `302` | Redirect to original URL |
| `400` | Bad request or invalid URL |
| `404` | Unknown short code |
| `429` | Rate limit exceeded |
| `500` | Internal server error |
| `503` | Could not allocate a short code after retries |

### POST /encode

Encodes a long URL into a short URL.

Request:

```json
{ "url": "https://codesubmit.io/library/react" }
```

Response:

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

Example:

```bash
curl -s -X POST http://localhost:8080/encode \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://codesubmit.io/library/react"}'
```

Encoding the same URL more than once returns the same short URL.

### POST /decode

Decodes a short URL into the original URL.

Request:

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

Response:

```json
{ "url": "https://codesubmit.io/library/react" }
```

Example:

```bash
curl -s -X POST http://localhost:8080/decode \
  -H 'Content-Type: application/json' \
  -d '{"short_url":"http://localhost:8080/GeAi9K"}'
```

### GET /{shortCode}

Redirects to the original URL with `302 Found`.

Example:

```bash
curl -i http://localhost:8080/GeAi9K
```

## Persistence

Mappings are stored in SQLite. The database schema is applied automatically at
startup:

```sql
CREATE TABLE IF NOT EXISTS urls (
    short_code   TEXT PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE
);
```

`short_code` is indexed by the primary key for decode and redirect lookups.
`original_url` is unique so `/encode` is idempotent: the same long URL maps to
the same short code.

The SQLite file persists across restarts. Locally it defaults to `./urls.db`;
in the Docker image it defaults to `/data/urls.db`.

## Security

### URL validation

Submitted URLs are validated before being stored. The input is trimmed of
surrounding whitespace, then the following rules are enforced:

- **Length cap of 2048 bytes.** Oversized input is rejected before parsing.
- **Must parse as a URL.** Anything `net/url` cannot parse is rejected.
- **`http` or `https` scheme only.** This rejects targets like `javascript:`,
  `data:`, and `file:` URLs.
- **Non-empty host.** Relative or malformed URLs are rejected.
- **No embedded user credentials.** URLs like `http://user:pass@host` are
  rejected to avoid phishing tricks and credential leakage.
- **Non-empty hostname with no whitespace.** This defends against malformed
  authorities and header or redirect injection via control characters.

### Payload size limits

`/encode` and `/decode` cap request bodies with `http.MaxBytesReader` before
JSON binding. This prevents large bodies from consuming excessive memory.

### Rate limiting

Requests are rate-limited per client IP using an in-memory token bucket from
`golang.org/x/time/rate`.

The write path is stricter than the read paths:

| Route | Limit |
|-------|-------|
| `POST /encode` | 1 request/sec, burst 3 |
| `POST /decode` | 5 requests/sec, burst 15 |
| `GET /{shortCode}` | 5 requests/sec, burst 15 |

The in-memory limiter is sufficient for a single-node deployment. In a
multi-node deployment, rate limit state should move to a shared store such as
Redis so limits apply consistently across instances.

The app trusts the Docker bridge network (`172.16.0.0/12`) as a proxy, so when
deployed behind the bundled Caddy reverse proxy the limiter keys on the real
client IP via `X-Forwarded-For`. Requests arriving from outside that range have
their forwarded headers ignored and are keyed on the direct connection address,
so a client cannot spoof its IP to dodge the limit.

### Short code generation

Short codes are generated with `crypto/rand` from a 62-character alphabet:
`a-z`, `A-Z`, and `0-9`. The default length is 6 characters, giving about
56.8 billion possible codes.

The service does not use `math/rand`, because predictable random output would
make code guessing easier.

### Collision handling

Short-code collisions are handled by the database. If an insert fails because
the generated code already exists, `/encode` generates a new code and retries.
If another request stored the same original URL concurrently, the handler reads
and returns the winning code.

For a 6-character Base62 code, the code space is:

```text
62^6 = 56,800,235,584 possible codes
```

The relevant collision risk for each new insert is the chance that one freshly
generated code is already present:

```text
P(collision) = stored_urls / 56,800,235,584
```

Examples:

| Stored URLs | Collision chance for next insert |
|-------------|----------------------------------|
| 1,000 | 0.0000018% |
| 1,000,000 | 0.0018% |
| 10,000,000 | 0.018% |
| 100,000,000 | 0.18% |

Even at 100 million stored URLs, a single generated code has about a 1 in 568
chance of colliding. The retry loop makes repeated failure much rarer: two
collisions in a row at that size are about `0.18% * 0.18% = 0.0003%`, or roughly
1 in 315,000 attempts.

### SSRF

The service does not make outbound HTTP requests to submitted URLs. It stores
and redirects only. Because there is no link preview, reachability check, or
metadata fetch, the usual server-side request forgery risk is avoided.

### Open redirect abuse

Any URL shortener can be abused to hide a malicious destination behind a trusted
short domain. This implementation mitigates the highest-risk schemes by only
accepting `http` and `https` URLs. At production scale, submitted URLs should
also be checked against a malware or phishing feed such as Google Safe Browsing.

### SQL injection

All SQLite queries use parameterized statements through `database/sql`.

## Scalability

This implementation is intentionally single-node for the assignment.

### Current bottlenecks

- **SQLite write serialization.** SQLite is sufficient for demo persistence,
  but it serializes writes. High `/encode` traffic would eventually require a
  server database such as PostgreSQL.
- **In-memory rate limiting.** Each process has its own limiter state. With
  multiple application nodes, limits must move to a shared store.
- **No read cache.** Every decode and redirect lookup currently goes to SQLite.

### Future improvements

- Add Redis as a read-through cache for decode and redirect lookups. Short-code
  mappings are immutable, so cache invalidation is simple.
- Move the database from SQLite to PostgreSQL for better concurrent writes and
  operational tooling.
- Use Redis-backed rate limiting when running more than one application
  instance.
- Increase short code length from 6 to 7 characters if stored URL count grows
  large enough that collisions become common.
- At very large scale, shard by short code so decode and redirect requests route
  deterministically to the right storage partition.

## AI Usage

AI assistance was used to speed up the mechanical parts of this project:
generating boilerplate, scaffolding the package layout, and setting up tooling
(Makefile, Dockerfile, test scaffolds).

The substance is my own. The architecture, the logic of each code path, the
error and concurrency handling, and the security and scalability trade-offs
documented above were all reasoned through and decided by me.
