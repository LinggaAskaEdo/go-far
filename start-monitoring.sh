#!/bin/bash

set -e  # Exit on error

echo "=========================================="
echo "Starting Monitoring Stack"
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if port is listening
wait_for_port() {
    local port=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}Waiting for $service_name on port $port...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if netstat -tuln 2>/dev/null | grep -q ":$port " || ss -tuln 2>/dev/null | grep -q ":$port "; then
            echo -e "${GREEN}✓ $service_name is listening on port $port${NC}"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}✗ $service_name failed to start on port $port${NC}"
    return 1
}

# Function to stop and remove container if exists
cleanup_container() {
    local container_name=$1
    if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
        echo "Removing existing container: $container_name"
        docker stop $container_name 2>/dev/null || true
        docker rm $container_name 2>/dev/null || true
    fi
}

# Cleanup function
cleanup() {
    echo -e "\n${RED}Shutting down...${NC}"
    docker stop loki-docker tempo-docker prometheus-docker grafana-docker promtail-docker 2>/dev/null || true
    exit 1
}

trap cleanup SIGINT SIGTERM

# Step 1: Start Loki
echo -e "\n${YELLOW}[1/5] Starting Loki...${NC}"
cleanup_container loki-docker

docker run -d \
	--name loki-docker \
	--network otis-network \
	-p 3100:3100 \
	-v loki-volume:/loki \
	grafana/loki:latest \
	-config.file=/etc/loki/local-config.yaml

wait_for_port 3100 "Loki"
	
# Step 2: Start Tempo
echo -e "\n${YELLOW}[2/5] Starting Tempo...${NC}"
cleanup_container tempo-docker

docker run -d \
	--name tempo-docker \
	--network otis-network \
	--user root \
	-p 3200:3200 \
	-p 4317:4317 \
	-p 4318:4318 \
	-v $(pwd)/etc/grafana/tempo.yaml:/etc/tempo.yaml \
	-v tempo-volume:/tmp/tempo \
	grafana/tempo:latest \
	-config.file=/etc/tempo.yaml

wait_for_port 4317 "Tempo"

# Step 3: Start Prometheus
echo -e "\n${YELLOW}[3/5] Starting Prometheus...${NC}"
cleanup_container prometheus-docker

docker run -d \
	--name prometheus-docker \
	--network otis-network \
	-p 9090:9090 \
	-v $(pwd)/etc/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
	-v prometheus-volume:/prometheus \
	prom/prometheus:latest

wait_for_port 9090 "Prometheus"
	
# Step 4: Start Grafana
echo -e "\n${YELLOW}[4/5] Starting Grafana...${NC}"
cleanup_container grafana-docker

docker run -d \
	--name grafana-docker \
	--network otis-network \
	-p 3000:3000 \
	-e GF_SECURITY_ADMIN_USER=admin \
    -e GF_SECURITY_ADMIN_PASSWORD=admin \
    -e GF_USERS_ALLOW_SIGN_UP=false \
	-e GF_FEATURE_TOGGLES_ENABLE=traceqlEditor \
	-v grafana-volume:/var/lib/grafana \
	-v $(pwd)/etc/grafana/provisioning:/etc/grafana/provisioning \
	grafana/grafana:latest

wait_for_port 3000 "Grafana"
	
# Step 5: Start Promtail
echo -e "\n${YELLOW}[5/5] Starting Promtail...${NC}"
cleanup_container promtail-docker

docker run -d \
	--name promtail-docker \
	--network otis-network \
	-p 9080:9080 \
	-v $(pwd)/etc/grafana/promtail-config.yaml:/etc/promtail/config.yaml \
	-v ./logs/:/var/log/app:ro \
	grafana/promtail:latest \
	-config.file=/etc/promtail/config.yaml

wait_for_port 9080 "Promtail"

# Summary
echo -e "\n${GREEN}=========================================="
echo "✓ All services started successfully!"
echo "==========================================${NC}"
