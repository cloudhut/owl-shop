package fake

import (
	"github.com/brianvoe/gofakeit/v5"
	"github.com/mroth/weightedrand"

	shoppb "github.com/cloudhut/owl-shop/pkg/protogen/shop/v1"
)

type CustomerType string

const (
	CustomerTypePersonal CustomerType = "PERSONAL"
	CustomerTypeBusiness CustomerType = "BUSINESS"
)

type Customer struct {
	// VersionedStruct
	Version int `json:"version"`

	ID           string       `json:"id"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	Gender       string       `json:"gender"`
	CompanyName  *string      `json:"companyName"`
	Email        string       `json:"email"`
	CustomerType CustomerType `json:"customerType"` // PERSONAL | BUSINESS
	Revision     int          `json:"revision"`     // Each change on the customer increments the revision
}

func (c *Customer) Protobuf() *shoppb.Customer {
	companyName := ""
	if c.CompanyName != nil {
		companyName = *c.CompanyName
	}

	customerType := shoppb.Customer_CUSTOMER_TYPE_PERSONAL
	if c.CustomerType == CustomerTypeBusiness {
		customerType = shoppb.Customer_CUSTOMER_TYPE_BUSINESS
	}

	return &shoppb.Customer{
		Version:      int32(c.Version),
		Id:           c.ID,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		Gender:       c.Gender,
		CompanyName:  companyName,
		Email:        c.Email,
		CustomerType: customerType,
		Revision:     int32(c.Revision),
	}
}

func NewCustomer() Customer {
	person := gofakeit.Person()

	var companyName *string
	customerType := newCustomerType()
	if customerType == CustomerTypeBusiness {
		company := gofakeit.Company()
		companyName = &company
	}

	return Customer{
		Version:      0,
		ID:           gofakeit.UUID(),
		FirstName:    person.FirstName,
		LastName:     person.LastName,
		Gender:       person.Gender,
		CompanyName:  companyName,
		Email:        gofakeit.Email(),
		CustomerType: customerType,
	}
}

// newCustomerType returns a customer type based on a weighted random choice
func newCustomerType() CustomerType {
	c, err := weightedrand.NewChooser(
		weightedrand.Choice{Item: CustomerTypePersonal, Weight: 99},
		weightedrand.Choice{Item: CustomerTypeBusiness, Weight: 1},
	)
	if err != nil {
		panic(err)
	}
	customerType := c.Pick().(CustomerType)
	return customerType
}
