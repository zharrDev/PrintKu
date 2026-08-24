// PrintKu Backend — server API Go (Gin).
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"printmart/backend/internal/config"
	"printmart/backend/internal/handler"
	"printmart/backend/internal/middleware"
	"printmart/backend/internal/queue"
	"printmart/backend/internal/repository"
	"printmart/backend/pkg/websocket"
)

func main() {
	seedOnly := flag.Bool("seed", false, "jalankan seed database lalu keluar")
	flag.Parse()

	cfg := config.Load()

	// Validasi production secrets — server harus gagal start
	// jika secret lemah di production.
	if err := middleware.ValidateProductionSecret(cfg); err != nil {
		log.Fatalf("[main] %v", err)
	}
	if err := middleware.ValidateProductionAdminPassword(cfg); err != nil {
		log.Fatalf("[main] %v", err)
	}

	db, err := repository.Open(cfg)
	if err != nil {
		log.Fatalf("[main] gagal buka database: %v", err)
	}
	defer db.Close()

	if err := repository.Seed(db, cfg); err != nil {
		log.Fatalf("[main] gagal seed: %v", err)
	}

	if *seedOnly {
		log.Printf("[main] seed selesai (%s). Keluar.", cfg.DatabasePath)
		return
	}

	queue.Init(cfg.RedisAddr)

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("[main] gagal buat upload dir: %v", err)
	}

	// Buat WebSocket hub dengan JWT secret dan CORS origins.
	origins := strings.Split(cfg.CorsOrigin, ",")
	hub := websocket.NewHub(cfg.JWTSecret, origins)

	// Jalankan cleanup loop untuk WebSocket
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.StartCleanupLoop(ctx)

	r := handler.NewRouter(db, cfg, hub)

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("[main] PrintKu API jalan di http://localhost%s (%s)", addr, cfg.AppEnv)

	// Gunakan http.Server dengan timeouts untuk keamanan
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("[main] menerima signal %v, shutdown graceful...", sig)

		// Batalkan cleanup loop
		cancel()

		// Stop queue worker
		queue.Stop()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("[main] shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[main] server error: %v", err)
	}
	log.Printf("[main] server stopped")
}
