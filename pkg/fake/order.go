package fake

import (
	"github.com/brianvoe/gofakeit/v5"
	"github.com/cloudhut/owl-shop/pkg/protobuf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func NewOrder(customer Customer) Order {
	return Order{
		Version:       0,
		ID:            gofakeit.UUID(),
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		DeliveredAt:   nil,
		CompletedAt:   nil,
		Customer:      customer,
		OrderValue:    gofakeit.Number(5000, 250000),
		LineItems:     newOrderLineItems(),
		Payment: OrderPayment{
			PaymentID: gofakeit.UUID(),
			Method:    gofakeit.RandomString([]string{"CASH", "DEBIT", "CREDIT_CARD", "PAYPAL"}),
		},
		DeliveryAddress: NewAddress(customer),
		Revision:        0,
	}
}

type Order struct {
	// VersionedStruct
	Version int `json:"version"`

	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastUpdatedAt time.Time  `json:"lastUpdatedAt"`
	DeliveredAt   *time.Time `json:"deliveredAt"`
	CompletedAt   *time.Time `json:"completedAt"`

	Customer        Customer        `json:"customer"`
	OrderValue      int             `json:"orderValue"`
	LineItems       []OrderLineItem `json:"lineItems"`
	Payment         OrderPayment    `json:"payment"`
	DeliveryAddress Address         `json:"deliveryAddress"`
	Revision        int             `json:"revision"`
}

func (o *Order) Protobuf() *protobuf.Order {
	lineItems := make([]*protobuf.Order_LineItem, len(o.LineItems))
	for i, item := range o.LineItems {
		lineItems[i] = item.Protobuf()
	}

	order := protobuf.Order{
		Version:         int32(o.Version),
		Id:              o.ID,
		CreatedAt:       timestamppb.New(o.CreatedAt),
		LastUpdatedAt:   timestamppb.New(o.LastUpdatedAt),
		DeliveredAt:     protobuf.NewTimestamp(o.DeliveredAt),
		CompletedAt:     protobuf.NewTimestamp(o.CompletedAt),
		Customer:        o.Customer.Protobuf(),
		OrderValue:      int32(o.OrderValue),
		LineItems:       lineItems,
		Payment:         o.Payment.Protobuf(),
		DeliveryAddress: o.DeliveryAddress.Protobuf(),
		Revision:        int32(o.Revision),
	}

	return &order
}

func newOrderLineItems() []OrderLineItem {
	itemCount := gofakeit.Number(8, 45)
	items := make([]OrderLineItem, itemCount)
	for i := 0; i < itemCount; i++ {
		items[i] = newOrderLineItem()
	}

	return items
}

func newOrderLineItem() OrderLineItem {
	quantity := gofakeit.Number(1, 500)
	unitPrice := gofakeit.Number(1, 1000)
	return OrderLineItem{
		ArticleID:    gofakeit.UUID(),
		Name:         gofakeit.Vegetable(),
		Quantity:     quantity,
		QuantityUnit: gofakeit.RandomString([]string{"pieces", "gram"}),
		UnitPrice:    unitPrice,
		TotalPrice:   quantity * unitPrice,
	}
}

type OrderLineItem struct {
	ArticleID    string `json:"articleId"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	QuantityUnit string `json:"quantityUnit"`
	UnitPrice    int    `json:"unitPrice"`
	TotalPrice   int    `json:"totalPrice"`
}

func (o *OrderLineItem) Protobuf() *protobuf.Order_LineItem {
	return &protobuf.Order_LineItem{
		ArticleId:    o.ArticleID,
		Name:         o.Name,
		Quantity:     int32(o.Quantity),
		QuantityUnit: o.QuantityUnit,
		UnitPrice:    int32(o.UnitPrice),
		TotalPrice:   int32(o.TotalPrice),
	}
}

type OrderPayment struct {
	PaymentID string `json:"paymentId"`
	Method    string `json:"method"` // PAYPAL | CREDIT_CARD | DEBIT | CASH
}

func (o *OrderPayment) Protobuf() *protobuf.Order_Payment {
	return &protobuf.Order_Payment{
		PaymentId: o.PaymentID,
		Method:    o.Method,
	}
}

type OrderDeliveryAddress struct{}
