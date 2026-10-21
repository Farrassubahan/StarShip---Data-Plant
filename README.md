# 🚀 Starship Data Plane

Proyek **Data Plane** untuk mata kuliah Komputasi Awan (Cloud Computing) oleh kelompok **Starship**.  
Dibuat dengan **Golang (Standard Library)** agar bersih, cepat, tanpa dependensi rumit, dan mudah dipahami bersama tim.

---

## 📂 Struktur Folder Proyek

```text
├── cmd/
│   └── server/
│       └── main.go              # Titik masuk utama aplikasi (Entrypoint)
├── internal/
│   ├── config/config.go         # Konfigurasi & pembacaan Environment Variable
│   ├── controlplane/client.go   # Komunikasi ke Control Plane (Pendaftaran & Heartbeat)
│   ├── handlers/routes.go       # Rute HTTP (Workload, Simulasi, Landing Page)
│   └── health/health.go         # Pengecekan kesehatan (Liveness & Readiness)
├── deployments/k8s/
│   ├── 01-deployment.yaml       # Deployment K8s (2 Pod + Probes otomatis)
│   ├── 02-service.yaml          # Service internal ClusterIP
│   └── 03-ingress.yaml          # Ingress Controller + Wildcard DNS (*.starship.local)
├── scripts/
│   ├── deploy.ps1 / .sh         # Script otomatisasi deployment K8s
│   └── verify_health.ps1 / .sh  # Script pengecekan "servis beneran hidup apa enggak"
├── Dockerfile                   # Docker build multi-stage ringan
└── README.md
```

---

## 🏃 1. Cara Menjalankan Secara Lokal (Tanpa Kubernetes Dulu)

Buka terminal di folder ini, lalu ketik:

```powershell
go run ./cmd/server
```

Aplikasi Data Plane akan langsung menyala di port `8080`.

Buka browser atau terminal baru untuk mencoba endpoint:
* **Halaman Utama:** `http://localhost:8080/`
* **Liveness Probe:** `http://localhost:8080/healthz`
* **Readiness Probe:** `http://localhost:8080/readyz`
* **Telemetry Sistem:** `http://localhost:8080/api/v1/status`
* **Simulasi Workload Data:** `http://localhost:8080/api/v1/workload`

---

## 🧪 2. Menjalankan Verifikasi "Servis Beneran Hidup Apa Enggak"

Jalankan script verifikasi otomatis yang sudah disiapkan:

```powershell
# Di Windows PowerShell:
.\scripts\verify_health.ps1

# Di Linux / macOS / Git Bash:
./scripts/verify_health.sh
```

---

## 🐳 3. Menjalankan via Docker (Untuk Kelompok)

Jika ingin menjalankan aplikasi sebagai kontainer Docker yang portabel:

```powershell
# 1. Build image Data Plane:
docker build -t starship-dataplane:v1.0.0 .

# 2. Jalankan container (menggunakan port 8085 di laptop agar tidak bentrok):
docker run -d -p 8085:8080 --name dataplane starship-dataplane:v1.0.0

# 3. Atau cara paling gampang dengan docker-compose:
docker compose up -d
```

Untuk menghentikan:
```powershell
docker compose down
```

---

## 📡 4. Jalur Komunikasi ke Control Plane

Data Plane sudah dilengkapi agen internal:
1. **Registrasi Otomatis:** Saat pertama kali boot, Data Plane mengirim identitas node ke URL Control Plane (`POST /api/v1/nodes/register`).
2. **Heartbeat Loop (Detak Jantung):** Setiap beberapa detik, Data Plane mengirim sinyal status `HEALTHY` ke Control Plane (`POST /api/v1/nodes/heartbeat`).
3. **Resilient (Tahan Banting):** Jika Control Plane teman kelompok belum dinyalakan, Data Plane **tidak akan crash**, melainkan mencatat peringatan santai di log dan tetap melayani trafik pengguna.

---

## 🌐 4. Konfigurasi Ingress Controller & Wildcard DNS

Di file `deployments/k8s/03-ingress.yaml`:
* Rute utama memakai rule wildcard: `*.starship.local`.
* Apapun subdomain yang dipanggil (contoh: `api.starship.local`, `transaksi.starship.local`, `testing.starship.local`), Ingress Controller akan meneruskannya ke Service Data Plane.

---

## 📦 5. Otomasi Deployment ke Kubernetes

Cukup jalankan satu script otomasi:

```powershell
# Di Windows PowerShell:
.\scripts\deploy.ps1

# Di Linux / Git Bash:
./scripts/deploy.sh
```

Script tersebut otomatis:
1. Menerapkan Deployment, Service, dan Ingress ke cluster.
2. Menunggu (`rollout status`) sampai seluruh pod siap melayani.
3. Melakukan smoke test kesehatan secara otomatis.
