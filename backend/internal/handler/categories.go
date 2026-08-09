package handler

import "github.com/gin-gonic/gin"

func (s *Server) listCategories(c *gin.Context) {
	query := `SELECT * FROM categories ORDER BY name`
	var rows []map[string]any
	var err error
	if t := c.Query("type"); t != "" {
		rows, err = allRows(s.DB, query+" WHERE type = ?", t)
	} else {
		rows, err = allRows(s.DB, query)
	}
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(200, gin.H{"categories": rows})
}