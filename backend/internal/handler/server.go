// Package handler — HTTP handlers (controller) PrintKu API.
package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/config"
	"printmart/backend/internal/middleware"
	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
	"printmart/backend/pkg/websocket"
)

type Server struct {
	DB  *sql.DB
	Cfg *config.Config
	Hub *websocket.Hub
}

func (s *Server) broadcast(userID string, msg map[string]any) {
	s.Hub.BroadcastToUser(userID, msg)
}

func (s *Server) user(c *gin.Context) middleware.AuthUser {
	u, _ := c.Get("user")
	au, ok := u.(middleware.AuthUser)
	if !ok {
		return middleware.AuthUser{}
	}
	return au
}

func (s *Server) setOrderStatus(orderID, status string) {
	service.SetOrderStatus(s.DB, orderID, status, s.broadcast)
}

// --- helper transaksi untuk handler ---

const txKey = "printku_tx"

func (s *Server) setTx(c *gin.Context, tx *sql.Tx) {
	c.Set(txKey, tx)
}

func (s *Server) tx(c *gin.Context) *sql.Tx {
	if v, ok := c.Get(txKey); ok {
		if t, ok := v.(*sql.Tx); ok {
			return t
		}
	}
	return nil
}

func (s *Server) q(c *gin.Context, query string, args ...any) ([]map[string]any, error) {
	if t := s.tx(c); t != nil {
		return allRowsTx(t, query, args...)
	}
	return allRows(s.DB, query, args...)
}

func (s *Server) q1(c *gin.Context, query string, args ...any) (map[string]any, error) {
	if t := s.tx(c); t != nil {
		return getRowTx(t, query, args...)
	}
	return getRow(s.DB, query, args...)
}

func (s *Server) now() string { return repository.Now() }

func (s *Server) jsonErr(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func jsonOK(c *gin.Context, data gin.H)     { c.JSON(http.StatusOK, data) }
func jsonCreated(c *gin.Context, data gin.H) { c.JSON(http.StatusCreated, data) }

func jsonBad(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

func jsonNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

func jsonForbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
}