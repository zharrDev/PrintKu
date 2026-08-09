// Package websocket — real-time hub untuk status pesanan.
package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool
	upgrader websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		clients: map[string]map[*websocket.Conn]bool{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeWS menangani koneksi /ws/orders/:userId atau /ws?userId=...
func (h *Hub) ServeWS(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		userId = c.Query("userId")
	}
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId wajib diisi"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	if h.clients[userId] == nil {
		h.clients[userId] = map[*websocket.Conn]bool{}
	}
	h.clients[userId][conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients[userId], conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

// BroadcastToUser mengirim pesan JSON ke semua koneksi user.
func (h *Hub) BroadcastToUser(userId string, msg map[string]any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.clients[userId]
	for conn := range conns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("[ws] kirim gagal ke %s: %v", userId, err)
		}
	}
}