package models

import (
        "time"
)

// Product represents a product in the inventory
type Product struct {
        ID            int       `json:"id"`
        Name          string    `json:"name"`
        Price         float64   `json:"price"`
        Stock         int       `json:"stock"`
        CategoryID    int       `json:"category_id"`
        LowStockAlert int       `json:"low_stock_alert"` // Threshold for low stock alerts
        DefaultSupplierID int   `json:"default_supplier_id"`
        SKU           string    `json:"sku"`
        Description   string    `json:"description"`
        CreatedAt     time.Time `json:"created_at"`
        UpdatedAt     time.Time `json:"updated_at"`
}

// ProductWithDetails represents a product with its related data
type ProductWithDetails struct {
        Product
        CategoryName string `json:"category_name"`
        SupplierName string `json:"supplier_name"`
        BatchCount   int    `json:"batch_count"`
        LocationsCount int  `json:"locations_count"`
        HasExpiredBatches bool `json:"has_expired_batches"`
        IsLowStock   bool   `json:"is_low_stock"`
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

// IsLowOnStock checks if the product stock is below the low stock alert threshold
func (p *Product) IsLowOnStock() bool {
        // If LowStockAlert is 0, no alert is set
        if p.LowStockAlert == 0 {
                return false
        }
        return p.Stock <= p.LowStockAlert
}

// Category represents a product category
type Category struct {
        ID          int       `json:"id"`
        Name        string    `json:"name"`
        Description string    `json:"description"`
        ParentID    int       `json:"parent_id"` // For hierarchical categories
        CreatedAt   time.Time `json:"created_at"`
        UpdatedAt   time.Time `json:"updated_at"`
}

// Location represents a physical storage location for inventory
type Location struct {
        ID          int       `json:"id"`
        Name        string    `json:"name"`
        Address     string    `json:"address"`
        Description string    `json:"description"`
        IsActive    bool      `json:"is_active"`
        CreatedAt   time.Time `json:"created_at"`
        UpdatedAt   time.Time `json:"updated_at"`
}

// Supplier represents a product supplier
type Supplier struct {
        ID          int       `json:"id"`
        Name        string    `json:"name"`
        Contact     string    `json:"contact"`
        Email       string    `json:"email"`
        Phone       string    `json:"phone"`
        Address     string    `json:"address"`
        Notes       string    `json:"notes"`
        IsActive    bool      `json:"is_active"`
        CreatedAt   time.Time `json:"created_at"`
        UpdatedAt   time.Time `json:"updated_at"`
}

// ProductBatch represents a batch of products with expiration date
type ProductBatch struct {
        ID           int       `json:"id"`
        ProductID    int       `json:"product_id"`
        LocationID   int       `json:"location_id"`
        SupplierID   int       `json:"supplier_id"`
        Quantity     int       `json:"quantity"`
        BatchNumber  string    `json:"batch_number"`
        ExpiryDate   time.Time `json:"expiry_date"`
        ManufactureDate time.Time `json:"manufacture_date"`
        CostPrice    float64   `json:"cost_price"`
        ReceiptDate  time.Time `json:"receipt_date"`
        CreatedAt    time.Time `json:"created_at"`
        UpdatedAt    time.Time `json:"updated_at"`
}

// ProductLocation represents the inventory of a product at a specific location
type ProductLocation struct {
        ID         int       `json:"id"`
        ProductID  int       `json:"product_id"`
        LocationID int       `json:"location_id"`
        Quantity   int       `json:"quantity"`
        CreatedAt  time.Time `json:"created_at"`
        UpdatedAt  time.Time `json:"updated_at"`
}

// IsExpired checks if the batch has expired
func (b *ProductBatch) IsExpired() bool {
        return !b.ExpiryDate.IsZero() && b.ExpiryDate.Before(time.Now())
}

// IsExpiringSoon checks if the batch is expiring within the given number of days
func (b *ProductBatch) IsExpiringSoon(days int) bool {
        if b.ExpiryDate.IsZero() {
                return false
        }
        expiryThreshold := time.Now().AddDate(0, 0, days)
        return b.ExpiryDate.Before(expiryThreshold) && !b.IsExpired()
}
