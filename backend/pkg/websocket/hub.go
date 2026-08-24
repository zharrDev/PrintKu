// Package websocket — real-time hub untuk status pesanan.
// Keamanan:
// - Autentikasi koneksi via JWT token (query param atau header)
// - Origin allowlist dari konfigurasi
// - Ping/pong heartbeat dengan cleanup koneksi mati
// - User hanya boleh subscribe ke event miliknya sendiri
package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Claims harus match dengan middleware.Claims untuk validasi token.
type Claims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool

	// Konfigurasi keamanan
	jwtSecret    []byte
	allowedOrigins map[string]bool

	upgrader websocket.Upgrader
}

// NewHub membuat hub baru dengan konfigurasi keamanan.
func NewHub(jwtSecret string, allowedOrigins []string) *Hub {
	origins := make(map[string]bool)
	for _, o := range allowedOrigins {
		origins[o] = true
	}

	h := &Hub{
		clients:        make(map[string]map[*websocket.Conn]bool),
		jwtSecret:      []byte(jwtSecret),
		allowedOrigins: origins,
	}

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Jangan gunakan CheckOrigin yang selalu true.
		// Gunakan allowlist dari konfigurasi.
		CheckOrigin: h.checkOrigin,
	}

	return h
}

// checkOrigin memvalidasi Origin header terhadap allowlist.
func (h *Hub) checkOrigin(r *http.Request) bool {
	// Di development, izinkan semua origin
	if len(h.allowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // non-browser clients
	}
	return h.allowedOrigins[origin]
}

// authenticateToken memverifikasi JWT dari query param atau header.
func (h *Hub) authenticateToken(r *http.Request) (string, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		auth := r.Header.Get("Authorization")
		tokenStr = removeBearerPrefix(auth)
	}
	if tokenStr == "" {
		return "", jwt.ErrTokenMalformed
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid || claims.UserID == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return claims.UserID, nil
}

func removeBearerPrefix(s string) string {
	if len(s) > 7 && s[:7] == "Bearer " {
		return s[7:]
	}
	return s
}

// ServeWS menangani koneksi WebSocket dengan autentikasi.
// User hanya bisa subscribe ke event miliknya sendiri.
func (h *Hub) ServeWS(c *gin.Context) {
	// Autentikasi koneksi — jangan percaya userId dari URL.
	userID, err := h.authenticateToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: token tidak valid"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ws] upgrade gagal: %v", err)
		return
	}

	// Daftarkan koneksi
	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]bool)
	}
	h.clients[userID][conn] = true
	h.mu.Unlock()

	log.Printf("[ws] user %s terhubung", userID)

	// Set ping/pong deadline
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Ping routine — kirim ping secara berkala
	pingDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-pingDone:
				return
			}
		}
	}()

	// Read loop — baca pesan (atau detect koneksi mati)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] error user %s: %v", userID, err)
			}
			break
		}
	}

	// Cleanup
	close(pingDone)
	h.mu.Lock()
	delete(h.clients[userID], conn)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
	h.mu.Unlock()
	conn.Close()
	log.Printf("[ws] user %s terputus", userID)
}

// BroadcastToUser mengirim pesan JSON ke semua koneksi user.
func (h *Hub) BroadcastToUser(userId string, msg map[string]any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.clients[userId]
	for conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("[ws] kirim gagal ke %s: %v", userId, err)
			// Cleanup akan dilakukan oleh read loop
		}
	}
}

// CleanupStaleConnections menutup koneksi yang sudah mati.
// Bisa dipanggil secara periodik dari goroutine.
func (h *Hub) CleanupStaleConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for userID, conns := range h.clients {
		for conn := range conns {
			// Kirim ping untuk test koneksi
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				conn.Close()
				delete(conns, conn)
			}
		}
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
	}
}

// StartCleanupLoop menjalankan pembersihan berkala (panggil sebagai goroutine).
func (h *Hub) StartCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.CleanupStaleConnections()
		}
	}
}
