package restapi

type NewSmscRequest struct {
	URL         string
	System      System
	Enquiry     *Enquiry
	Transaction *Transaction
}

type System struct {
	Id       string
	Password string
	Type     string
}

type Enquiry struct {
	Link        int
	LinkTimeout int
}

type Transaction struct {
	Timeout int
}
