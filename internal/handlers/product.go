package handlers

import (
	"database/sql"
	"fmt"
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

	// Insert the product
	query := `
		INSERT INTO products (name, price, stock, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()

	var id int64
	err := db.Transaction(func(tx *sql.Tx) error {
		result, err := tx.Exec(query, product.Name, product.Price, product.Stock, now, now)
		if err != nil {
			return err
		}

		id, err = result.LastInsertId()
		return err
	})

	if err != nil {
		return 0, fmt.Errorf("failed to add product: %w", err)
	}

	return int(id), nil
}

// GetProductByID retrieves a product by its ID
func GetProductByID(id int) (models.Product, error) {
	var product models.Product

	query := `
		SELECT id, name, price, stock, created_at, updated_at
		FROM products
		WHERE id = ?
	`

	err := db.DB.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Product{}, models.ErrProductNotFound
		}
		return models.Product{}, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// GetAllProducts retrieves all products
func GetAllProducts() ([]models.Product, error) {
	var products []models.Product

	query := `
		SELECT id, name, price, stock, created_at, updated_at
		FROM products
		ORDER BY name
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// UpdateProductStock updates the stock of a product
func UpdateProductStock(id int, quantity int) error {
	if quantity < 0 {
		return models.ErrInvalidStock
	}

	query := `
		UPDATE products
		SET stock = ?, updated_at = ?
		WHERE id = ?
	`

	err := db.Transaction(func(tx *sql.Tx) error {
		// Check if the product exists
		var exists bool
		err := tx.QueryRow("SELECT 1 FROM products WHERE id = ?", id).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrProductNotFound
			}
			return err
		}

		// Update the stock
		_, err = tx.Exec(query, quantity, time.Now(), id)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	return nil
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
