// Package queue — antrian print job dengan Redis atau fallback in-memory.
// Untuk production, Redis harus tersedia. Fallback memory hanya untuk development.
package queue

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisCli  *redis.Client
	redisOK   bool
	memoryJobs []printJob
	memoryMu   sync.Mutex
	cancelFunc context.CancelFunc
)

const (
	QueueKey       = "printku:jobs:pending"
	ProcessingKey  = "printku:jobs:processing"
	DeadLetterKey  = "printku:jobs:dead_letter"
	MaxRetries     = 3
	RetryDelay     = 5 * time.Second
)

type printJob struct {
	ID        string         `json:"id"`
	OrderID   string         `json:"order_id"`
	Retry     int            `json:"retry"`
	CreatedAt time.Time      `json:"created_at"`
	Data      map[string]any `json:"data"`
}

type Stats struct {
	Mode       string `json:"mode"`
	Pending    int    `json:"pending"`
	Processing int    `json:"processing"`
	Failed     int    `json:"failed"`
}

// Init mencoba koneksi Redis; bila gagal gunakan mode memory (hanya development).
func Init(addr string) {
	if addr == "" {
		log.Printf("[queue] Redis addr kosong — pakai mode memory (development only)")
		return
	}

	cli := redis.NewClient(&redis.Options{
		Addr:        addr,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cli.Ping(ctx).Err()
	if err != nil {
		log.Printf("[queue] Redis tidak tersedia di %s — pakai mode memory (development only)", addr)
		return
	}

	redisCli = cli
	redisOK = true
	log.Printf("[queue] Redis terhubung di %s", addr)

	// Mulai worker untuk memproses job
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancelFunc = cancel2
	go startWorker(ctx2)
}

// Stop menghentikan worker secara graceful.
func Stop() {
	if cancelFunc != nil {
		cancelFunc()
	}
	if redisOK && redisCli != nil {
		redisCli.Close()
	}
}

func IsRedis() bool { return redisOK }

// PushPrintJob menambah job ke antrian.
func PushPrintJob(job map[string]any) {
	if redisOK {
		pushRedis(job)
	} else {
		pushMemory(job)
	}
}

func pushRedis(job map[string]any) {
	j := printJob{
		ID:        getString(job, "jobId"),
		OrderID:   getString(job, "orderId"),
		CreatedAt: time.Now(),
		Data:      job,
	}
	data, err := json.Marshal(j)
	if err != nil {
		log.Printf("[queue] gagal marshal job: %v", err)
		return
	}
	ctx := context.Background()
	err = redisCli.LPush(ctx, QueueKey, data).Err()
	if err != nil {
		log.Printf("[queue] gagal push job ke Redis: %v", err)
		return
	}
	log.Printf("[queue] job %s di-enqueue ke Redis", j.ID)
}

func pushMemory(job map[string]any) {
	memoryMu.Lock()
	defer memoryMu.Unlock()
	j := printJob{
		ID:        getString(job, "jobId"),
		OrderID:   getString(job, "orderId"),
		CreatedAt: time.Now(),
		Data:      job,
	}
	memoryJobs = append(memoryJobs, j)
	log.Printf("[queue] job %s di-enqueue ke memory (%d pending)", j.ID, len(memoryJobs))
}

// startWorker memproses job dari Redis queue secara berkala.
func startWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[queue] worker stopped")
			return
		case <-ticker.C:
			processJobs(ctx)
		}
	}
}

func processJobs(ctx context.Context) {
	for {
		// Ambil job dari queue (blocking pop dengan timeout)
		result, err := redisCli.BRPop(ctx, 1*time.Second, QueueKey).Result()
		if err != nil {
			if err == redis.Nil || ctx.Err() != nil {
				return
			}
			log.Printf("[queue] error pop job: %v", err)
			return
		}

		if len(result) < 2 {
			continue
		}

		var j printJob
		if err := json.Unmarshal([]byte(result[1]), &j); err != nil {
			log.Printf("[queue] gagal unmarshal job: %v", err)
			continue
		}

		log.Printf("[queue] memproses job %s (retry: %d)", j.ID, j.Retry)

		// Simulasi pemrosesan job
		// Di production, ini akan memanggil service pemrosesan file sebenarnya
		success := processPrintJob(j)

		if !success {
			if j.Retry < MaxRetries {
				// Retry dengan delay
				j.Retry++
				j.CreatedAt = time.Now()
				data, _ := json.Marshal(j)

				go func() {
					time.Sleep(RetryDelay * time.Duration(j.Retry))
					redisCli.LPush(ctx, QueueKey, data)
					log.Printf("[queue] job %s di-retry (attempt %d)", j.ID, j.Retry)
				}()
			} else {
				// Max retries exceeded — kirim ke dead letter
				data, _ := json.Marshal(j)
				redisCli.LPush(ctx, DeadLetterKey, data)
				log.Printf("[queue] job %s gagal setelah %d retry — di-dead-letter", j.ID, MaxRetries)
			}
		} else {
			log.Printf("[queue] job %s berhasil diproses", j.ID)
		}
	}
}

// processPrintJob memproses satu print job.
// Di production, ini akan memanggil service pemrosesan file sebenarnya.
func processPrintJob(j printJob) bool {
	// Simulasi: selalu berhasil
	// Di production:
	// 1. Download file dari storage
	// 2. Proses sesuai spesifikasi
	// 3. Update status di database
	// 4. Notify user via WebSocket
	return true
}

func GetStats() Stats {
	if redisOK {
		return getRedisStats()
	}
	return getMemoryStats()
}

func getRedisStats() Stats {
	ctx := context.Background()
	pending, _ := redisCli.LLen(ctx, QueueKey).Result()
	processing, _ := redisCli.LLen(ctx, ProcessingKey).Result()
	failed, _ := redisCli.LLen(ctx, DeadLetterKey).Result()
	return Stats{
		Mode:       "redis",
		Pending:    int(pending),
		Processing: int(processing),
		Failed:     int(failed),
	}
}

func getMemoryStats() Stats {
	memoryMu.Lock()
	defer memoryMu.Unlock()
	return Stats{
		Mode:       "memory",
		Pending:    len(memoryJobs),
		Processing: 0,
		Failed:     0,
	}
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
