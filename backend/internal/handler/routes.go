package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/config"
	"printmart/backend/internal/middleware"
	"printmart/backend/pkg/websocket"
)

// NewServer menyiapkan handler server dengan DB, config, dan hub.
func NewServer(db *sql.DB, cfg *config.Config, hub *websocket.Hub) *Server {
	return &Server{DB: db, Cfg: cfg, Hub: hub}
}

// RegisterRoutes memasang seluruh route API pada router Gin.
func RegisterRoutes(r *gin.Engine, s *Server) {
	admin := middleware.AdminRequired
	auth := middleware.AuthRequired(s.Cfg)
	optional := middleware.OptionalAuth(s.Cfg)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": "printku-api", "time": s.now()})
	})

	api := r.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.POST("/register", s.register)
	authGroup.POST("/login", s.login)
	authGroup.GET("/me", auth, s.me)
	authGroup.GET("/addresses", auth, s.listAddresses)
	authGroup.POST("/addresses", auth, s.createAddress)
	authGroup.DELETE("/addresses/:id", auth, s.deleteAddress)

	api.GET("/categories", s.listCategories)

	products := api.Group("/products", optional)
	products.GET("", s.listProducts)
	products.GET("/:id", s.getProduct)
	products.POST("", auth, admin, s.createProduct)
	products.PUT("/:id", auth, admin, s.updateProduct)
	products.DELETE("/:id", auth, admin, s.deleteProduct)

	// File download — membutuhkan authorization.
	// User hanya dapat mengakses file miliknya atau file order miliknya.
	files := api.Group("/files", auth)
	files.GET("/download/:filename", s.downloadFile)

	cart := api.Group("/cart", auth)
	cart.GET("", s.getCart)
	cart.POST("", s.addToCart)
	cart.PATCH("/:cartItemId", s.patchCartItem)
	cart.DELETE("/:cartItemId", s.deleteCartItem)
	cart.DELETE("", s.clearCart)

	orders := api.Group("/orders", auth)
	orders.POST("", s.createOrder)
	orders.GET("", s.listOrders)
	orders.GET("/:id", s.getOrder)
	orders.PATCH("/:id/status", admin, s.updateOrderStatus)

	jobs := api.Group("/print-jobs")
	jobs.GET("/options", s.printOptions)
	jobs.POST("/upload", auth, s.uploadFile)
	jobs.POST("", auth, s.createPrintJob)
	jobs.GET("", auth, s.listMyPrintJobs)
	jobs.GET("/:id", auth, s.getPrintJob)
	jobs.PATCH("/:id", auth, s.patchPrintJob)
	jobs.PATCH("/:id/status", auth, admin, s.patchPrintJobStatus)

	// Payment — sandbox create membutuhkan auth.
	payments := api.Group("/payments", auth)
	payments.POST("/create", s.createPayment)
	payments.GET("/:orderId", s.getPayment)

	// Payment webhook — TIDAK membutuhkan JWT user.
	// Webhook dari payment gateway harus public, tapi nanti
	// akan diverifikasi signature-nya (Phase 1).
	api.POST("/payments/webhook", s.paymentWebhook)

	notifications := api.Group("/notifications", auth)
	notifications.GET("", s.listNotifications)
	notifications.POST("/read-all", s.markAllNotificationsRead)
	notifications.POST("/:id/read", s.markNotificationRead)

	vouchers := api.Group("/vouchers", auth)
	vouchers.POST("/validate", s.validateVoucher)

	adm := api.Group("/admin", auth, admin)
	adm.GET("/stats", s.adminStats)
	adm.GET("/orders", s.adminOrders)
	adm.GET("/print-jobs", s.adminPrintJobs)
	adm.GET("/vouchers", s.adminListVouchers)
	adm.POST("/vouchers", s.adminCreateVoucher)
	adm.PATCH("/vouchers/:id", s.adminUpdateVoucher)
	adm.DELETE("/vouchers/:id", s.adminDeleteVoucher)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint tidak ditemukan"})
	})
}
