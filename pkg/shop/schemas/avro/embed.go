package avro

import _ "embed"

var (
	//go:embed customer_v1.avsc
	CustomerV1Avro string
	//go:embed customer_v2.avsc
	CustomerV2Avro string
	//go:embed address.avsc
	AddressAvro string
	//go:embed order.avsc
	OrderAvro string
)
