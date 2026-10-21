package config

import (
	"os"
	"strconv"
	"time"
)

// Config menyimpan seluruh konfigurasi aplikasi Data Plane.
// Nilai bisa diubah melalui Environment Variable (berguna di Kubernetes).
type Config struct {
	Port              string        // Port HTTP server Data Plane (default: 8080)
	NodeID            string        // Identitas unik worker node ini
	NodeRole          string        // Peran node (default: "data-plane")
	ControlPlaneURL   string        // Alamat server Control Plane untuk pelaporan heartbeat
	HeartbeatInterval time.Duration // Jarak waktu pengiriman sinyal hidup ke Control Plane
}

// LoadConfig membaca variabel lingkungan atau menggunakan nilai default.
func LoadConfig() *Config {
	port := getEnv("PORT", "8080")
	nodeID := getEnv("NODE_ID", "starship-worker-01")
	nodeRole := getEnv("NODE_ROLE", "data-plane")
	controlPlaneURL := getEnv("CONTROL_PLANE_URL", "http://localhost:9090")

	intervalSecondsStr := getEnv("HEARTBEAT_INTERVAL_SECONDS", "5")
	intervalSeconds, err := strconv.Atoi(intervalSecondsStr)
	if err != nil {
		intervalSeconds = 5
	}

	return &Config{
		Port:              port,
		NodeID:            nodeID,
		NodeRole:          nodeRole,
		ControlPlaneURL:   controlPlaneURL,
		HeartbeatInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

// getEnv adalah fungsi pembantu untuk membaca environment variable dengan fallback default.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
