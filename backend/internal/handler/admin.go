package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) adminStats(c *gin.Context) {
	stats := gin.H{
		"totalRevenue": 0,
		"orderCounts":  gin.H{"total": 0, "pending": 0, "processing": 0},
		"productCount": 0,
		"printJobs":    gin.H{"total": 0, "menunggu": 0, "diproses": 0},
		"userCount":    0,
	}

	var revenue int64
	s.DB.QueryRow(`SELECT COALESCE(SUM(o.total_price), 0) FROM orders o JOIN payments p ON p.order_id = o.id WHERE p.status = 'success'`).Scan(&revenue)
	stats["totalRevenue"] = revenue

	var total, pending, processing int64
	s.DB.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'menunggu_pembayaran' THEN 1 ELSE 0 END), 0), COALESCE(SUM(CASE WHEN status = 'diproses' THEN 1 ELSE 0 END), 0) FROM orders`).
		Scan(&total, &pending, &processing)
	stats["orderCounts"] = gin.H{"total": total, "pending": pending, "processing": processing}

	var productCount int64
	s.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE is_active = 1`).Scan(&productCount)
	stats["productCount"] = productCount

	var pjTotal, pjWaiting, pjProcessing int64
	s.DB.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'menunggu' THEN 1 ELSE 0 END), 0), COALESCE(SUM(CASE WHEN status = 'diproses' THEN 1 ELSE 0 END), 0) FROM print_jobs`).
		Scan(&pjTotal, &pjWaiting, &pjProcessing)
	stats["printJobs"] = gin.H{"total": pjTotal, "menunggu": pjWaiting, "diproses": pjProcessing}

	var userCount int64
	s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'user'`).Scan(&userCount)
	stats["userCount"] = userCount

	recentOrders, _ := allRows(s.DB, `SELECT o.*, u.name AS user_name FROM orders o
		JOIN users u ON u.id = o.user_id ORDER BY o.created_at DESC LIMIT 8`)
	revenueByDay, _ := allRows(s.DB, `SELECT date(o.created_at) AS day, SUM(o.total_price) AS total
		FROM orders o JOIN payments p ON p.order_id = o.id
		WHERE p.status = 'success' GROUP BY date(o.created_at) ORDER BY day DESC LIMIT 7`)

	c.JSON(http.StatusOK, gin.H{
		"stats": stats, "revenueByDay": revenueByDay, "recentOrders": recentOrders,
	})
}

func (s *Server) adminOrders(c *gin.Context) {
	status := c.Query("status")
	var rows []map[string]any
	var err error
	if status != "" {
		rows, err = allRows(s.DB, `SELECT o.*, u.name AS user_name,
			(SELECT p.status FROM payments p WHERE p.order_id = o.id ORDER BY p.created_at DESC LIMIT 1) AS payment_status
			FROM orders o JOIN users u ON u.id = o.user_id
			WHERE o.status = ? ORDER BY o.created_at DESC`, status)
	} else {
		rows, err = allRows(s.DB, `SELECT o.*, u.name AS user_name,
			(SELECT p.status FROM payments p WHERE p.order_id = o.id ORDER BY p.created_at DESC LIMIT 1) AS payment_status
			FROM orders o JOIN users u ON u.id = o.user_id
			ORDER BY o.created_at DESC`)
	}
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": rows})
}

func (s *Server) adminPrintJobs(c *gin.Context) {
	rows, err := allRows(s.DB, `SELECT pj.*, o.user_id, o.total_price FROM print_jobs pj
		LEFT JOIN orders o ON o.id = pj.order_id ORDER BY pj.created_at DESC`)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"printJobs": rows})
}