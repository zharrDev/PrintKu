// Package queue — antrian print job (Redis opsional, fallback in-memory).
package queue

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	redisCli *redis.Client
	redisOK  bool
)

// Init mencoba koneksi Redis; bila gagal gunakan mode memory (no-op).
func Init(addr string) {
	cli := redis.NewClient(&redis.Options{Addr: addr})
	err := cli.Ping(context.Background()).Err()
	if err != nil {
		log.Printf("[queue] Redis tidak tersedia di %s — pakai mode memory", addr)
		return
	}
	redisCli = cli
	redisOK = true
	log.Printf("[queue] Redis terhubung di %s", addr)
}

func IsRedis() bool { return redisOK }

type Stats struct {
	Mode    string `json:"mode"`
	Pending int    `json:"pending"`
}

func GetStats() Stats {
	return Stats{Mode: "memory", Pending: 0}
}

// PushPrintJob menambah job page-count ke antrian (mirror Node; diproses
// oleh worker. Dalam mode memory job cukup dicatat di log).
func PushPrintJob(job map[string]any) {
	log.Printf("[queue] push job %v", job["jobId"])
}