#!/bin/bash

# Simple script to test Agent mode API with JWT authentication

# Start Agent mode server in the background
echo "Starting Agent mode server..."
go run ./cmd/pos agent --port 8000 &
SERVER_PID=$!

# Give the server a moment to start
sleep 2

# Get JWT token by logging in
echo -e "\nGetting JWT token..."
JWT_TOKEN=$(go run ./cmd/pos staff login admin password123 | grep "JWT Token" | awk '{print $3}')

if [ -z "$JWT_TOKEN" ]; then
    echo "Failed to get JWT token!"
    kill $SERVER_PID
    exit 1
fi

echo "Got JWT token: ${JWT_TOKEN:0:20}..."

# Test the products endpoint with JWT authentication
echo -e "\nTesting /products endpoint with JWT authentication..."
curl -s -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    http://localhost:8000/products | jq

# Test the reports endpoint with JWT authentication
echo -e "\nTesting /reports/summary endpoint with JWT authentication..."
curl -s -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    http://localhost:8000/reports/summary | jq

# Test authentication error by using an invalid token
echo -e "\nTesting authentication error with invalid token..."
curl -s -X GET \
    -H "Authorization: Bearer invalidtoken" \
    -H "Content-Type: application/json" \
    http://localhost:8000/products 

# Clean up
echo -e "\nStopping Agent mode server..."
kill $SERVER_PID

echo -e "\nDone!"