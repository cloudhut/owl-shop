package proto

import _ "embed"

var (
	//go:embed shop/v1/address.proto
	Address string
	//go:embed shop/v1/customer.proto
	Customer string
	//go:embed shop/v1/order.proto
	Order string
)
