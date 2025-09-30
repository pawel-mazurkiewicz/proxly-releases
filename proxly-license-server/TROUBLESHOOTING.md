# Troubleshooting

## Common Issues and Solutions

### 1. "invalid input syntax for type json" Error

**Error:** `pq: invalid input syntax for type json`

**Cause:** This error occurs when trying to insert invalid JSON into the PostgreSQL JSONB `metadata` field.

**Solution:**
- When creating licenses without metadata, pass `null` or omit the field entirely
- When including metadata, ensure it's valid JSON format
- The system now automatically handles null values for metadata

**Example Fix:**
```bash
# ✅ Correct - No metadata
curl -X POST /v1/admin/licenses \
  -d '{"key": "TEST-001", "max_activations": 5}'

# ✅ Correct - Valid JSON metadata
curl -X POST /v1/admin/licenses \
  -d '{"key": "TEST-002", "max_activations": 5, "metadata": "{\"plan\": \"premium\"}"}'

# ❌ Incorrect - Invalid JSON
curl -X POST /v1/admin/licenses \
  -d '{"key": "TEST-003", "max_activations": 5, "metadata": "invalid"}'
```

### 2. Docker Build Issues

**Error:** `fork/exec docker-buildx: no such file or directory`

**Solution:**
1. Ensure Docker Desktop is running
2. Update Docker to the latest version
3. If using Docker Compose, remove the deprecated `version` field from docker-compose.yml

### 3. Database Connection Issues

**Error:** Connection refused to database

**Solution:**
1. Verify PostgreSQL is running: `docker-compose ps`
2. Check database logs: `docker-compose logs db`
3. Ensure the database URL is correct in your environment variables
4. Wait for health checks to pass before starting the API

### 4. Migration Failures

**Error:** Migration command not found or failed

**Solution:**
1. Install golang-migrate: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
2. Add GOPATH/bin to your PATH: `export PATH="$PATH:$(go env GOPATH)/bin"`
3. Run migrations manually: `./scripts/run_migrations.sh`

### 5. Rate Limiting Issues

**Error:** 429 Too Many Requests

**Solution:**
1. Check rate limit configuration in environment variables
2. Implement exponential backoff in your client
3. Consider increasing rate limits for your use case

### 6. HMAC Signature Errors

**Error:** Invalid request signature

**Solution:**
1. Ensure timestamp is current (within 5 minutes)
2. Verify HMAC secret matches between client and server
3. Check signature generation: `timestamp + clientId + body`
4. Use exact same encoding (UTF-8) for all components

### 7. JWT Token Issues

**Error:** Invalid or expired JWT token

**Solution:**
1. Generate a new token with `tools/jwtgen`: `./tools/jwtgen/jwtgen --secret "$JWT_ADMIN_SECRET" --sub admin`
2. Check token expiration (24 hours default)
3. Verify JWT secret configuration

## Testing the Fix

Run the test script to verify everything is working:

```bash
# Start the server first (with Docker or locally)
./scripts/test_fix.sh
```

This will test both metadata and non-metadata license creation scenarios.

## Debugging

### Enable Debug Logging

Set `LOG_LEVEL=debug` in your environment to get detailed request/response logging.

### Database Inspection

Connect to the database directly:
```bash
docker-compose exec db psql -U proxly -d proxly
\dt  # List tables
SELECT * FROM licenses LIMIT 5;
```

### Health Check

Verify the service is healthy:
```bash
curl http://localhost:8080/healthz
```

### View Metrics

Check Prometheus metrics:
```bash
curl http://localhost:8080/metrics
```
