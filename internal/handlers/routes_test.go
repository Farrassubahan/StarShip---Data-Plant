package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"starship-dataplane/internal/config"
	"starship-dataplane/internal/health"
)

func TestRouterEndpoints(t *testing.T) {
	cfg := &config.Config{
		Port:              "8080",
		NodeID:            "starship-test-node",
		NodeRole:          "data-plane",
		ControlPlaneURL:   "http://localhost:9090",
		HeartbeatInterval: 1 * time.Second,
	}
	checker := health.NewChecker(cfg.NodeID)
	router := SetupRouter(cfg, checker)

	// Test 1: Root "/"
	t.Run("GET / Root Endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GET / return status %d, expected 200", rr.Code)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(rr.Body).Decode(&body)
		if body["node_id"] != "starship-test-node" {
			t.Errorf("Node ID tidak cocok: %v", body["node_id"])
		}
	})

	// Test 2: Workload "/api/v1/workload"
	t.Run("GET /api/v1/workload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workload", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GET /api/v1/workload return status %d, expected 200", rr.Code)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(rr.Body).Decode(&body)
		if body["status"] != "SUCCESS" {
			t.Errorf("Workload status salah: %v", body["status"])
		}
	})

	// Test 3: Toggle Ready "/api/v1/simulate/toggle-ready"
	t.Run("GET /api/v1/simulate/toggle-ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/simulate/toggle-ready", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Toggle ready return %d, expected 200", rr.Code)
		}

		if checker.IsReady() != false {
			t.Errorf("Seharusnya status ready berubah jadi false")
		}
	})

	// Test 4: 404 Not Found
	t.Run("GET /halaman-tidak-ada", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/halaman-tidak-ada", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for unknown route, got %d", rr.Code)
		}
	})
}
