package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"

	"printmart/backend/internal/config"
	"printmart/backend/internal/handler"
	"printmart/backend/internal/repository"
	wshub "printmart/backend/pkg/websocket"
)

// buildPdf menghasilkan PDF minimal dengan pageCount halaman (port dari
// helper smoke-test.js) sehingga heuristik penghitung halaman mendeteksi
// jumlah yang tepat.
func buildPdf(pageCount int) []byte {
	objs := []string{"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n"}
	kids := []string{}
	for i := 0; i < pageCount; i++ {
		base := 3 + i*3
		kids = append(kids, fmt.Sprintf("%d 0 R", base))
		objs = append(objs, fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents %d 0 R >>\nendobj\n", base, base+1))
		objs = append(objs, fmt.Sprintf("%d 0 obj\n<< /Length 44 >>\nstream\nBT /F1 24 Tf 72 720 Td (Halaman %d) Tj ET\nendstream\nendobj\n", base+1, i+1))
	}
	pagesObj := fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), pageCount)
	objs = append([]string{pagesObj}, objs...)

	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	for _, o := range objs {
		sb.WriteString(o)
	}
	sb.WriteString("%%EOF")
	return []byte(sb.String())
}

// mget menelusuri map bersarang: mget(m, "order", "total_price").
func mget(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

func arr(v any) []any {
	a, _ := v.([]any)
	return a
}

func str(v any) string { return fmt.Sprintf("%v", v) }

func TestSmoke(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{
		Port:            0,
		AppEnv:          "test",
		DatabasePath:    filepath.Join(tmp, "smoke.db"),
		RedisAddr:       "",
		JWTSecret:       "smoke-test-secret",
		JWTExpiresIn:    24 * time.Hour,
		PaymentProvider: "midtrans-sandbox",
		UploadDir:       filepath.Join(tmp, "uploads"),
		MaxUploadSizeMB: 20,
		CorsOrigin:      "*",
		AdminEmail:      "admin@printmart.local",
		AdminPassword:   "admin123",
	}

	db, err := repository.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := repository.Seed(db, cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	hub := wshub.NewHub()
	r := handler.NewRouter(db, cfg, hub)
	ts := httptest.NewServer(r)
	defer ts.Close()
	client := ts.Client()

	passed := 0
	ok := func(label string) { passed++; t.Logf("  PASS %s", label) }

	api := func(method, path string, body any, token string) (int, map[string]any) {
		var rdr io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rdr = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, ts.URL+path, rdr)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		defer res.Body.Close()
		var j map[string]any
		json.NewDecoder(res.Body).Decode(&j)
		return res.StatusCode, j
	}

	var token, adminToken, userID string

	// ===== AUTH =====
	{
		st, j := api("POST", "/api/auth/register", map[string]any{"name": "Budi Santoso", "email": "budi@test.com", "password": "rahasia123", "phone": "0812-3456-7890"}, "")
		if st != 201 {
			t.Fatalf("register: status %d (%v)", st, j)
		}
		token = str(j["token"])
		userID = str(mget(j, "user", "id"))
		ok("register user baru")

		st, _ = api("POST", "/api/auth/register", map[string]any{"name": "X", "email": "budi@test.com", "password": "rahasia123"}, "")
		if st != 409 {
			t.Fatalf("register duplikat: status %d", st)
		}
		ok("register email duplikat ditolak")

		st, j = api("POST", "/api/auth/login", map[string]any{"email": "budi@test.com", "password": "rahasia123"}, "")
		if st != 200 || j["token"] == nil {
			t.Fatalf("login: status %d (%v)", st, j)
		}
		ok("login berhasil")

		st, j = api("POST", "/api/auth/login", map[string]any{"email": "admin@printmart.local", "password": "admin123"}, "")
		if st != 200 {
			t.Fatalf("login admin: status %d", st)
		}
		adminToken = str(j["token"])
		ok("login admin seed berhasil")
	}

	// ===== PRODUK & CATEGORY =====
	var productID string
	var productPrice float64
	{
		st, j := api("GET", "/api/products", nil, "")
		products := arr(mget(j, "products"))
		if st != 200 || len(products) < 10 {
			t.Fatalf("list produk: status %d, count %d", st, len(products))
		}
		p0 := products[0].(map[string]any)
		productID = str(p0["id"])
		productPrice = p0["price"].(float64)
		ok(fmt.Sprintf("list produk (%d item)", len(products)))

		_, j = api("GET", "/api/products?search=pulpen", nil, "")
		if len(arr(mget(j, "products"))) == 0 {
			t.Fatalf("search produk kosong")
		}
		ok("search produk")

		_, j = api("GET", "/api/categories", nil, "")
		if len(arr(mget(j, "categories"))) < 4 {
			t.Fatalf("kategori < 4")
		}
		ok("list kategori")

		st, j = api("GET", "/api/products/"+productID, nil, "")
		if st != 200 || str(mget(j, "product", "id")) != productID {
			t.Fatalf("detail produk: status %d", st)
		}
		ok("detail produk")
	}

	// ===== CART =====
	{
		_, j := api("GET", "/api/cart", nil, token)
		if len(arr(j["items"])) != 0 {
			t.Fatalf("cart awal tidak kosong")
		}
		st, j := api("POST", "/api/cart", map[string]any{"product_id": productID, "quantity": 2}, token)
		items := arr(j["items"])
		if st != 201 || len(items) != 1 || str(items[0].(map[string]any)["quantity"]) != "2" {
			t.Fatalf("tambah cart: status %d (%v)", st, j)
		}
		ok("tambah item ke cart")
	}

	// ===== ORDER PRODUK + PAYMENT =====
	{
		st, j := api("POST", "/api/orders", map[string]any{"orderType": "product", "deliveryMethod": "kirim", "delivery_address": "Jl. Merdeka No.1, Bandung 40111", "customer_name": "Budi"}, token)
		orderID := str(mget(j, "order", "id"))
		if st != 201 || mget(j, "order", "total_price").(float64) != productPrice*2 || str(mget(j, "order", "status")) != "menunggu_pembayaran" {
			t.Fatalf("checkout: status %d (%v)", st, j)
		}
		ok("checkout dari cart (stok berkurang, cart kosong)")

		_, j = api("GET", "/api/cart", nil, token)
		if len(arr(j["items"])) != 0 {
			t.Fatalf("cart tidak dibersihkan")
		}
		ok("cart dibersihkan")

		st, j = api("POST", "/api/payments/create", map[string]any{"orderId": orderID}, token)
		if st != 200 && st != 201 {
			t.Fatalf("payment create: status %d", st)
		}
		if str(mget(j, "payment", "va_number")) == "" {
			t.Fatalf("va_number kosong")
		}
		ok(fmt.Sprintf("payment dibuat (VA %s)", str(mget(j, "payment", "va_number"))))

		_, j = api("POST", "/api/payments/webhook", map[string]any{"orderId": orderID, "status": "success"}, token)
		if str(mget(j, "payment", "status")) != "success" || str(mget(j, "order", "status")) != "diproses" {
			t.Fatalf("webhook success: %v", j)
		}
		ok("webhook sukses -> order diproses")

		_, j = api("GET", "/api/orders/"+orderID, nil, token)
		if str(mget(j, "order", "status")) != "diproses" || len(arr(mget(j, "order", "items"))) != 1 {
			t.Fatalf("detail order: %v", j)
		}
		ok("detail order lengkap (items + payment)")
	}

	// ===== JASA PRINT (upload + hitung halaman + harga) =====
	var printJobID string
	var finishingPrice float64
	{
		// multipart upload
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile("file", "makalah-sistem.pdf")
		fw.Write(buildPdf(4))
		mw.Close()
		req, _ := http.NewRequest("POST", ts.URL+"/api/print-jobs/upload", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("upload: %v", err)
		}
		var up map[string]any
		json.NewDecoder(res.Body).Decode(&up)
		res.Body.Close()
		if res.StatusCode != 201 || str(mget(up, "upload", "pageCount")) != "4" {
			t.Fatalf("upload: status %d (%v)", res.StatusCode, up)
		}
		ok(fmt.Sprintf("upload PDF -> terdeteksi %s halaman", str(mget(up, "upload", "pageCount"))))

		_, opts := api("GET", "/api/print-jobs/options", nil, "")
		var finishingID string
		for _, f := range arr(mget(opts, "finishingOptions")) {
			fm := f.(map[string]any)
			if fm["name"] == "Jilid Spiral" {
				finishingID = str(fm["id"])
				finishingPrice = fm["price"].(float64)
			}
		}
		if finishingID == "" {
			t.Fatalf("finishing Jilid Spiral tidak ada")
		}
		ok("opsi print & finishing tersedia")

		st, j := api("POST", "/api/print-jobs", map[string]any{
			"fileUrl":           str(mget(up, "upload", "fileUrl")),
			"fileName":          str(mget(up, "upload", "fileName")),
			"pageCount":         mget(up, "upload", "pageCount"),
			"colorMode":         "bw",
			"paperSize":         "a4",
			"copies":            2,
			"duplex":            1,
			"finishingOptionId": finishingID,
		}, token)
		printJobID = str(mget(j, "printJob", "id"))
		expected := 500.0*4*2 + finishingPrice
		if st != 201 || mget(j, "pricing", "total").(float64) != expected {
			t.Fatalf("create print job: status %d, total %v want %v", st, mget(j, "pricing", "total"), expected)
		}
		ok(fmt.Sprintf("print job dibuat, harga: %s", str(mget(j, "pricing", "formatted"))))

		st, j = api("PATCH", "/api/print-jobs/"+printJobID, map[string]any{"colorMode": "color", "paperSize": "a3", "copies": 1}, token)
		if st != 200 || mget(j, "pricing", "total").(float64) != 1500.0*2*4*1+finishingPrice {
			t.Fatalf("re-spec: total %v", mget(j, "pricing", "total"))
		}
		ok("ubah spesifikasi -> harga baru")

		_, j = api("PATCH", "/api/print-jobs/"+printJobID, map[string]any{"colorMode": "bw", "paperSize": "a4", "copies": 2}, token)
		if mget(j, "pricing", "total").(float64) != expected {
			t.Fatalf("re-spec back: total %v", mget(j, "pricing", "total"))
		}
	}

	// ===== ORDER PRINT + STATUS REAL-TIME =====
	{
		st, j := api("POST", "/api/orders", map[string]any{"orderType": "print", "printJobIds": []string{printJobID}, "deliveryMethod": "ambil", "customer_name": "Budi"}, token)
		orderID := str(mget(j, "order", "id"))
		if st != 201 || mget(j, "order", "total_price").(float64) != 500.0*4*2+finishingPrice {
			t.Fatalf("order print: status %d, total %v", st, mget(j, "order", "total_price"))
		}
		ok("order jasa print dibuat")

		api("POST", "/api/payments/webhook", map[string]any{"orderId": orderID, "status": "success"}, token)

		// buka websocket lalu picu perubahan status
		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/orders/" + userID
		conn, _, err := gws.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("ws dial: %v", err)
		}
		defer conn.Close()

		got := make(chan map[string]any, 1)
		go func() {
			for {
				var m map[string]any
				if err := conn.ReadJSON(&m); err != nil {
					return
				}
				if m["type"] == "order_status" {
					got <- m
					return
				}
			}
		}()
		time.Sleep(100 * time.Millisecond)
		api("PATCH", "/api/print-jobs/"+printJobID+"/status", map[string]any{"status": "diproses"}, adminToken)

		select {
		case m := <-got:
			ok(fmt.Sprintf("WebSocket real-time: %q", str(m["message"])))
		case <-time.After(5 * time.Second):
			t.Fatalf("tidak ada pesan WebSocket order_status")
		}

		api("PATCH", "/api/print-jobs/"+printJobID+"/status", map[string]any{"status": "siap"}, adminToken)
		api("PATCH", "/api/print-jobs/"+printJobID+"/status", map[string]any{"status": "selesai"}, adminToken)
		_, j = api("GET", "/api/orders/"+orderID, nil, token)
		if str(mget(j, "order", "status")) != "selesai" {
			t.Fatalf("order auto-selesai gagal: %v", mget(j, "order", "status"))
		}
		ok("order auto-selesai saat semua print job selesai")
	}

	// ===== ADMIN =====
	{
		st, _ := api("GET", "/api/admin/stats", nil, "")
		if st != 401 {
			t.Fatalf("admin tanpa token: status %d", st)
		}
		ok("admin stats menolak tanpa token")

		st, _ = api("GET", "/api/admin/stats", nil, token)
		if st != 403 {
			t.Fatalf("admin sebagai user: status %d", st)
		}
		ok("admin stats menolak user biasa")

		st, j := api("GET", "/api/admin/stats", nil, adminToken)
		if st != 200 || mget(j, "stats", "totalRevenue").(float64) < 500*4*2+finishingPrice {
			t.Fatalf("admin stats: %v", j)
		}
		if mget(j, "stats", "userCount").(float64) != 1 {
			t.Fatalf("userCount != 1: %v", mget(j, "stats", "userCount"))
		}
		ok(fmt.Sprintf("dashboard stats: revenue %v", mget(j, "stats", "totalRevenue")))

		_, cats := api("GET", "/api/categories", nil, "")
		catID := str(arr(mget(cats, "categories"))[0].(map[string]any)["id"])
		st, cr := api("POST", "/api/products", map[string]any{"category_id": catID, "name": "Produk Test", "price": 1000, "stock": 5}, adminToken)
		if st != 201 {
			t.Fatalf("create product: status %d", st)
		}
		newProdID := str(mget(cr, "product", "id"))
		_, ur := api("PUT", "/api/products/"+newProdID, map[string]any{"price": 2000}, adminToken)
		if mget(ur, "product", "price").(float64) != 2000 {
			t.Fatalf("update price gagal: %v", mget(ur, "product", "price"))
		}
		st, _ = api("DELETE", "/api/products/"+newProdID, nil, adminToken)
		if st != 200 {
			t.Fatalf("delete product: status %d", st)
		}
		ok("CRUD produk admin lengkap")
	}

	// ===== VOUCHER / DISKON =====
	{
		// validate voucher persen terhadap subtotal
		st, j := api("POST", "/api/vouchers/validate", map[string]any{"code": "HEMAT10", "subtotal": 50000}, token)
		if st != 200 || mget(j, "discount").(float64) != 5000 || mget(j, "total").(float64) != 45000 {
			t.Fatalf("validate HEMAT10: status %d (%v)", st, j)
		}
		ok("validasi voucher persen (HEMAT10 -> diskon 5.000)")

		// min-spend ditolak
		st, _ = api("POST", "/api/vouchers/validate", map[string]any{"code": "POTONG5K", "subtotal": 10000}, token)
		if st != 400 {
			t.Fatalf("min-spend seharusnya ditolak, status %d", st)
		}
		ok("voucher min-spend ditolak saat subtotal kurang")

		// checkout produk dengan voucher -> total terpotong
		_, prods := api("GET", "/api/products", nil, "")
		var expensiveID string
		var expensivePrice float64
		for _, p := range arr(mget(prods, "products")) {
			pm := p.(map[string]any)
			if pm["price"].(float64) >= 40000 {
				expensiveID = str(pm["id"])
				expensivePrice = pm["price"].(float64)
				break
			}
		}
		if expensiveID == "" {
			t.Fatalf("tidak ada produk >= 40000 untuk uji voucher")
		}
		api("POST", "/api/cart", map[string]any{"product_id": expensiveID, "quantity": 1}, token)
		st, j = api("POST", "/api/orders", map[string]any{"orderType": "product", "deliveryMethod": "ambil", "voucher_code": "POTONG5K", "customer_name": "Budi"}, token)
		if st != 201 {
			t.Fatalf("checkout dgn voucher: status %d (%v)", st, j)
		}
		if mget(j, "order", "discount").(float64) != 5000 {
			t.Fatalf("discount tersimpan salah: %v", mget(j, "order", "discount"))
		}
		if mget(j, "order", "subtotal").(float64) != expensivePrice {
			t.Fatalf("subtotal salah: %v want %v", mget(j, "order", "subtotal"), expensivePrice)
		}
		if mget(j, "order", "total_price").(float64) != expensivePrice-5000 {
			t.Fatalf("total setelah diskon salah: %v", mget(j, "order", "total_price"))
		}
		if str(mget(j, "order", "voucher_code")) != "POTONG5K" {
			t.Fatalf("voucher_code tidak tersimpan: %v", mget(j, "order", "voucher_code"))
		}
		ok(fmt.Sprintf("checkout pakai voucher POTONG5K (total %v)", mget(j, "order", "total_price")))

		// admin CRUD voucher
		st, cr := api("POST", "/api/admin/vouchers", map[string]any{"code": "smoke50", "discount_type": "percent", "discount_value": 50, "description": "uji"}, adminToken)
		if st != 201 || str(mget(cr, "voucher", "code")) != "SMOKE50" {
			t.Fatalf("admin create voucher: status %d (%v)", st, cr)
		}
		vID := str(mget(cr, "voucher", "id"))
		_, lv := api("GET", "/api/admin/vouchers", nil, adminToken)
		if len(arr(mget(lv, "vouchers"))) < 4 {
			t.Fatalf("admin list voucher < 4")
		}
		st, _ = api("DELETE", "/api/admin/vouchers/"+vID, nil, adminToken)
		if st != 200 {
			t.Fatalf("admin delete voucher: status %d", st)
		}
		ok("admin CRUD voucher lengkap")
	}

	// ===== NOTIFIKASI =====
	{
		_, j := api("GET", "/api/notifications", nil, token)
		if len(arr(mget(j, "notifications"))) < 3 {
			t.Fatalf("notifikasi < 3: %d", len(arr(mget(j, "notifications"))))
		}
		ok(fmt.Sprintf("notifikasi tersimpan (%d)", len(arr(mget(j, "notifications")))))
	}

	t.Logf("SMOKE TEST SELESAI - %d pengujian LULUS", passed)
}

