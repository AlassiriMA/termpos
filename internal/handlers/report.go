package handlers

import (
	"fmt"

	"termpos/internal/db"
	"termpos/internal/models"
)

// GetRevenueReport generates a revenue report by product
func GetRevenueReport() ([]models.RevenueReport, error) {
	var report []models.RevenueReport

	query := `
		SELECT 
			p.id,
			p.name,
			SUM(s.quantity) as units_sold,
			SUM(s.total) as revenue
		FROM 
			sales s
		JOIN 
			products p ON s.product_id = p.id
		GROUP BY 
			p.id, p.name
		ORDER BY 
			revenue DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query revenue report: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RevenueReport
		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.UnitsSold,
			&item.Revenue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan revenue report item: %w", err)
		}
		report = append(report, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating revenue report: %w", err)
	}

	return report, nil
}

// GetDailySalesReport generates a daily sales report
func GetDailySalesReport() (map[string]float64, error) {
	report := make(map[string]float64)

	query := `
		SELECT 
			date(sale_date) as sale_day,
			SUM(total) as daily_total
		FROM 
			sales
		GROUP BY 
			date(sale_date)
		ORDER BY 
			sale_day DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily sales report: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var total float64
		err := rows.Scan(&day, &total)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily sales report item: %w", err)
		}
		report[day] = total
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily sales report: %w", err)
	}

	return report, nil
}

// GetTopSellingProducts returns the top selling products by quantity
func GetTopSellingProducts(limit int) ([]models.RevenueReport, error) {
	if limit <= 0 {
		limit = 5 // Default to top 5
	}

	var report []models.RevenueReport

	query := `
		SELECT 
			p.id,
			p.name,
			SUM(s.quantity) as units_sold,
			SUM(s.total) as revenue
		FROM 
			sales s
		JOIN 
			products p ON s.product_id = p.id
		GROUP BY 
			p.id, p.name
		ORDER BY 
			units_sold DESC
		LIMIT ?
	`

	rows, err := db.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RevenueReport
		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.UnitsSold,
			&item.Revenue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top selling product: %w", err)
		}
		report = append(report, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top selling products: %w", err)
	}

	return report, nil
}
