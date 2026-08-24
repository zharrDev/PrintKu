package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) listCategories(c *gin.Context) {
	t := c.Query("type")
	var rows []map[string]any
	var err error

	if t != "" {
		// WHERE sebelum ORDER BY
		rows, err = allRows(s.DB, `SELECT * FROM categories WHERE type = ? ORDER BY name`, t)
	} else {
		rows, err = allRows(s.DB, `SELECT * FROM categories ORDER BY name`)
	}
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": rows})
}
