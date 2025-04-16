package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"termpos/internal/db"
	"termpos/internal/models"
)

// RecordSale records a new sale
func RecordSale(sale models.Sale) (int, error) {
	// Validate the sale
	if err := sale.Validate(); err != nil {
		return 0, err
	}

	var id int64
	err := db.Transaction(func(tx *sql.Tx) error {
		// Get the product
		var product models.Product
		err := tx.QueryRow(
			"SELECT id, name, price, stock FROM products WHERE id = ?",
			sale.ProductID,
		).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrProductNotFound
			}
			return err
		}

		// Check if there's enough stock
		if product.Stock < sale.Quantity {
			return models.ErrInsufficientStock
		}

		// Calculate the total
		total := product.Price * float64(sale.Quantity)

		// Insert the sale
		result, err := tx.Exec(
			"INSERT INTO sales (product_id, quantity, price_per_unit, total, sale_date) VALUES (?, ?, ?, ?, ?)",
			sale.ProductID, sale.Quantity, product.Price, total, time.Now(),
		)
		if err != nil {
			return err
		}

		// Get the sale ID
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}

		// Update the product stock
		return DecrementProductStock(tx, sale.ProductID, sale.Quantity)
	})

	if err != nil {
		return 0, fmt.Errorf("failed to record sale: %w", err)
	}

	return int(id), nil
}

// GetAllSales retrieves all sales with product information
func GetAllSales() ([]models.Sale, error) {
	var sales []models.Sale

	query := `
		SELECT s.id, s.product_id, p.name, s.quantity, s.price_per_unit, s.total, s.sale_date
		FROM sales s
		JOIN products p ON s.product_id = p.id
		ORDER BY s.sale_date DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sale models.Sale
		err := rows.Scan(
			&sale.ID,
			&sale.ProductID,
			&sale.ProductName,
			&sale.Quantity,
			&sale.PricePerUnit,
			&sale.Total,
			&sale.SaleDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}
		sales = append(sales, sale)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales: %w", err)
	}

	return sales, nil
}
