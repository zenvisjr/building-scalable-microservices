package order

type Order struct {
	ID         string  `json:"id"`
	Created_at string  `json:"created_at"`
	Account_id string  `json:"account_id"`
	Price      float64 `json:"price"`
}

