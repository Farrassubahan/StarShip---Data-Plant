package health

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Checker mengelola status kesehatan (Liveness & Readiness) dari Data Plane.
type Checker struct {
	mu        sync.RWMutex
	isReady   bool
	startTime time.Time
	nodeID    string
}

// NewChecker menginisialisasi sistem health check.
func NewChecker(nodeID string) *Checker {
	return &Checker{
		isReady:   true, // Awalnya langsung siap (bisa diubah sesuai inisialisasi modul)
		startTime: time.Now(),
		nodeID:    nodeID,
	}
}

// SetReady mengubah status kesiapan melayani trafik.
func (c *Checker) SetReady(ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isReady = ready
}

// IsReady mengecek apakah Data Plane sedang siap.
func (c *Checker) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isReady
}

// LivenessHandler menangani endpoint /healthz untuk Kubernetes Liveness Probe.
// Tujuannya: Memastikan proses aplikasi tidak macet (deadlock/hang).
// Jika endpoint ini gagal, Kubernetes akan ME-RESTART pod secara otomatis.
func (c *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "ALIVE",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(c.startTime).String(),
	}
	_ = json.NewEncoder(w).Encode(response)
}

// ReadinessHandler menangani endpoint /readyz untuk Kubernetes Readiness Probe.
// Tujuannya: Memastikan pod sudah siap menerima trafik nyata dari Ingress / Service.
// Jika endpoint ini mengembalikan 503, Ingress TIDAK AKAN mengirim trafik ke pod ini.
func (c *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	c.mu.RLock()
	ready := c.isReady
	c.mu.RUnlock()

	if !ready {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "NOT_READY",
			"message": "Data plane sedang inisialisasi atau dalam perbaikan internal",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "READY",
		"node_id": c.nodeID,
	})
}

// StatusHandler mengembalikan informasi detail sistem untuk monitoring atau dashboard.
func (c *Checker) StatusHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	status := map[string]interface{}{
		"node_id":    c.nodeID,
		"status":     "RUNNING",
		"is_ready":   c.IsReady(),
		"uptime":     time.Since(c.startTime).String(),
		"num_cpu":    runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
		"memory_alloc_mb": float64(m.Alloc) / 1024 / 1024,
	}
	_ = json.NewEncoder(w).Encode(status)
}
