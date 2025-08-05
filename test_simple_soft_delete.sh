#!/bin/bash

echo "Testing Soft Delete Functionality"
echo "================================"

# Base URL
BASE_URL="http://localhost"

# Test 1: Check if API is running
echo -e "\n1. Checking API health..."
HEALTH_RESPONSE=$(curl -s -X GET "$BASE_URL/health")
echo "Health Response: $HEALTH_RESPONSE"

# Test 2: Get all locations (should be empty initially)
echo -e "\n2. Getting all locations (should be empty)..."
GET_ALL_RESPONSE=$(curl -s -X GET "$BASE_URL/locations")
echo "Get All Locations Response: $GET_ALL_RESPONSE"

# Test 3: Create a test location
echo -e "\n3. Creating a test location..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/locations" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Location",
    "description": "Test location for soft delete",
    "latitude": 10.762622,
    "longitude": 106.660172,
    "address": "Test Address",
    "country": "Vietnam",
    "city": "Ho Chi Minh City"
  }')

echo "Create Response: $CREATE_RESPONSE"

# Extract location UUID from response
LOCATION_UUID=$(echo $CREATE_RESPONSE | grep -o '"uuid":"[^"]*"' | cut -d'"' -f4)

if [ -z "$LOCATION_UUID" ]; then
    echo "Failed to extract location UUID. Please check the create response."
    exit 1
fi

echo "Location UUID: $LOCATION_UUID"

# Test 4: Get the created location
echo -e "\n4. Getting the created location..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/locations/$LOCATION_UUID")
echo "Get Response: $GET_RESPONSE"

# Test 5: Get all locations (should show the created location)
echo -e "\n5. Getting all locations (should show the created location)..."
GET_ALL_RESPONSE=$(curl -s -X GET "$BASE_URL/locations")
echo "Get All Response: $GET_ALL_RESPONSE"

# Test 6: Delete the location (soft delete)
echo -e "\n6. Soft deleting the location..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/locations/$LOCATION_UUID")
echo "Delete Response: $DELETE_RESPONSE"

# Test 7: Try to get the deleted location (should not be found)
echo -e "\n7. Trying to get the deleted location (should not be found)..."
GET_DELETED_RESPONSE=$(curl -s -X GET "$BASE_URL/locations/$LOCATION_UUID")
echo "Get Deleted Response: $GET_DELETED_RESPONSE"

# Test 8: Get all locations (deleted location should not appear)
echo -e "\n8. Getting all locations (deleted location should not appear)..."
GET_ALL_RESPONSE=$(curl -s -X GET "$BASE_URL/locations")
echo "Get All Response: $GET_ALL_RESPONSE"

echo -e "\nTest completed!" 