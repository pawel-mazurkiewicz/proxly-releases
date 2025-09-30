#!/bin/bash

# Test script for license activation with proper HMAC signature
# This demonstrates how to properly activate a license

set -e

API_BASE_URL="http://localhost:3939/v1"
HMAC_SECRET="hmac-secret-key"  # This should match the server configuration

echo "🔑 Testing license activation with proper HMAC signature..."

# Generate current timestamp
TIMESTAMP=$(date +%s)
CLIENT_ID="test-client"
LICENSE_KEY="DEMO-001"

# Create the request body
BODY='{"license_key": "'$LICENSE_KEY'", "client_id": "'$CLIENT_ID'"}'

echo "📝 Request details:"
echo "  Timestamp: $TIMESTAMP"
echo "  Client ID: $CLIENT_ID"
echo "  Body: $BODY"

# Generate HMAC signature
# Format: timestamp + clientId + body
MESSAGE="${TIMESTAMP}${CLIENT_ID}${BODY}"
SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "$HMAC_SECRET" | cut -d' ' -f2)

echo "  Message to sign: $MESSAGE"
echo "  HMAC Secret: $HMAC_SECRET"
echo "  Signature: $SIGNATURE"

# Make the activation request
echo ""
echo "🚀 Making activation request..."

RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_BASE_URL/activate" \
  -H "Content-Type: application/json" \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Client-ID: $CLIENT_ID" \
  -H "X-Signature: $SIGNATURE" \
  -d "$BODY")

# Parse response and HTTP status
HTTP_BODY=$(echo "$RESPONSE" | sed '$d')
HTTP_STATUS=$(echo "$RESPONSE" | tail -1 | sed 's/HTTP_STATUS://')

echo "📋 Response:"
echo "  Status: $HTTP_STATUS"
echo "  Body: $HTTP_BODY"

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo ""
    echo "✅ License activation successful!"

    # Pretty print the response if jq is available
    if command -v jq &> /dev/null; then
        echo ""
        echo "📊 Activation Details:"
        echo "$HTTP_BODY" | jq .
    fi
else
    echo ""
    echo "❌ License activation failed!"
    echo "Check the server logs and HMAC secret configuration."
fi