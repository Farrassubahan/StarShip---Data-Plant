package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"starship-dataplane/internal/config"
	"starship-dataplane/internal/health"
)

// SetupRouter merangkai semua endpoint rute HTTP Data Plane.
func SetupRouter(cfg *config.Config, h *health.Checker) http.Handler {
	mux := http.NewServeMux()

	// 1. Endpoint Kubernetes Probes (Pengecekan Kesehatan)
	mux.HandleFunc("/healthz", h.LivenessHandler)
	mux.HandleFunc("/readyz", h.ReadinessHandler)
	mux.HandleFunc("/api/v1/status", h.StatusHandler)

	// 2. Endpoint Halaman Utama (Tampilan saat diakses via Browser / Ingress)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"project":     "Starship Cloud System",
			"component":   "Data Plane (Worker)",
			"node_id":     cfg.NodeID,
			"role":        cfg.NodeRole,
			"host_header": r.Host, // Menunjukkan domain apa yang dipakai user (membuktikan Ingress & Wildcard DNS bekerja!)
			"remote_addr": r.RemoteAddr,
			"message":     "🚀 Selamat datang di Starship Data Plane! Servis berhasil berjalan sempurna.",
			"available_endpoints": []string{
				"/healthz (Liveness check)",
				"/readyz (Readiness check)",
				"/api/v1/status (System telemetry)",
				"/api/v1/workload (Simulasi komputasi data)",
				"/api/v1/simulate/toggle-ready (Simulasi kegagalan servis untuk pengujian)",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	})

	// 3. Endpoint Workload / Komputasi (Fungsi inti Data Plane)
	mux.HandleFunc("/api/v1/workload", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Simulasi pekerjaan komputasi kecil
		total := 0
		for i := 1; i <= 100000; i++ {
			total += i
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "SUCCESS",
			"task":              "Data Processing Task",
			"computed_checksum": total,
			"execution_time_ms": time.Since(start).Seconds() * 1000,
			"processed_by":      cfg.NodeID,
		})
	})

	// 4. Endpoint Simulasi Crash / Not Ready (Sangat berguna untuk demo ke dosen/tim)
	mux.HandleFunc("/api/v1/simulate/toggle-ready", func(w http.ResponseWriter, r *http.Request) {
		current := h.IsReady()
		newState := !current
		h.SetReady(newState)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   fmt.Sprintf("Status readiness diubah dari %v menjadi %v", current, newState),
			"new_ready": newState,
			"notice":    "Jika status false, /readyz akan mengembalikan HTTP 503 dan Ingress akan berhenti mengalirkan trafik ke sini.",
		})
	})

	return mux
}
