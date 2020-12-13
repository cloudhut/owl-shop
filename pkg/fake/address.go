package fake

import (
	"github.com/brianvoe/gofakeit/v5"
	"github.com/cloudhut/owl-shop/pkg/protobuf"
	"github.com/mroth/weightedrand"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math/rand"
	"strconv"
	"time"
)

type AddressType string

const (
	AddressTypeInvoice  AddressType = "INVOICE"
	AddressTypeDelivery AddressType = "DELIVERY"
)

func NewAddress(customer Customer) Address {
	address := gofakeit.Address()

	return Address{
		Version: 0,
		ID:      gofakeit.UUID(),
		Customer: AddressCustomer{
			CustomerID:   customer.ID,
			CustomerType: customer.CustomerType,
		},

		// Address info
		Type:                  newAddressType(),
		FirstName:             customer.FirstName,
		LastName:              customer.LastName,
		State:                 address.State,
		Street:                address.Street,
		HouseNumber:           strconv.Itoa(gofakeit.Number(1, 1000)),
		City:                  address.City,
		Zip:                   address.Zip,
		Latitude:              address.Latitude,
		Longitude:             address.Longitude,
		Phone:                 gofakeit.PhoneFormatted(),
		AdditionalAddressInfo: newAdditionalAddressInfo(),
		CreatedAt:             time.Now(),
		Revision:              0,
	}
}

type Address struct {
	// VersionedStruct
	Version int    `json:"version"`
	ID      string `json:"id"`

	Customer AddressCustomer `json:"customer"`

	// Address info
	Type                  AddressType `json:"type"`
	FirstName             string      `json:"firstName"`
	LastName              string      `json:"lastName"`
	State                 string      `json:"state"`
	Street                string      `json:"street"`
	HouseNumber           string      `json:"houseNumber"`
	City                  string      `json:"city"`
	Zip                   string      `json:"zip"`
	Latitude              float64     `json:"latitude"`
	Longitude             float64     `json:"longitude"`
	Phone                 string      `json:"phone"`
	AdditionalAddressInfo string      `json:"additionalAddressInfo"`
	CreatedAt             time.Time   `json:"createdAt"`
	Revision              int         `json:"revision"` // Each change on the customer increments the revision
}

func (a *Address) Protobuf() *protobuf.Address {
	return &protobuf.Address{
		Version:               int32(a.Version),
		Id:                    a.ID,
		Customer:              a.Customer.Protobuf(),
		Type:                  string(a.Type),
		FirstName:             a.FirstName,
		LastName:              a.LastName,
		State:                 a.State,
		HouseNumber:           a.HouseNumber,
		City:                  a.City,
		Zip:                   a.Zip,
		Latitude:              float32(a.Latitude),
		Longitude:             float32(a.Longitude),
		Phone:                 a.Phone,
		AdditionalAddressInfo: a.AdditionalAddressInfo,
		CreatedAt:             timestamppb.New(a.CreatedAt),
		Revision:              int32(a.Revision),
	}
}

type AddressCustomer struct {
	CustomerID   string       `json:"id"`
	CustomerType CustomerType `json:"type"`
}

func (a *AddressCustomer) Protobuf() *protobuf.Address_Customer {
	return &protobuf.Address_Customer{
		CustomerId:   a.CustomerID,
		CustomerType: string(a.CustomerType),
	}
}

// newAddressType returns an address type based on a weighted random choice
func newAddressType() AddressType {
	rand.Seed(time.Now().UTC().UnixNano())
	c, err := weightedrand.NewChooser(
		weightedrand.Choice{Item: AddressTypeDelivery, Weight: 20},
		weightedrand.Choice{Item: AddressTypeInvoice, Weight: 80},
	)
	if err != nil {
		panic(err)
	}
	addressType := c.Pick().(AddressType)
	return addressType
}

func newAdditionalAddressInfo() string {
	rand.Seed(time.Now().UTC().UnixNano())
	c, err := weightedrand.NewChooser(
		weightedrand.Choice{Item: "", Weight: 200},

		// 100 Sum
		weightedrand.Choice{Item: gofakeit.RandomString([]string{"a", "b", "c"}), Weight: 60},
		weightedrand.Choice{Item: gofakeit.HipsterWord(), Weight: 15},
		weightedrand.Choice{Item: gofakeit.HipsterSentence(4), Weight: 10},
		weightedrand.Choice{Item: gofakeit.Noun(), Weight: 14},
		weightedrand.Choice{Item: gofakeit.Emoji(), Weight: 1},
	)
	if err != nil {
		panic(err)
	}
	addressInfo := c.Pick().(string)
	return addressInfo
}
