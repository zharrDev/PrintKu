package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               int
	AppEnv             string
	DatabasePath       string
	RedisAddr          string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	PaymentProvider    string
	PaymentIsProduction bool
	UploadDir          string
	MaxUploadSizeMB    int64
	CorsOrigin         string
	AdminEmail         string
	AdminPassword      string
}

func Load() *Config {
	loadDotEnv()
	return &Config{
		Port:                getInt("PORT", 8080),
		AppEnv:              getStr("APP_ENV", "development"),
		DatabasePath:        absPath(getStr("DATABASE_PATH", "./data/printmart.db")),
		RedisAddr:           getStr("REDIS_ADDR", "localhost:6379"),
		JWTSecret:           getStr("JWT_SECRET", "printmart-dev-secret"),
		JWTExpiresIn:        getDuration("JWT_EXPIRES_IN", 24*time.Hour),
		PaymentProvider:     getStr("PAYMENT_PROVIDER", "midtrans-sandbox"),
		PaymentIsProduction: getBool("PAYMENT_IS_PRODUCTION"),
		UploadDir:           absPath(getStr("UPLOAD_DIR", "./uploads")),
		MaxUploadSizeMB:     getInt64("MAX_UPLOAD_SIZE_MB", 20),
		CorsOrigin:          getStr("CORS_ORIGIN", "http://localhost:5173"),
		AdminEmail:          getStr("ADMIN_EMAIL", "admin@printmart.local"),
		AdminPassword:       getStr("ADMIN_PASSWORD", "admin123"),
	}
}

func getStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string) bool {
	return strings.EqualFold(os.Getenv(key), "true")
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	wd, err := os.Getwd()
	if err != nil {
		return p
	}
	return filepath.Join(wd, p)
}

func loadDotEnv() {
	path := filepath.Join(".env")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}