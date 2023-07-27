package proto

import _ "embed"

var (
	//go:embed shop/v1/address.proto
	Address string
	//go:embed shop/v1/customer_v1.proto
	CustomerV1 string
	//go:embed shop/v1/customer.proto
	CustomerV2 string
	//go:embed shop/v1/order.proto
	Order string
)
