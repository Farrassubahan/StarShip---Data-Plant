package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"starship-dataplane/internal/config"
	"starship-dataplane/internal/health"
)

// Client bertugas mengelola komunikasi dari Data Plane menuju Control Plane.
// Pola: Data Plane bertindak sebagai agen/pekerja yang melapor status ke Control Plane.
type Client struct {
	cfg                     *config.Config
	health                  *health.Checker
	httpClient              *http.Client
	isControlPlaneReachable bool
	hasLoggedOffline        bool
}

// NodePayload adalah struktur data saat Data Plane mendaftarkan diri ke Control Plane.
type NodePayload struct {
	NodeID    string    `json:"node_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Port      string    `json:"port"`
}

// HeartbeatPayload adalah data detak jantung yang dikirim secara periodik ke Control Plane.
type HeartbeatPayload struct {
	NodeID    string    `json:"node_id"`
	Status    string    `json:"status"`
	IsReady   bool      `json:"is_ready"`
	Timestamp time.Time `json:"timestamp"`
}

// NewClient membuat instance baru penghubung Control Plane.
func NewClient(cfg *config.Config, h *health.Checker) *Client {
	return &Client{
		cfg:    cfg,
		health: h,
		httpClient: &http.Client{
			Timeout: 4 * time.Second, // Timeout agar tidak menggantung jika server lambat
		},
	}
}

// RegisterNode mengirim sinyal pendaftaran awal ke server Control Plane.
func (c *Client) RegisterNode(ctx context.Context) error {
	payload := NodePayload{
		NodeID:    c.cfg.NodeID,
		Role:      c.cfg.NodeRole,
		Status:    "ONLINE",
		Timestamp: time.Now().UTC(),
		Port:      c.cfg.Port,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gagal serialize register payload: %w", err)
	}

	targetURL := fmt.Sprintf("%s/api/v1/nodes/register", c.cfg.ControlPlaneURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Control plane belum aktif (skenario mandiri / dev lokal)
		log.Printf("[DATA-PLANE] ℹ️ Control Plane (%s) belum aktif. Berjalan dalam mode mandiri.", c.cfg.ControlPlaneURL)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.isControlPlaneReachable = true
		log.Printf("[DATA-PLANE] [OK] Berhasil registrasi ke Control Plane (%s)", targetURL)
	} else {
		log.Printf("[DATA-PLANE] [WARN] Registrasi ditolak dengan status: %d", resp.StatusCode)
	}

	return nil
}

// StartHeartbeat menjalankan loop pengiriman sinyal hidup (heartbeat) secara terus-menerus di latar belakang.
func (c *Client) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.HeartbeatInterval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.sendHeartbeat(ctx)
			}
		}
	}()
}

// sendHeartbeat mengirim 1 kali sinyal detak jantung.
func (c *Client) sendHeartbeat(ctx context.Context) {
	status := "HEALTHY"
	if !c.health.IsReady() {
		status = "DEGRADED"
	}

	payload := HeartbeatPayload{
		NodeID:    c.cfg.NodeID,
		Status:    status,
		IsReady:   c.health.IsReady(),
		Timestamp: time.Now().UTC(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	targetURL := fmt.Sprintf("%s/api/v1/nodes/heartbeat", c.cfg.ControlPlaneURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Jika offline, cukup beri info 1 kali di awal agar tidak spam terminal tiap 5 detik!
		if !c.hasLoggedOffline {
			log.Printf("[DATA-PLANE] ℹ️ Sinyal detak jantung berjalan di latar belakang (menunggu Control Plane online)...")
			c.hasLoggedOffline = true
			c.isControlPlaneReachable = false
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Log saat pertama kali terhubung kembali
		if !c.isControlPlaneReachable {
			log.Printf("[DATA-PLANE] [CONNECT] Terhubung ke Control Plane!")
			c.isControlPlaneReachable = true
			c.hasLoggedOffline = false
		}
	}
}
