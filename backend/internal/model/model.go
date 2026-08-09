package model

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Role         string `json:"role"`
	CreatedAt    string `json:"created_at"`
	PasswordHash string `json:"-"`
}

type PublicUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type Address struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Label       string `json:"label"`
	FullAddress string `json:"full_address"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
	CreatedAt   string `json:"created_at"`
}

type Product struct {
	ID          string `json:"id"`
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64   `json:"price"`
	Stock       int64   `json:"stock"`
	ImageURL    string `json:"image_url"`
	IsActive    int64   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type ProductWithCategory struct {
	Product
	CategoryName string `json:"category_name"`
}

type CartItem struct {
	CartItemID string `json:"cart_item_id"`
	Quantity   int64  `json:"quantity"`
	Product
}

type CartResponse struct {
	Items []CartItem `json:"items"`
	Total int64      `json:"total"`
}

type OrderItem struct {
	ID          string `json:"id"`
	OrderID     string `json:"order_id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int64  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Subtotal    int64  `json:"subtotal"`
}

type PrintJob struct {
	ID                string  `json:"id"`
	OrderID           *string `json:"order_id"`
	FileURL           string  `json:"file_url"`
	FileName          string  `json:"file_name"`
	PageCount         int64   `json:"page_count"`
	ColorMode         string  `json:"color_mode"`
	PaperSize         string  `json:"paper_size"`
	Copies            int64   `json:"copies"`
	Duplex            int64   `json:"duplex"`
	FinishingOptionID *string `json:"finishing_option_id"`
	EstimatedPrice    int64   `json:"estimated_price"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
}

type PrintJobDetail struct {
	PrintJob
	Finishing      *Finishing `json:"finishing"`
	FinishingName  string     `json:"finishing_name,omitempty"`
	FinishingPrice int64      `json:"finishing_price,omitempty"`
	TotalPages     int64      `json:"totalPages"`
}

type Payment struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	Provider      string  `json:"provider"`
	TransactionID *string `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        int64   `json:"amount"`
	VANumber      *string `json:"va_number"`
	PaidAt        *string `json:"paid_at"`
	CreatedAt     string  `json:"created_at"`
}

type Order struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	OrderType        string `json:"order_type"`
	Status           string `json:"status"`
	TotalPrice       int64  `json:"total_price"`
	DeliveryMethod   string `json:"delivery_method"`
	DeliveryAddress  string `json:"delivery_address"`
	Notes            string `json:"notes"`
	CustomerName     string `json:"customer_name"`
	CreatedAt        string `json:"created_at"`
	UserName         string `json:"user_name,omitempty"`
	PaymentStatus    string `json:"payment_status,omitempty"`
	Items            []OrderItem    `json:"items,omitempty"`
	PrintJobs        []PrintJobDetail `json:"printJobs,omitempty"`
	Payment          *Payment       `json:"payment,omitempty"`
}

type Finishing struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Price     int64  `json:"price"`
	CreatedAt string `json:"created_at"`
}

type Category struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	OrderID   *string `json:"order_id"`
	Message   string `json:"message"`
	IsRead    int64  `json:"is_read"`
	CreatedAt string `json:"created_at"`
}