#!/bin/bash

set -e

echo "🚀 Building Map Memories API Docker images..."

# Load environment variables
if [ -f .env.production ]; then
    export $(cat .env.production | grep -v '^#' | xargs)
fi

# Build API image
echo "📦 Building API image..."
docker build -t map-memories-api:latest .

# Build with specific tag
TAG=${1:-latest}
docker tag map-memories-api:latest map-memories-api:$TAG

echo "✅ Build completed successfully!"
echo "📋 Images created:"
docker images | grep map-memories-api 