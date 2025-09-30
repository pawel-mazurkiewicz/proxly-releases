# Proxly License Server — Project Specification

This document describes the architecture, requirements, and design for the Proxly License Server.  
It should be used as the blueprint for implementation in **Go (Golang)**.

---

## 1) Goals & Non-Goals

**Goals**
- Small, secure Go API to manage license keys and activation counts.
- Simple persistence (PostgreSQL or SQLite) with migrations.
- Authenticated CRUD endpoints for license keys + activation workflow.
- Notifies when an activation threshold is crossed (webhook, email stub, or queue).
- One-command deploy via `docker-compose`.
- First-class client integration for `@Proxly/Services/LicenseManager.swift`.

**Non-Goals**
- Full billing, user accounts, or portal UI.
- Complex RBAC or multi-tenant support (leave hooks).

---

## 2) Tech Choices

- **Language:** Go 1.22+
- **Framework:** net/http + chi OR echo (chi suggested)
- **DB:** PostgreSQL (default), SQLite for dev
- **Migrations:** golang-migrate or goose
- **Auth:** HMAC-signed requests (primary) + optional JWT for server-to-server admin
- **Crypto:** Ed25519 or HMAC-SHA256 with per-key secret
- **Container:** Docker + docker-compose
- **Observability:** zap logs, Prometheus metrics
- **Rate limiting:** in-memory token bucket

---

## 3) Repository Layout

```
proxly-license-server/
  cmd/api/main.go
  internal/
    http/
      router.go
      middleware.go
      handlers/
        licenses.go
        health.go
        activations.go
        admin.go
    core/
      license.go
      activation.go
      notifier.go
      auth.go
      ratelimit.go
    store/
      db.go
      licenses_repo.go
      migrations/
    config/
      config.go
    notify/
      webhook.go
      email_stub.go
  pkg/
    errorsx/
    types/
  api/
    openapi.yaml
  scripts/
    dev_seed.sh
    run_migrations.sh
  docker/
    Dockerfile
    docker-compose.yml
  Makefile
  .env.example
  README.md
```

---

## 4) Data Model & Migrations

**licenses**
- `id` UUID PK
- `key` TEXT UNIQUE NOT NULL
- `max_activations` INT NOT NULL
- `activation_count` INT DEFAULT 0
- `status` TEXT (active|revoked|suspended|expired)
- `metadata` JSONB
- `created_at` TIMESTAMPTZ
- `updated_at` TIMESTAMPTZ

**activations**
- `id` UUID PK
- `license_id` UUID FK
- (WON'T BE USED FOR NOW, BUT IT CAN BE ALREADY IN THE SCHEMA - NOT REQUIRED)`client_id` TEXT
- (WON'T BE USED FOR NOW, BUT IT CAN BE ALREADY IN THE SCHEMA - NOT REQUIRED)`device_fingerprint` TEXT
- `ip` INET
- `user_agent` TEXT
- `occurred_at` TIMESTAMPTZ

**webhooks**
- `id` UUID PK
- `url` TEXT NOT NULL
- `secret` TEXT NULL
- `events` TEXT[] NOT NULL

**audit_logs**
- `id` UUID PK
- `actor` TEXT
- `action` TEXT
- `entity` TEXT
- `payload` JSONB
- `created_at` TIMESTAMPTZ

---

## 5) API Contract

See `api/openapi.yaml` for full specification.  
Key endpoints:
- `/v1/licenses` → CRUD (admin, JWT protected)
- `/v1/licenses/{id}` → Manage license
- `/v1/activate` → Client activation flow
- `/v1/webhooks` → Manage webhooks
- `/healthz` → Liveness check

---

## 6) Core Server Logic

- Normalize incoming license keys.
- Activation:
  - If new → store & set activation_count=1
  - If existing → increment activation_count
  - Return remaining activations
  - Notify if threshold reached or exceeded
- Audit every change

---

## 7) Authentication Strategies

**HMAC (recommended baseline)**  
- Clients sign requests with `HMAC-SHA256(secret, timestamp + clientId + body)`.  
- Server validates timestamp (±5 min).  

**Ed25519 signatures**  
- Client signs with private key, server verifies with public key.  

**JWT (admin only)**  
- Used for CRUD endpoints.  

**mTLS (optional)**  
- Strong but heavy distribution overhead.  

---

## 8) Notifications

- Events: `threshold.crossed`, `license.revoked`, `activation.exceeded`
- Webhooks with HMAC signature
- Email stub for extension
- Structured logs

---

## 9) Rate Limiting

- Per-IP and per-license token bucket
- Reject skewed timestamps
- Optional device fingerprint binding

---

## 10) Configuration

`.env.example`
```
APP_PORT=8080
DB_DRIVER=postgres
DB_DSN=postgres://proxly:proxly@db:5432/proxly?sslmode=disable
JWT_ADMIN_SECRET=change-me
CLIENT_SIG_ALGO=hmac
DEFAULT_MAX_ACTIVATIONS=5
AUTO_CREATE_ON_ACTIVATION=false
WEBHOOK_TIMEOUT_MS=3000
RATE_LIMIT_PER_LICENSE_PER_MIN=10
RATE_LIMIT_PER_IP_PER_MIN=30
LOG_LEVEL=info
```

---

## 11) Docker & Compose

**docker-compose.yml**
```yaml
version: "3.9"
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: proxly
      POSTGRES_PASSWORD: proxly
      POSTGRES_DB: proxly
    ports: ["5432:5432"]
    volumes: [dbdata:/var/lib/postgresql/data]
  api:
    build: ..
    environment:
      DB_DRIVER: postgres
      DB_DSN: postgres://proxly:proxly@db:5432/proxly?sslmode=disable
      JWT_ADMIN_SECRET: change-me
    ports: ["8080:8080"]
    depends_on: [db]
volumes:
  dbdata:
```

---

## 12) Swift Client Integration

- Store a `clientId` (Keychain)
- On activation:
  - Build JSON payload
  - Sign request
  - POST to `/v1/activate`
- Parse response:
  - Remaining activations
  - Status (`ok`, `at_threshold`, `exceeded`, etc.)

---

## 13) Security Notes

- Enforce TLS everywhere
- Strict header validation
- Replay protection via nonce or timestamp window
- Audit logging
- Obfuscate secrets in app binary
- Admin behind JWT + IP allowlist

---

## 14) Observability

- Structured logs (zap)
- `/metrics` for Prometheus
- Audit logs in DB

---

## 15) Testing Strategy

- Unit tests: domain logic
- Integration tests: with Postgres
- E2E tests: with docker-compose
- Fuzz tests: for key parsing

---

## 16) Seed & Admin

- `dev_seed.sh` inserts demo licenses
- Admin endpoints: create, update, revoke licenses, manage webhooks

---

## 17) Sample Responses

**Activation OK**
```json
{
  "license_id": "0b0c…",
  "activation_count": 3,
  "max_activations": 5,
  "remaining_activations": 2,
  "status": "ok"
}
```

**Threshold crossed**
```json
{
  "license_id": "0b0c…",
  "activation_count": 5,
  "max_activations": 5,
  "remaining_activations": 0,
  "status": "at_threshold"
}
```

**Exceeded**
```json
{
  "license_id": "0b0c…",
  "activation_count": 6,
  "max_activations": 5,
  "remaining_activations": -1,
  "status": "exceeded"
}
```

---

## 18) Implementation Guidance for Claude Code

- use context7 for documentation style.  
- use subagents for splitting tasks  
- use websearch and use perplexity during the research phase to validate design choices and implementation details.  

---

## 19) Next Steps

1. Scaffold repo
2. Implement migrations
3. Add HMAC auth middleware
4. Implement `/v1/activate` with notifications
5. Add CRUD endpoints
6. Add metrics, rate limits, tests
7. Provide README with usage examples

---
