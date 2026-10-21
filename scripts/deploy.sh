#!/usr/bin/env bash
# ==========================================================
# Script Otomasi Deployment Starship Data Plane (Bash)
# ==========================================================
set -e

NAMESPACE="${1:-default}"
TEST_HOST="${2:-http://localhost:8080}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=========================================================="
echo "🚀 MEMULAI OTOMASI DEPLOYMENT: STARSHIP DATA PLANE"
echo "=========================================================="

echo -e "\n[1/4] Menerapkan manifest Kubernetes..."
kubectl apply -f "$DIR/../deployments/k8s/" -n "$NAMESPACE"

echo -e "\n[2/4] Menunggu Pod siap (Rollout Wait)..."
kubectl rollout status deployment/starship-dataplane -n "$NAMESPACE" --timeout=90s

echo -e "\n[3/4] Status Resource Kubernetes:"
kubectl get pods,svc,ingress -l project=starship -n "$NAMESPACE"

echo -e "\n[4/4] Menjalankan Verifikasi Kesehatan Akhir..."
"$DIR/verify_health.sh" "$TEST_HOST"
