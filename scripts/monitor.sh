#!/bin/bash

echo "🔍 Monitoring Map Memories API..."

# Check container status
echo "📊 Container Status:"
docker-compose -f docker-compose.prod.yml ps

# Check resource usage
echo "💾 Resource Usage:"
docker stats --no-stream

# Check logs
echo "📝 Recent Logs:"
docker-compose -f docker-compose.prod.yml logs --tail=20 api

# Health check
echo "🏥 Health Check:"
curl -f http://localhost/health || echo "❌ Health check failed" 