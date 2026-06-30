package repository

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/order/domain"
)

type OrderModel struct {
	ID             string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BuyerID        string  `gorm:"type:uuid;not null;index"`
	StoreID        string  `gorm:"type:uuid;not null;index"`
	RecipientName  string  `gorm:"not null"`
	Phone          string  `gorm:"not null"`
	FullAddress    string  `gorm:"not null"`
	DeliveryMethod string  `gorm:"not null"`
	Subtotal       int64   `gorm:"not null"`
	Discount       int64   `gorm:"not null;default:0"`
	DiscountCode   string  `gorm:"not null;default:''"`
	DeliveryFee    int64   `gorm:"not null"`
	Tax            int64   `gorm:"not null"`
	Total          int64   `gorm:"not null"`
	Status         string  `gorm:"not null"`
	DriverID       *string `gorm:"type:uuid"`
	DriverEarning  int64   `gorm:"not null;default:0"`
	TakenAt        *time.Time
	CompletedAt    *time.Time
	RefundedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Items   []OrderItemModel          `gorm:"foreignKey:OrderID"`
	History []OrderStatusHistoryModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }

type OrderItemModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID     string `gorm:"type:uuid;not null;index"`
	ProductID   string `gorm:"type:uuid;not null"`
	ProductName string `gorm:"not null"`
	Price       int64  `gorm:"not null"`
	Quantity    int    `gorm:"not null"`
	Subtotal    int64  `gorm:"not null"`
}

func (OrderItemModel) TableName() string { return "order_items" }

type OrderStatusHistoryModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID   string `gorm:"type:uuid;not null;index"`
	Status    string `gorm:"not null"`
	Note      string `gorm:"not null;default:''"`
	CreatedAt time.Time
}

func (OrderStatusHistoryModel) TableName() string { return "order_status_histories" }

func (m OrderModel) toDomain() *domain.Order {
	items := make([]domain.OrderItem, 0, len(m.Items))
	for _, it := range m.Items {
		items = append(items, domain.OrderItem{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		})
	}
	history := make([]domain.OrderStatus, 0, len(m.History))
	for _, h := range m.History {
		history = append(history, domain.OrderStatus{
			Status:    h.Status,
			Note:      h.Note,
			CreatedAt: h.CreatedAt,
		})
	}
	driverID := ""
	if m.DriverID != nil {
		driverID = *m.DriverID
	}
	return &domain.Order{
		ID:             m.ID,
		BuyerID:        m.BuyerID,
		StoreID:        m.StoreID,
		RecipientName:  m.RecipientName,
		Phone:          m.Phone,
		FullAddress:    m.FullAddress,
		DeliveryMethod: m.DeliveryMethod,
		Subtotal:       m.Subtotal,
		Discount:       m.Discount,
		DiscountCode:   m.DiscountCode,
		DeliveryFee:    m.DeliveryFee,
		Tax:            m.Tax,
		Total:          m.Total,
		Status:         m.Status,
		DriverID:       driverID,
		DriverEarning:  m.DriverEarning,
		Refunded:       m.RefundedAt != nil,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		Items:          items,
		StatusHistory:  history,
	}
}
