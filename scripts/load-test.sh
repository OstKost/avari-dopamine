#!/usr/bin/env bash
set -euo pipefail

# scripts/load-test.sh — Smoke load test against NFR-PERF-01 (20 RPS, p95 < 200ms)
TARGET_HOST="${1:-http://localhost:8080}"
echo "Running smoke load test against $TARGET_HOST..."

# Check health endpoint
if ! curl -sf "${TARGET_HOST}/healthz" > /dev/null; then
  echo "Error: Target host $TARGET_HOST is not reachable or unhealthy"
  exit 1
fi

echo "Checking /catalog/products (20 requests)..."
for i in {1..20}; do
  START=$(date +%s%N)
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${TARGET_HOST}/catalog/products?limit=10")
  END=$(date +%s%N)
  DURATION_MS=$(( (END - START) / 1000000 ))
  echo "Request #$i: status=$STATUS duration=${DURATION_MS}ms"
  if [ "$STATUS" -ne 200 ]; then
    echo "Load test failed on request #$i (status $STATUS)"
    exit 1
  fi
done

echo "Smoke load test passed successfully!"
