package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) listNotifications(c *gin.Context) {
	u := s.user(c)
	rows, err := allRows(s.DB, `SELECT * FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT 50`, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": rows})
}

func (s *Server) markNotificationRead(c *gin.Context) {
	u := s.user(c)
	s.DB.Exec(`UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`, c.Param("id"), u.ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) markAllNotificationsRead(c *gin.Context) {
	u := s.user(c)
	s.DB.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ?`, u.ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}