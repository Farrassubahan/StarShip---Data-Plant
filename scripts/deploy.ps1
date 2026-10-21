<#
.SYNOPSIS
    Script Otomasi Deployment Starship Data Plane (PowerShell)
    Menerapkan manifest Kubernetes, menunggu pod siap, dan melakukan verifikasi.
#>

param (
    [string]$Namespace = "default",
    [string]$TestHost = "http://localhost:8080"
)

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   OTOMASI DEPLOYMENT: STARSHIP DATA PLANE (KUBERNETES)  " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Terapkan seluruh manifest Kubernetes
Write-Host "`n[1/4] Menerapkan manifest Kubernetes..." -ForegroundColor Yellow
kubectl apply -f deployments/k8s/ -n $Namespace
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Gagal menerapkan manifest Kubernetes!" -ForegroundColor Red
    exit 1
}

# 2. Tunggu hingga Pod selesai booting dan lulus Readiness Probe
Write-Host "`n[2/4] Menunggu Pod siap (Rollout Status)..." -ForegroundColor Yellow
kubectl rollout status deployment/starship-dataplane -n $Namespace --timeout=90s
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Pod gagal siap dalam batas waktu 90 detik!" -ForegroundColor Red
    exit 1
}

# 3. Tampilkan status Pod, Service, dan Ingress
Write-Host "`n[3/4] Status Resource Kubernetes:" -ForegroundColor Yellow
kubectl get pods,svc,ingress -l project=starship -n $Namespace

# 4. Jalankan script verifikasi kesehatan
Write-Host "`n[4/4] Menjalankan Verifikasi Kesehatan Akhir..." -ForegroundColor Yellow
& "$PSScriptRoot\verify_health.ps1" -BaseUrl $TestHost
