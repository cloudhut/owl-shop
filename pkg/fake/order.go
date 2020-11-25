package fake

import (
	"github.com/brianvoe/gofakeit/v5"
	"time"
)

func NewOrder() *Order {
	var o Order
	gofakeit.Struct(&o)
	return &o
}

type Order struct {
	// VersionedStruct
	Version int

	ID            string
	CreatedAt     time.Time
	LastUpdatedAt time.Time
	DeliveredAt   time.Time
	CompletedAt   time.Time

	Customer        OrderCustomer
	OrderValue      int
	LineItems       []OrderLineItem
	Payment         OrderPayment
	DeliveryAddress OrderDeliveryAddress
}

type OrderLineItem struct {
	ArticleID    string `fake:"{uuid}"`
	Name         string `fake:"{number:1,15}"`
	Quantity     int    `fake:"{number:1,15}"`
	QuantityUnit string
	UnitPrice    int
	TotalPrice   int
}

type OrderPayment struct {
	PaymentID string
	Method    string // PAYPAL | CREDIT_CARD | DEBIT | CASH
}

type OrderDeliveryAddress struct{}
