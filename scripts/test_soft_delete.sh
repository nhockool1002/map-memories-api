#!/bin/bash

echo "Testing Soft Delete Functionality"
echo "================================="

# Build the application
echo "Building application..."
go build -o map-memories-api main.go

# Run migrations
echo "Running migrations..."
go run cmd/migrate/main.go

# Test 1: Create a location
echo ""
echo "Test 1: Creating a test location..."
curl -X POST http://localhost:8222/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "Test Location",
    "description": "A test location for soft delete",
    "latitude": 10.0,
    "longitude": 20.0,
    "address": "Test Address",
    "country": "Test Country",
    "city": "Test City"
  }'

echo ""
echo "Test 2: Creating a memory for this location..."
curl -X POST http://localhost:8222/api/v1/memories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "location_id": 1,
    "title": "Test Memory",
    "content": "This is a test memory for soft delete",
    "is_public": true
  }'

echo ""
echo "Test 3: Getting locations (should show the test location)..."
curl -X GET http://localhost:8222/api/v1/locations

echo ""
echo "Test 4: Deleting the location..."
curl -X DELETE http://localhost:8222/api/v1/locations/LOCATION_UUID_HERE

echo ""
echo "Test 5: Getting locations again (should not show the deleted location)..."
curl -X GET http://localhost:8222/api/v1/locations

echo ""
echo "Test 6: Getting memories (should show memory without location)..."
curl -X GET http://localhost:8222/api/v1/memories

echo ""
echo "Soft delete test completed!"
echo "Check the responses above to verify the functionality." 