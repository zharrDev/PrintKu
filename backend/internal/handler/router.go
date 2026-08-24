package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/config"
	"printmart/backend/pkg/websocket"
)

// NewRouter membangun *gin.Engine lengkap: CORS, download endpoint (auth),
// websocket, dan seluruh route API. Dipakai oleh cmd/server dan smoke test
// agar wiring selalu identik.
func NewRouter(db *sql.DB, cfg *config.Config, hub *websocket.Hub) *gin.Engine {
	srv := NewServer(db, cfg, hub)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	origins := strings.Split(cfg.CorsOrigin, ",")
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, o := range origins {
			if o == "*" || o == origin {
				allowed := "*"
				if o != "*" {
					allowed = origin
				}
				c.Header("Access-Control-Allow-Origin", allowed)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// JANGAN expose uploads directory secara publik!
	// Gunakan endpoint download yang membutuhkan authorization.
	r.GET("/ws/orders/:userId", hub.ServeWS)

	RegisterRoutes(r, srv)
	return r
}
