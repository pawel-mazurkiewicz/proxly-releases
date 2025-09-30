# Proxly License Server

A secure, scalable Go API for managing license keys and activation counts. Built for the Proxly browser chooser application.

## Features

- 🔐 **HMAC-signed authentication** for client requests
- 🛡️ **JWT authentication** for admin operations
- 📊 **Activation tracking** with threshold notifications
- 🪝 **Webhook notifications** for events
- 🚦 **Rate limiting** per IP and per license
- 📈 **Prometheus metrics** for monitoring
- 🐳 **Docker-ready** with docker-compose setup
- 🔍 **Structured logging** with zap
- 🧪 **Comprehensive tests** with mocks
- 📝 **OpenAPI specification** for API documentation

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and setup**:
   ```bash
   cd proxly-license-server
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start the services**:
   ```bash
   make docker-up
   # or
   docker-compose -f docker/docker-compose.yml up -d
   ```

3. **Run migrations and seed data**:
   ```bash
   # Wait for database to be ready, then:
   ./scripts/run_migrations.sh
   ./scripts/dev_seed.sh
   ```

4. **The API is now available at**: `http://localhost:8080`

### Local Development

1. **Prerequisites**:
   - Go 1.22+
   - PostgreSQL 16+
   - [golang-migrate](https://github.com/golang-migrate/migrate)

2. **Setup database**:
   ```bash
   createdb proxly
   make migrate-up
   ```

3. **Run the server**:
   ```bash
   make run
   # or
   go run cmd/api/main.go
   ```

## API Usage

### Authentication

The API uses two authentication methods:

#### HMAC Authentication (Client Requests)
For activation requests, use HMAC-SHA256 signatures:

```bash
timestamp=$(date +%s)
client_id="test-client"
body='{"license_key": "DEMO-001", "client_id": "test-client"}'
signature=$(echo -n "${timestamp}${client_id}${body}" | openssl dgst -sha256 -hmac "hmac-secret-key" | cut -d' ' -f2)

curl -X POST http://localhost:8080/v1/activate \
  -H "Content-Type: application/json" \
  -H "X-Timestamp: $timestamp" \
  -H "X-Client-ID: $client_id" \
  -H "X-Signature: $signature" \
  -d "$body"
```

#### JWT Authentication (Admin Operations)

Admin routes require a JWT signed with `JWT_ADMIN_SECRET`. Generate one locally with the helper in `tools/jwtgen`:

```bash
TOKEN=$(tools/jwtgen/jwtgen --secret "$JWT_ADMIN_SECRET" --sub admin --hours 12)

curl -X GET http://localhost:8080/v1/admin/licenses \
  -H "Authorization: Bearer $TOKEN"
```

### Core Endpoints

#### License Activation
```bash
POST /v1/activate
```
Process a license activation with HMAC authentication.

**Example Response**:
```json
{
  "license_id": "0b0c4f3e-4b4c-4c4c-4c4c-4c4c4c4c4c4c",
  "activation_count": 3,
  "max_activations": 5,
  "remaining_activations": 2,
  "status": "ok"
}
```

> Re-activating the same license from the same device fingerprint is idempotent—the activation count is not incremented and the existing record is reused.

#### Trials
```bash
POST /v1/trial/start
GET  /v1/trial/status
```
Start a 7-day (configurable) trial or query its status. Both endpoints require the same HMAC headers as `/v1/activate` and expect `client_id` plus `device_fingerprint`.

#### License Management (Admin)
```bash
GET    /v1/admin/licenses           # List licenses
POST   /v1/admin/licenses           # Create license
GET    /v1/admin/licenses/{id}      # Get license
PUT    /v1/admin/licenses/{id}      # Update license
DELETE /v1/admin/licenses/{id}      # Delete license
GET    /v1/admin/licenses/{id}/activations        # List activations for a license
POST   /v1/admin/licenses/{id}/reset-activations  # Reset activation count and purge records
GET    /v1/admin/trials               # List trials
POST   /v1/admin/trials/{id}/status   # Update trial status (cancel/convert/expire)
```

#### Webhook Management (Admin)
```bash
GET  /v1/admin/webhooks             # List webhooks
POST /v1/admin/webhooks             # Create webhook
```

#### System Endpoints
```bash
GET /healthz                        # Health check
GET /metrics                        # Prometheus metrics
```

## Configuration

Configuration is done via environment variables. See `.env.example` for all options.

**Key Settings**:
- `APP_PORT`: Server port (default: 8080)
- `DB_DSN`: PostgreSQL connection string
- `JWT_ADMIN_SECRET`: Secret for JWT token signing
- `HMAC_SECRET`: Shared secret for public API request signing
- `DEFAULT_MAX_ACTIVATIONS`: Default activation limit (default: 5)
- `AUTO_CREATE_ON_ACTIVATION`: Auto-create licenses on first activation
- `RATE_LIMIT_PER_IP_PER_MIN`: IP-based rate limit (default: 30)
- `TRIAL_LENGTH_DAYS`: Free-trial duration (default: 7)
- `TRIAL_RETENTION_DAYS`: Days to keep historical trials before cleanup (default: 365)

## Webhooks

The server sends webhook notifications for key events:

### Event Types
- `threshold.crossed`: License reaches maximum activations
- `activation.exceeded`: License activation attempts exceed limit
- `license.revoked`: License is revoked by admin

### Webhook Payload
```json
{
  "event": "threshold.crossed",
  "timestamp": "2025-01-20T10:30:00Z",
  "data": {
    "license_id": "uuid",
    "key": "LICENSE-KEY",
    "activation_count": 5,
    "max_activations": 5
  }
}
```

### Webhook Security
Webhooks are signed with HMAC-SHA256. Verify the `X-Proxly-Signature` header:

```go
expectedSig := "sha256=" + hmac256(payload, webhookSecret)
if hmac.Equal([]byte(receivedSig), []byte(expectedSig)) {
    // Valid webhook
}
```

## Architecture

```
cmd/api/                    # Application entry point
internal/
  ├── core/                # Business logic
  │   ├── license.go      # License service
  │   ├── auth.go         # Authentication
  │   └── ratelimit.go    # Rate limiting
  ├── http/               # HTTP layer
  │   ├── handlers/       # HTTP handlers
  │   ├── middleware/     # HTTP middleware
  │   └── router/         # Route configuration
  ├── store/              # Data persistence
  │   ├── db.go          # Database connection
  │   ├── licenses_repo.go # Repository implementation
  │   └── migrations/     # Database migrations
  ├── notify/             # Notification system
  │   ├── webhook.go      # Webhook notifier
  │   └── email_stub.go   # Email notifier (stub)
  └── config/             # Configuration
pkg/
  ├── types/              # Shared types
  └── errorsx/            # Error definitions
```

## Development

### Running Tests
```bash
make test
# or
go test -v ./...
```

### Database Migrations
```bash
# Create new migration
migrate create -ext sql -dir internal/store/migrations -seq migration_name

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Building
```bash
make build              # Build binary
make docker-build       # Build Docker image
```

## Monitoring

### Health Checks
The `/healthz` endpoint provides health status including database connectivity.

### Metrics
Prometheus metrics are available at `/metrics` including:
- HTTP request counts and durations
- Database connection pool stats
- Rate limiting metrics
- License activation metrics

### Logging
Structured JSON logging with configurable levels:
- Request/response logging
- Security events (auth failures, rate limits)
- Business events (activations, threshold crossings)
- Error tracking

## Security

### HMAC Request Signing
All client activation requests must be signed with HMAC-SHA256:
1. Concatenate: `timestamp + clientId + requestBody`
2. Sign with shared secret
3. Include signature in `X-Signature` header

### Rate Limiting
- **Per-IP limiting**: Prevents abuse from single sources
- **Per-License limiting**: Prevents license key farming
- **Configurable limits**: Adjust based on usage patterns

### Timestamp Validation
- Request timestamps must be within ±5 minutes
- Prevents replay attacks
- Configurable time window

### Admin Security
- JWT tokens for admin operations
- Separate authentication from client requests
- Token expiration (24 hours default)

## Swift Client Integration

For integration with the Proxly macOS app:

```swift
class LicenseManager {
    private let apiBase = "https://license.proxly.app/v1"
    private let clientId = // Store in Keychain
    private let hmacSecret = // Obfuscate in binary

    func activate(licenseKey: String) async throws -> ActivationResponse {
        let timestamp = String(Date().timeIntervalSince1970)
        let body = ActivationRequest(licenseKey: licenseKey, clientId: clientId)
        let bodyData = try JSONEncoder().encode(body)

        let signature = hmacSHA256(
            message: timestamp + clientId + String(data: bodyData, encoding: .utf8)!,
            secret: hmacSecret
        )

        var request = URLRequest(url: URL(string: "\(apiBase)/activate")!)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(timestamp, forHTTPHeaderField: "X-Timestamp")
        request.setValue(clientId, forHTTPHeaderField: "X-Client-ID")
        request.setValue(signature, forHTTPHeaderField: "X-Signature")
        request.httpBody = bodyData

        // Execute request and handle response
    }
}
```

## Deployment

### Production Checklist
- [ ] Change default JWT secret
- [ ] Configure HMAC secret for clients
- [ ] Set up database backups
- [ ] Configure webhook endpoints
- [ ] Set up monitoring and alerting
- [ ] Enable TLS/HTTPS
- [ ] Configure rate limits for production load
- [ ] Set up log aggregation

### Environment Variables (Production)
```bash
APP_PORT=8080
DB_DSN=postgres://user:pass@host:5432/proxly?sslmode=require
JWT_ADMIN_SECRET=long-random-secret-here
WEBHOOK_TIMEOUT_MS=5000
RATE_LIMIT_PER_IP_PER_MIN=100
RATE_LIMIT_PER_LICENSE_PER_MIN=20
LOG_LEVEL=info
```

## API Documentation

Full API documentation is available in [OpenAPI format](api/openapi.yaml).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is part of the Proxly browser chooser application.

---

For questions or support, please check the GitHub issues or create a new issue.

## Admin UI

A lightweight web console lives in `../tools/license-admin-web`. Open `index.html` in a modern browser, supply your admin JWT and server URL, and you can list, edit, and reset licenses without touching curl. See the tool README for security notes and usage tips.
