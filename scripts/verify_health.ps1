<#
.SYNOPSIS
    Script Verifikasi Kesehatan Starship Data Plane (PowerShell)
    Mengecek apakah servis beneran hidup dan siap melayani trafik.
#>

param (
    [string]$BaseUrl = ""
)

if ([string]::IsNullOrEmpty($BaseUrl)) {
    $BaseUrl = "http://localhost:8085"
    try {
        $null = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method Get -TimeoutSec 1 -ErrorAction Stop
    } catch {
        $BaseUrl = "http://localhost:8080"
    }
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   STARSHIP DATA PLANE: HEALTH & VERIFICATION CHECK" -ForegroundColor Cyan
Write-Host "   Target URL: $BaseUrl" -ForegroundColor Gray
Write-Host "==========================================================" -ForegroundColor Cyan

$allPassed = $true

# Fungsi pembantu untuk cek ketersediaan server
function Test-HttpEndpoint {
    param(
        [string]$Url,
        [int]$TimeoutSec = 3
    )
    try {
        $response = Invoke-RestMethod -Uri $Url -Method Get -TimeoutSec $TimeoutSec -ErrorAction Stop
        return @{ Success = $true; Data = $response; Error = $null }
    } catch {
        $statusCode = $null
        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        }
        return @{ Success = $false; Data = $null; StatusCode = $statusCode; Error = $_.Exception.Message }
    }
}

# 1. Cek Liveness (/healthz)
Write-Host "`n[1/3] Uji Liveness Probe (/healthz)..." -NoNewline
$resLiveness = Test-HttpEndpoint -Url "$BaseUrl/healthz"
if ($resLiveness.Success -and $resLiveness.Data.status -eq "ALIVE") {
    Write-Host " [PASS]" -ForegroundColor Green
    Write-Host "      Status Aplikasi : ALIVE" -ForegroundColor DarkGray
    Write-Host "      Uptime Server   : $($resLiveness.Data.uptime)" -ForegroundColor DarkGray
} else {
    Write-Host " [FAIL]" -ForegroundColor Red
    $allPassed = $false
    if ($null -eq $resLiveness.StatusCode) {
        Write-Host "      Catatan : Server Data Plane belum berjalan di $BaseUrl" -ForegroundColor Yellow
        Write-Host "                Silakan jalankan 'go run ./cmd/server' terlebih dahulu." -ForegroundColor DarkGray
    } else {
        Write-Host "      HTTP Status: $($resLiveness.StatusCode)" -ForegroundColor Yellow
    }
}

# 2. Cek Readiness (/readyz)
Write-Host "`n[2/3] Uji Readiness Probe (/readyz)..." -NoNewline
$resReady = Test-HttpEndpoint -Url "$BaseUrl/readyz"
if ($resReady.Success -and $resReady.Data.status -eq "READY") {
    Write-Host " [PASS]" -ForegroundColor Green
    Write-Host "      Status Trafik   : READY (Siap menerima trafik Ingress)" -ForegroundColor DarkGray
    Write-Host "      Node ID         : $($resReady.Data.node_id)" -ForegroundColor DarkGray
} else {
    Write-Host " [FAIL]" -ForegroundColor Red
    $allPassed = $false
    if ($resReady.StatusCode -eq 503) {
        Write-Host "      Status : 503 SERVICE UNAVAILABLE (Pod sedang tidak siap)" -ForegroundColor Yellow
    } else {
        Write-Host "      Detail : $($resReady.Error)" -ForegroundColor DarkGray
    }
}

# 3. Cek Workload Execution (/api/v1/workload)
Write-Host "`n[3/3] Uji Komputasi Workload (/api/v1/workload)..." -NoNewline
$resWorkload = Test-HttpEndpoint -Url "$BaseUrl/api/v1/workload" -TimeoutSec 5
if ($resWorkload.Success -and $resWorkload.Data.status -eq "SUCCESS") {
    Write-Host " [PASS]" -ForegroundColor Green
    Write-Host "      Tugas Eksekusi  : $($resWorkload.Data.task)" -ForegroundColor DarkGray
    Write-Host "      Waktu Eksekusi  : $($resWorkload.Data.execution_time_ms) ms" -ForegroundColor DarkGray
    Write-Host "      Hasil Checksum  : $($resWorkload.Data.computed_checksum)" -ForegroundColor DarkGray
} else {
    Write-Host " [FAIL]" -ForegroundColor Red
    $allPassed = $false
}

Write-Host "`n----------------------------------------------------------" -ForegroundColor Cyan
if ($allPassed) {
    Write-Host "  HASIL: SEMPURNA! Starship Data Plane BENERAN HIDUP & SEHAT." -ForegroundColor Green
    Write-Host "==========================================================`n" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  HASIL: Pengecekan belum lolos." -ForegroundColor Yellow
    Write-Host "  Tips : Pastikan server sudah jalan dengan: go run ./cmd/server" -ForegroundColor DarkGray
    Write-Host "==========================================================`n" -ForegroundColor Cyan
    exit 1
}
