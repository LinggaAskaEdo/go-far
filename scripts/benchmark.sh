#!/bin/bash

SERVER_URL="http://localhost:8181"
TOTAL_REQUESTS=500
CONCURRENT_USERS=100

echo "=== Benchmark Script ==="
echo ""

echo "1. Logging in to get access token..."
LOGIN_RESPONSE=$(curl -s -X POST "$SERVER_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "x-app-lang: en" \
  -d '{
    "email": "admin@gofar.com",
    "password": "AdminPass123!"
  }')

echo "Login response: $LOGIN_RESPONSE"

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)

# if [ -z "$TOKEN" ]; then
#   echo "Failed to get access token. Using provided token..."
#   TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjNDWUxyTGRHd0JaUU05dUxxWWdzZXFxd1JLYiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTc3NjU1ODA1NCwibmFtZSI6IkFkbWluIFVzZXIiLCJyb2xlIjoiYWRtaW4iLCJ1c2VyX2lkIjoiMDE5ZDhhMzktOGRjYS03NjNkLTk0MjMtYWU1MTg1ODY0NGYyIn0.60ExFxxrd5YID6xaAkESK8pZo5btkK9AftYbYAlCQ0U"
# fi

echo "Access Token: ${TOKEN:0:50}..."
echo ""

echo "=========================================================="
echo "2. Benchmarking /users endpoint (v1)..."
echo "   Total Requests: $TOTAL_REQUESTS, Concurrent: $CONCURRENT_USERS"
echo "=========================================================="

ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS -H "Authorization: Bearer $TOKEN" -H "x-app-lang: en" -H "cache-control: must-revalidate" "$SERVER_URL/users?name=Regular%20User&email=user%40gofar.com&min_age=25&max_age=30&page=1&page_size=10&sort_by=created_at&sort_dir=desc" > /tmp/benchmark_v1.txt 2>&1

cat /tmp/benchmark_v1.txt | grep -E "Requests per second|Time per request|Failed requests|Transfer rate"
echo ""

echo "=========================================================="
echo "3. Benchmarking /v2/users endpoint..."
echo "   Total Requests: $TOTAL_REQUESTS, Concurrent: $CONCURRENT_USERS"
echo "=========================================================="

ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS -H "Authorization: Bearer $TOKEN" -H "x-app-lang: en" -H "cache-control: must-revalidate" "$SERVER_URL/v2/users?name=Regular%20User&email=user%40gofar.com&min_age=25&max_age=30&page=1&page_size=10" > /tmp/benchmark_v2.txt 2>&1

cat /tmp/benchmark_v2.txt | grep -E "Requests per second|Time per request|Failed requests|Transfer rate"
echo ""

echo "=========================================================="
echo "=== SUMMARY ==="
echo "=========================================================="

echo ""
echo "--- /users (v1) ---"
grep "Requests per second" /tmp/benchmark_v1.txt | head -1
grep "Time per request" /tmp/benchmark_v1.txt | head -1

echo ""
echo "--- /v2/users ---"
grep "Requests per second" /tmp/benchmark_v2.txt | head -1
grep "Time per request" /tmp/benchmark_v2.txt | head -1

echo ""
echo "Raw results saved to /tmp/benchmark_v1.txt and /tmp/benchmark_v2.txt"