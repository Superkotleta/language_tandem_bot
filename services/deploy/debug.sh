#!/bin/bash

echo "=== Docker Debug Script ==="

# Check container status
echo "=== Container status ==="
docker compose ps -a

# Get PostgreSQL IP
echo "=== PostgreSQL IP ==="
docker inspect pg --format '{{ .NetworkSettings.IPAddress }}'

# Check network connectivity
echo "=== Network status ==="
docker network inspect deploy_app-network | jq '.Containers'

# Get bot logs
echo "=== Last 20 bot logs ==="
docker compose logs bot --tail=20

# Check database from outside
echo "=== Test PostgreSQL connection ==="
docker compose exec postgres psql -U langbot -d languagebot -c "SELECT count(*) FROM interest_translations;" 2>/dev/null || echo "Database connection failed"
