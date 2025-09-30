#!/bin/bash

# Simple test to verify license creation flows

set -e

API_BASE_URL="http://localhost:8080/v1"
ADMIN_API_BASE_URL="http://localhost:8080/v1/admin"

TOKEN="${ADMIN_TOKEN:-}"
if [ -z "$TOKEN" ]; then
  echo "❌ ADMIN_TOKEN environment variable not set"
  echo "   Generate one with tools/jwtgen, then run:"
  echo "   ADMIN_TOKEN=\"<jwt-from-tools/jwtgen>\" $0"
  exit 1
fi

echo "🧪 Testing license creation without metadata..."

RESPONSE=$(curl -s -X POST "$ADMIN_API_BASE_URL/licenses" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"key": "TEST-001", "max_activations": 5}')

if echo "$RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
  echo "✅ License created successfully without metadata"
  LICENSE_ID=$(echo "$RESPONSE" | jq -r '.id')
  echo "License ID: $LICENSE_ID"
else
  echo "❌ Failed to create license without metadata"
  echo "Response: $RESPONSE"
  exit 1
fi

echo "📄 Testing license creation with JSON metadata..."
RESPONSE2=$(curl -s -X POST "$ADMIN_API_BASE_URL/licenses" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"key": "TEST-002", "max_activations": 10, "metadata": "{\"plan\": \"premium\", \"features\": [\"feature1\", \"feature2\"]}"}')

if echo "$RESPONSE2" | jq -e '.id' > /dev/null 2>&1; then
  echo "✅ License created successfully with JSON metadata"
  LICENSE_ID2=$(echo "$RESPONSE2" | jq -r '.id')
  echo "License ID: $LICENSE_ID2"
else
  echo "❌ Failed to create license with JSON metadata"
  echo "Response: $RESPONSE2"
  exit 1
fi

echo ""
echo "🎉 All tests passed!"
echo "✅ The JSON/JSONB metadata issue has been fixed"
