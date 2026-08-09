// PrintMart Backend — server API Go (Gin).
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"printmart/backend/internal/config"
	"printmart/backend/internal/handler"
	"printmart/backend/internal/queue"
	"printmart/backend/internal/repository"
	"printmart/backend/pkg/websocket"
)

func main() {
	seedOnly := flag.Bool("seed", false, "jalankan seed database lalu keluar")
	flag.Parse()

	cfg := config.Load()

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

	hub := websocket.NewHub()
	r := handler.NewRouter(db, cfg, hub)

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("[main] PrintMart API jalan di http://localhost%s (%s)", addr, cfg.AppEnv)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("[main] server error: %v", err)
	}
}
