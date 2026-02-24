package order

// apiCoffee mirrors the coffee object returned by the HashiCups API.
type apiCoffee struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Teaser      string  `json:"teaser"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

// apiOrderItem mirrors the order item object returned by the HashiCups API.
type apiOrderItem struct {
	Coffee   apiCoffee `json:"coffee"`
	Quantity int       `json:"quantity"`
}

// apiOrder mirrors the order object returned by the HashiCups API.
type apiOrder struct {
	ID    int            `json:"id"`
	Items []apiOrderItem `json:"items"`
}

// apiOrderItemPayload is the payload shape expected by the HashiCups orders endpoint.
type apiOrderItemPayload struct {
	Coffee   apiOrderItemCoffeePayload `json:"coffee"`
	Quantity int                       `json:"quantity"`
}

type apiOrderItemCoffeePayload struct {
	ID int `json:"id"`
}
