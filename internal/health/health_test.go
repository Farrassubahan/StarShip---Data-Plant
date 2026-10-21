package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessHandler(t *testing.T) {
	checker := NewChecker("test-node-01")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	checker.LivenessHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("LivenessHandler mengembalikan status code salah: didapat %v seharusnya %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Gagal parse response JSON: %v", err)
	}

	if resp["status"] != "ALIVE" {
		t.Errorf("Status salah: didapat %v seharusnya 'ALIVE'", resp["status"])
	}
}

func TestReadinessHandler(t *testing.T) {
	checker := NewChecker("test-node-01")

	// Kondisi 1: Ready = true
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()

	checker.ReadinessHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ReadinessHandler (saat ready) mengembalikan %v seharusnya %v", status, http.StatusOK)
	}

	// Kondisi 2: Ready diubah menjadi false (Simulasi kegagalan servis)
	checker.SetReady(false)
	rrFalse := httptest.NewRecorder()
	checker.ReadinessHandler(rrFalse, req)

	if status := rrFalse.Code; status != http.StatusServiceUnavailable {
		t.Fatalf("ReadinessHandler (saat not ready) mengembalikan %v seharusnya %v (503)", status, http.StatusServiceUnavailable)
	}
}

func TestStatusHandler(t *testing.T) {
	checker := NewChecker("test-node-01")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rr := httptest.NewRecorder()

	checker.StatusHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("StatusHandler mengembalikan %v seharusnya %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Gagal parse response status: %v", err)
	}

	if resp["node_id"] != "test-node-01" {
		t.Errorf("Node ID salah: %v", resp["node_id"])
	}
}
