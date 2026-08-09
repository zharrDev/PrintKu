package service

import (
	"database/sql"
	"log"
	"strings"

	"printmart/backend/internal/model"
	"printmart/backend/internal/repository"
)

// SetOrderStatus mengubah status order, mencatat notifikasi, dan broadcast
// via WebSocket hub (disuntikkan lewat callback).
func SetOrderStatus(db *sql.DB, orderID, status string, broadcast func(userID string, msg map[string]any)) *model.Order {
	var order model.Order
	err := db.QueryRow(`SELECT id, user_id, order_type, status, total_price, delivery_method, delivery_address, notes, customer_name, created_at
		FROM orders WHERE id = ?`, orderID).Scan(
		&order.ID, &order.UserID, &order.OrderType, &order.Status, &order.TotalPrice,
		&order.DeliveryMethod, &order.DeliveryAddress, &order.Notes, &order.CustomerName, &order.CreatedAt)
	if err != nil {
		log.Printf("[orderstatus] order tidak ditemukan: %v", err)
		return nil
	}
	if _, err := db.Exec(`UPDATE orders SET status = ? WHERE id = ?`, status, orderID); err != nil {
		log.Printf("[orderstatus] update gagal: %v", err)
		return nil
	}
	repository.Notify(db, order.UserID, orderID, "Pesanan #"+shortOrder(orderID)+" berstatus \""+status+"\"")
	if broadcast != nil {
		broadcast(order.UserID, map[string]any{"type": "order_status", "orderId": orderID, "status": status})
	}
	order.Status = status
	return &order
}

// SyncOrderStatusFromJobs mengetatkan status order berdasarkan status
// seluruh print job-nya: semua selesai -> order selesai; semua siap/selesai
// dan tidak ada yang menunggu/diproses -> order siap.
func SyncOrderStatusFromJobs(db *sql.DB, orderID string, broadcast func(userID string, msg map[string]any)) {
	rows, err := db.Query(`SELECT status FROM print_jobs WHERE order_id = ?`, orderID)
	if err != nil {
		return
	}
	defer rows.Close()
	statuses := map[string]bool{}
	for rows.Next() {
		var s string
		if rows.Scan(&s) == nil {
			statuses[s] = true
		}
	}
	if len(statuses) == 0 {
		return
	}
	if len(statuses) == 1 && statuses["selesai"] {
		SetOrderStatus(db, orderID, "selesai", broadcast)
		return
	}
	allReady := true
	for s := range statuses {
		if s != "siap" && s != "selesai" {
			allReady = false
			break
		}
	}
	if allReady {
		SetOrderStatus(db, orderID, "siap", broadcast)
	}
}

func shortOrder(id string) string {
	if len(id) > 8 {
		return strings.ToUpper(id[:8])
	}
	return strings.ToUpper(id)
}