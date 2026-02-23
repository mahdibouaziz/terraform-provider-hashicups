package client

import (
	"context"
	"fmt"
	"net/http"
)

// Coffee represents a coffee product.
type Coffee struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Teaser      string       `json:"teaser"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Image       string       `json:"image"`
	Ingredient  []Ingredient `json:"ingredients"`
}

// Ingredient represents a coffee ingredient.
type Ingredient struct {
	ID int `json:"id"`
}

// Order represents an order.
type Order struct {
	ID    int         `json:"id"`
	Items []OrderItem `json:"items"`
}

// OrderItem represents a line item within an order.
type OrderItem struct {
	Coffee   Coffee `json:"coffee"`
	Quantity int    `json:"quantity"`
}

// GetCoffees returns the list of coffees.
func (c *Client) GetCoffees(ctx context.Context) ([]Coffee, error) {
	var coffees []Coffee
	if err := c.do(ctx, http.MethodGet, "/coffees", nil, &coffees); err != nil {
		return nil, err
	}
	return coffees, nil
}

// GetOrder returns a single order by ID.
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var order Order
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/orders/%s", orderID), nil, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// CreateOrder creates an order with the provided items.
func (c *Client) CreateOrder(ctx context.Context, items []OrderItem) (*Order, error) {
	payload := map[string]any{"items": items}
	var order Order
	if err := c.do(ctx, http.MethodPost, "/orders", payload, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder replaces the items of an order.
func (c *Client) UpdateOrder(ctx context.Context, orderID string, items []OrderItem) (*Order, error) {
	payload := map[string]any{"items": items}
	var order Order
	if err := c.do(ctx, http.MethodPut, fmt.Sprintf("/orders/%s", orderID), payload, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// DeleteOrder deletes an order.
func (c *Client) DeleteOrder(ctx context.Context, orderID string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/orders/%s", orderID), nil, nil)
}
