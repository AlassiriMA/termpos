package handlers

import (
        "database/sql"
        "time"

        "termpos/internal/db"
        "termpos/internal/models"
)

// AddProduct adds a new product to the database
func AddProduct(product models.Product) (int, error) {
        // Validate the product
        if err := product.Validate(); err != nil {
                return 0, err
        }

        // Use the enhanced database function that supports advanced inventory fields
        return db.AddProduct(product)
}

// GetProductByID retrieves a product by its ID
func GetProductByID(id int) (models.Product, error) {
        // Use the enhanced database function
        return db.GetProductByID(id)
}

// GetProductWithDetails retrieves a product with its related details
func GetProductWithDetails(id int) (models.ProductWithDetails, error) {
        // Use the enhanced database function
        return db.GetProductWithDetails(id)
}

// GetAllProducts retrieves all products
func GetAllProducts() ([]models.Product, error) {
        // Use the enhanced database function
        return db.GetAllProducts()
}

// GetAllProductsWithDetails retrieves all products with category and supplier details
func GetAllProductsWithDetails() ([]models.ProductWithDetails, error) {
        // Use the enhanced database function
        return db.GetAllProductsWithDetails()
}

// UpdateProductStock updates the stock of a product
func UpdateProductStock(id int, quantity int) error {
        // Use the enhanced database function
        return db.UpdateProductStock(id, quantity)
}

// DecrementProductStock decreases the stock of a product by the specified quantity
func DecrementProductStock(tx *sql.Tx, id int, quantity int) error {
        // Get the current stock
        var stock int
        err := tx.QueryRow("SELECT stock FROM products WHERE id = ?", id).Scan(&stock)
        if err != nil {
                if err == sql.ErrNoRows {
                        return models.ErrProductNotFound
                }
                return err
        }

        // Check if there's enough stock
        if stock < quantity {
                return models.ErrInsufficientStock
        }

        // Update the stock
        _, err = tx.Exec(
                "UPDATE products SET stock = stock - ?, updated_at = ? WHERE id = ?",
                quantity, time.Now(), id,
        )
        return err
}
