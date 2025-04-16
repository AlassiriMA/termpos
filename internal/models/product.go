package models

import (
	"time"
)

// Product represents a product in the inventory
type Product struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks if the product data is valid
func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrEmptyName
	}
	if p.Price <= 0 {
		return ErrInvalidPrice
	}
	if p.Stock < 0 {
		return ErrInvalidStock
	}
	return nil
}

// HasSufficientStock checks if the product has enough stock for a sale
func (p *Product) HasSufficientStock(quantity int) bool {
	return p.Stock >= quantity
}
