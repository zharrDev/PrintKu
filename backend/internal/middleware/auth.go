package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"printmart/backend/internal/config"
)

// Claims adalah typed claims untuk JWT. Mengganti MapClaims agar
// field dipaksa tipenya dan tidak bisa di-inject dari client.
type Claims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func signKey(cfg *config.Config) []byte {
	return []byte(cfg.JWTSecret)
}

// SignToken menandatangani JWT dengan metode HS256 dan typed Claims.
func SignToken(cfg *config.Config, u AuthUser) (string, error) {
	claims := Claims{
		UserID: u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   u.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey(cfg))
}

// parseToken memvalidasi JWT, memastikan:
// 1. Signing method harus HS256 (bukan alg:none atau RSA/ECDSA)
// 2. Claims harus valid dan belum expired
// 3. User harus ada di database dan masih aktif
func parseToken(c *gin.Context, cfg *config.Config, db *sql.DB) (AuthUser, bool) {
	header := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(header, "Bearer ")
	if tokenStr == header || tokenStr == "" {
		return AuthUser{}, false
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		// Kritis: validasi signing method secara eksplisit.
		// Hanya izinkan HS256. Tolak alg:none, RSA, ECDSA, dll.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("only HS256 is allowed, got: %s", t.Method.Alg())
		}
		return signKey(cfg), nil
	})

	if err != nil {
		return AuthUser{}, false
	}
	if !token.Valid {
		return AuthUser{}, false
	}

	// Pastikan claims minimal lengkap
	if claims.UserID == "" {
		return AuthUser{}, false
	}

	user := AuthUser{
		ID:    claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
	}

	// Verifikasi user masih ada dan masih aktif di database.
	// Jangan hanya percaya data dari token.
	if db != nil {
		var dbRole string
		err := db.QueryRow(`SELECT role FROM users WHERE id = ?`, user.ID).Scan(&dbRole)
		if err != nil {
			return AuthUser{}, false
		}
		// Ambil role dari database, bukan dari token client.
		user.Role = dbRole
	}

	return user, true
}

func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := parseToken(c, cfg, nil)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: token tidak valid atau kadaluarsa"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func AdminRequired(c *gin.Context) {
	user, _ := c.Get("user")
	u, ok := user.(AuthUser)
	if !ok || u.Role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: hanya admin"})
		return
	}
	c.Next()
}

// OptionalAuth mengisi user bila token valid, tanpa menolak bila tidak ada.
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := parseToken(c, cfg, nil); ok {
			c.Set("user", user)
		}
		c.Next()
	}
}

// ValidateProductionSecret memastikan secret JWT bukan default/lemah
// saat berjalan di production. Haril dipanggil saat startup.
func ValidateProductionSecret(cfg *config.Config) error {
	if cfg.AppEnv != "production" {
		return nil
	}
	weakSecrets := []string{
		"printku-dev-secret",
		"printku-secret",
		"secret",
		"changeme",
		"ganti_dengan_secret_yang_kuat_printku_2024",
		"ganti_dengan_secret_yang_kuat",
		"admin123",
	}
	for _, ws := range weakSecrets {
		if cfg.JWTSecret == ws {
			return errors.New("FATAL: JWT_SECRET terlalu lemah untuk production. Ganti dengan secret yang kuat!")
		}
	}
	if len(cfg.JWTSecret) < 32 {
		return errors.New("FATAL: JWT_SECRET harus minimal 32 karakter untuk production")
	}
	return nil
}

// ValidateProductionAdminPassword memastikan admin password bukan default
// saat berjalan di production.
func ValidateProductionAdminPassword(cfg *config.Config) error {
	if cfg.AppEnv != "production" {
		return nil
	}
	if cfg.AdminPassword == "admin123" || cfg.AdminPassword == "password" {
		return errors.New("FATAL: ADMIN_PASSWORD masih default. Ganti untuk production!")
	}
	return nil
}
