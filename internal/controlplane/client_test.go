package controlplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"starship-dataplane/internal/config"
	"starship-dataplane/internal/health"
)

func TestControlPlaneClient(t *testing.T) {
	var registerCalled int32
	var heartbeatCalled int32

	// Buat Mock Server Control Plane
	mockCP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/nodes/register":
			atomic.AddInt32(&registerCalled, 1)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "REGISTERED"})
		case "/api/v1/nodes/heartbeat":
			atomic.AddInt32(&heartbeatCalled, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ACK"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockCP.Close()

	cfg := &config.Config{
		Port:              "8080",
		NodeID:            "test-worker-01",
		NodeRole:          "data-plane",
		ControlPlaneURL:   mockCP.URL,
		HeartbeatInterval: 50 * time.Millisecond,
	}

	checker := health.NewChecker(cfg.NodeID)
	client := NewClient(cfg, checker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Test Registrasi Node
	err := client.RegisterNode(ctx)
	if err != nil {
		t.Fatalf("RegisterNode error: %v", err)
	}

	if atomic.LoadInt32(&registerCalled) != 1 {
		t.Errorf("Register endpoint seharusnya dipanggil 1 kali, tapi didapat: %d", registerCalled)
	}

	// 2. Test Heartbeat Loop
	client.StartHeartbeat(ctx)
	time.Sleep(120 * time.Millisecond) // Beri waktu ticker berjalan minimal 2 kali
	cancel()

	hbCount := atomic.LoadInt32(&heartbeatCalled)
	if hbCount < 1 {
		t.Errorf("Heartbeat seharusnya dipanggil minimal 1 kali, tapi didapat: %d", hbCount)
	}
}

func TestControlPlaneClient_OfflineGraceful(t *testing.T) {
	// Pastikan jika Control Plane mati / URL salah, Data Plane TIDAK panik / crash
	cfg := &config.Config{
		Port:              "8080",
		NodeID:            "test-worker-01",
		NodeRole:          "data-plane",
		ControlPlaneURL:   "http://127.0.0.1:54321", // Port mati
		HeartbeatInterval: 50 * time.Millisecond,
	}

	checker := health.NewChecker(cfg.NodeID)
	client := NewClient(cfg, checker)

	ctx := context.Background()

	// Harus selesai tanpa panic ataupun error fatal
	err := client.RegisterNode(ctx)
	if err != nil {
		t.Errorf("Harus graceful saat server offline, tapi didapat error: %v", err)
	}

	client.sendHeartbeat(ctx)
}
