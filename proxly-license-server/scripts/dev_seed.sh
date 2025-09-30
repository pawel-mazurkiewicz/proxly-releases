#!/bin/bash

# Development seed script for Proxly License Server
# This script creates demo licenses and webhooks for testing

set -e

API_BASE_URL="http://localhost:3939/v1"
ADMIN_API_BASE_URL="http://localhost:3939/v1/admin"

echo "🌱 Seeding Proxly License Server with demo data..."

# Expect token in environment (generate with tools/jwtgen)
TOKEN="${ADMIN_TOKEN:-}"
if [ -z "$TOKEN" ]; then
  echo "❌ ADMIN_TOKEN environment variable not set"
  echo "   Generate one with tools/jwtgen, e.g."
  echo '   TOKEN=$(tools/jwtgen/jwtgen --secret "$JWT_ADMIN_SECRET" --sub admin)'
  echo '   ADMIN_TOKEN="<jwt-from-tools/jwtgen>" scripts/dev_seed.sh'
  exit 1
fi

echo "✅ Using provided ADMIN_TOKEN"

# Create demo licenses
echo "📄 Creating demo licenses..."

LICENSES=(
  '{"key": "DEMO-001", "max_activations": 5}'
  '{"key": "DEMO-002", "max_activations": 10, "metadata": "{\"plan\": \"standard\", \"features\": [\"feature1\"]}"}'
  '{"key": "DEMO-003", "max_activations": 1}'
  '{"key": "UNLIMITED-001", "max_activations": 999999, "metadata": "{\"plan\": \"enterprise\", \"features\": [\"feature1\", \"feature2\", \"feature3\"]}"}'
)

for license in "${LICENSES[@]}"; do
  KEY=$(echo "$license" | jq -r '.key')
  echo "Creating license: $KEY"

  RESPONSE=$(curl -s -X POST "$ADMIN_API_BASE_URL/licenses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$license")

  if echo "$RESPONSE" | jq -e '.id' > /dev/null; then
    echo "✅ License $KEY created successfully"
  else
    echo "❌ Failed to create license $KEY"
    echo "Response: $RESPONSE"
  fi
done

# Create demo webhook
echo "🪝 Creating demo webhook..."
WEBHOOK_RESPONSE=$(curl -s -X POST "$ADMIN_API_BASE_URL/webhooks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "url": "http://localhost:3939/webhook",
    "secret": "demo-webhook-secret",
    "events": ["threshold.crossed", "activation.exceeded", "license.revoked"]
  }')

if echo "$WEBHOOK_RESPONSE" | jq -e '.id' > /dev/null; then
  echo "✅ Demo webhook created successfully"
else
  echo "❌ Failed to create demo webhook"
  echo "Response: $WEBHOOK_RESPONSE"
fi

echo ""
echo "🎉 Seeding complete!"
echo ""
echo "📋 Demo Data Summary:"
echo "  • Licenses created: DEMO-001 (5 max), DEMO-002 (10 max), DEMO-003 (1 max), UNLIMITED-001 (999999 max)"
echo "  • Webhook created: http://localhost:3939/webhook"
echo ""
echo "🧪 Test activation with:"
echo "curl -X POST $API_BASE_URL/activate \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'X-Timestamp: \$(date +%s)' \\"
echo "  -H 'X-Client-ID: test-client' \\"
echo "  -H 'X-Signature: <hmac-signature>' \\"
echo "  -d '{\"license_key\": \"DEMO-001\", \"client_id\": \"test-client\"}'"
