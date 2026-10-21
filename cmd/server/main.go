package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"starship-dataplane/internal/config"
	"starship-dataplane/internal/controlplane"
	"starship-dataplane/internal/handlers"
	"starship-dataplane/internal/health"
)

func main() {
	banner := `
=====================================================
       🚀 STARSHIP DATA PLANE (WORKER NODE) 🚀       
=====================================================
 Komponen : Data Plane (Workload & Traffic Execution)
 Bahasa   : Golang (Standard Library)
 Proyek   : Starship Cloud Computing
=====================================================
`
	fmt.Print(banner)

	// 1. Membaca Konfigurasi (Bisa dari ENV di Kubernetes)
	cfg := config.LoadConfig()
	log.Printf("[INIT] Node ID          : %s", cfg.NodeID)
	log.Printf("[INIT] Node Role        : %s", cfg.NodeRole)
	log.Printf("[INIT] HTTP Port        : %s", cfg.Port)
	log.Printf("[INIT] Control Plane URL: %s", cfg.ControlPlaneURL)

	// 2. Menginisialisasi Pengecek Kesehatan (Liveness & Readiness)
	healthChecker := health.NewChecker(cfg.NodeID)

	// 3. Menyiapkan Klien Komunikasi ke Control Plane
	cpClient := controlplane.NewClient(cfg, healthChecker)

	// Context untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kirim pendaftaran awal ke Control Plane
	_ = cpClient.RegisterNode(ctx)

	// Nyalakan pengiriman detak jantung (Heartbeat) di latar belakang
	cpClient.StartHeartbeat(ctx)

	// 4. Membangun Router HTTP
	router := handlers.SetupRouter(cfg, healthChecker)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 5. Menjalankan Server di Goroutine terpisah
	go func() {
		log.Printf("[SERVER] 🌐 Server Data Plane aktif di http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Gagal menyalakan server: %v", err)
		}
	}()

	// 6. Graceful Shutdown (Menangkap sinyal Ctrl+C atau SIGTERM dari Kubernetes)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\n[SHUTDOWN] Menerima sinyal berhenti, mematikan servis dengan aman...")
	cancel() // Matikan heartbeat goroutine

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[SHUTDOWN] Server shutdown error: %v", err)
	}

	log.Println("[SHUTDOWN] 👋 Starship Data Plane berhenti dengan sukses.")
}
