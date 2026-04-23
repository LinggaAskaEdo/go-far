#!/bin/bash

echo "=========================================="
echo "Stopping Monitoring Stack"
echo "=========================================="

# Stop containers in reverse order
echo "Stopping Promtail..."
docker stop promtail-docker 2>/dev/null || true

echo "Stopping Grafana..."
docker stop grafana-docker 2>/dev/null || true

echo "Stopping Prometheus..."
docker stop prometheus-docker 2>/dev/null || true

echo "Stopping Tempo..."
docker stop tempo-docker 2>/dev/null || true

echo "Stopping Loki..."
docker stop loki-docker 2>/dev/null || true

echo ""
echo "Removing containers..."
docker rm promtail-docker grafana-docker prometheus-docker tempo-docker loki-docker 2>/dev/null || true

echo ""
echo "✓ All services stopped"
echo ""
