package models

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrEmptyName      = errors.New("name cannot be empty")
	ErrInvalidPrice   = errors.New("price must be greater than zero")
	ErrInvalidStock   = errors.New("stock cannot be negative")
	ErrInvalidID      = errors.New("invalid ID")
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrProductNotFound = errors.New("product not found")
)

// Sale represents a sales transaction
type Sale struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"` // For reporting
	Quantity    int       `json:"quantity"`
	PricePerUnit float64  `json:"price_per_unit,omitempty"` // For reporting
	Total       float64   `json:"total,omitempty"`         // For reporting
	SaleDate    time.Time `json:"sale_date"`
}

// Validate checks if the sale data is valid
func (s *Sale) Validate() error {
	if s.ProductID <= 0 {
		return ErrInvalidID
	}
	if s.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

// RevenueReport represents product revenue data
type RevenueReport struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitsSold   int     `json:"units_sold"`
	Revenue     float64 `json:"revenue"`
}
