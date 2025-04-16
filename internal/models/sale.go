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
        ID              int       `json:"id"`
        ProductID       int       `json:"product_id"`
        ProductName     string    `json:"product_name,omitempty"` // For reporting
        Quantity        int       `json:"quantity"`
        PricePerUnit    float64   `json:"price_per_unit,omitempty"` // For reporting
        DiscountAmount  float64   `json:"discount_amount,omitempty"`
        DiscountCode    string    `json:"discount_code,omitempty"`
        TaxRate         float64   `json:"tax_rate,omitempty"`
        TaxAmount       float64   `json:"tax_amount,omitempty"`
        Subtotal        float64   `json:"subtotal,omitempty"`
        Total           float64   `json:"total,omitempty"`
        PaymentMethod   string    `json:"payment_method,omitempty"`
        PaymentReference string   `json:"payment_reference,omitempty"`
        SaleDate        time.Time `json:"sale_date"`
        ReceiptNumber   string    `json:"receipt_number,omitempty"`
        CustomerEmail   string    `json:"customer_email,omitempty"`
        CustomerPhone   string    `json:"customer_phone,omitempty"`
        Notes           string    `json:"notes,omitempty"`
        
        // Customer loyalty fields
        CustomerID      int       `json:"customer_id,omitempty"`
        CustomerName    string    `json:"customer_name,omitempty"`
        LoyaltyDiscount float64   `json:"loyalty_discount,omitempty"`
        PointsEarned    int       `json:"points_earned,omitempty"`
        PointsUsed      int       `json:"points_used,omitempty"`
        LoyaltyTier     string    `json:"loyalty_tier,omitempty"`
        RewardID        int       `json:"reward_id,omitempty"`
        RewardName      string    `json:"reward_name,omitempty"`
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
        ProductID    int     `json:"product_id"`
        ProductName  string  `json:"product_name"`
        CategoryName string  `json:"category_name,omitempty"`
        UnitsSold    int     `json:"units_sold"`
        Revenue      float64 `json:"revenue"`
        Cost         float64 `json:"cost,omitempty"`
        Profit       float64 `json:"profit,omitempty"`
        ProfitMargin float64 `json:"profit_margin,omitempty"`
}

// CategoryReport represents category-based revenue data
type CategoryReport struct {
        CategoryID   int     `json:"category_id"`
        CategoryName string  `json:"category_name"`
        ProductCount int     `json:"product_count"`
        UnitsSold    int     `json:"units_sold"`
        Revenue      float64 `json:"revenue"`
        Profit       float64 `json:"profit,omitempty"`
}

// SalesTrendReport represents sales trend data over time
type SalesTrendReport struct {
        Period       string  `json:"period"`
        SaleCount    int     `json:"sale_count"`
        TotalRevenue float64 `json:"total_revenue"`
        TotalItems   int     `json:"total_items"`
}

// ProfitLossReport represents a profit and loss summary
type ProfitLossReport struct {
        TotalRevenue   float64 `json:"total_revenue"`
        TotalCost      float64 `json:"total_cost"`
        GrossProfit    float64 `json:"gross_profit"`
        ProfitMargin   float64 `json:"profit_margin"`
        TotalSold      int     `json:"total_sold"`
        Transactions   int     `json:"transactions"`
        AvgTransaction float64 `json:"avg_transaction"`
}

// SaleReport represents detailed sales data for reporting
type SaleReport struct {
        ProductID    int     `json:"product_id"`
        ProductName  string  `json:"product_name"`
        CategoryName string  `json:"category_name,omitempty"`
        Quantity     int     `json:"quantity"`
        Revenue      float64 `json:"revenue"`
        Cost         float64 `json:"cost,omitempty"`
        Profit       float64 `json:"profit,omitempty"`
}
