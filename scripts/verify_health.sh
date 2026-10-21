#!/usr/bin/env bash
# ==========================================================
# Script Verifikasi Kesehatan Starship Data Plane (Bash)
# ==========================================================

BASE_URL="${1:-http://localhost:8080}"

echo "=========================================================="
echo "🔍 MEMERIKSA KESEHATAN STARSHIP DATA PLANE ($BASE_URL)"
echo "=========================================================="

FAILED=0

# 1. Cek Liveness (/healthz)
echo -n "[1/3] Menguji Liveness Endpoint (/healthz)... "
STATUS_LIVENESS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz" || echo "000")
if [ "$STATUS_LIVENESS" -eq 200 ]; then
  echo "✅ OK (HTTP 200)"
else
  echo "❌ GAGAL (HTTP $STATUS_LIVENESS)"
  FAILED=1
fi

# 2. Cek Readiness (/readyz)
echo -n "[2/3] Menguji Readiness Endpoint (/readyz)... "
STATUS_READINESS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/readyz" || echo "000")
if [ "$STATUS_READINESS" -eq 200 ]; then
  echo "✅ OK (HTTP 200)"
else
  echo "❌ NOT READY (HTTP $STATUS_READINESS)"
  FAILED=1
fi

# 3. Cek Workload Execution (/api/v1/workload)
echo -n "[3/3] Menguji Pemrosesan Workload (/api/v1/workload)... "
STATUS_WORKLOAD=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/workload" || echo "000")
if [ "$STATUS_WORKLOAD" -eq 200 ]; then
  echo "✅ OK (HTTP 200)"
else
  echo "❌ GAGAL (HTTP $STATUS_WORKLOAD)"
  FAILED=1
fi

echo "----------------------------------------------------------"
if [ "$FAILED" -eq 0 ]; then
  echo "🎉 SEMPURNA: Starship Data Plane BENERAN HIDUP dan Sehat!"
  exit 0
else
  echo "⚠️ PERINGATAN: Ada pengecekan yang gagal."
  exit 1
fi
