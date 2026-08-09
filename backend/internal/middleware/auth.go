package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"printmart/backend/internal/config"
)

type AuthUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func signKey(cfg *config.Config) []byte {
	return []byte(cfg.JWTSecret)
}

func SignToken(cfg *config.Config, u AuthUser) (string, error) {
	claims := jwt.MapClaims{
		"sub":    u.ID,
		"userId": u.ID,
		"name":   u.Name,
		"email":  u.Email,
		"role":   u.Role,
		"exp":    time.Now().Add(cfg.JWTExpiresIn).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey(cfg))
}

func parseToken(c *gin.Context, cfg *config.Config) (AuthUser, bool) {
	header := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(header, "Bearer ")
	if tokenStr == header || tokenStr == "" {
		return AuthUser{}, false
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return signKey(cfg), nil
	})
	if err != nil || !token.Valid {
		return AuthUser{}, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return AuthUser{}, false
	}
	user := AuthUser{
		ID:    asString(claims["userId"]),
		Name:  asString(claims["name"]),
		Email: asString(claims["email"]),
		Role:  asString(claims["role"]),
	}
	if user.ID == "" {
		return AuthUser{}, false
	}
	return user, true
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := parseToken(c, cfg)
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
		if user, ok := parseToken(c, cfg); ok {
			c.Set("user", user)
		}
		c.Next()
	}
}