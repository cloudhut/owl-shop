package fake

type OrderCustomer struct {
	// VersionedStruct
	Version int

	ID           string `fake:"{uuid}"`
	FirstName    string `fake:"{firstname}"`
	LastName     string `fake:"{lastname}"`
	CompanyName  string `fake:"{company}"`
	Email        string `fake:"{email}"`
	CustomerType string // PERSONAL | BUSINESS | NON_PROFIT
}
