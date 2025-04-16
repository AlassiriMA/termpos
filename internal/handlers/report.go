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
	// Use empty strings for startDate and endDate to get all-time data
	return GetTopSellingProductsDateRange(limit, "", "")
}

// GetTopSellingProductsDateRange returns the top selling products for a specific date range
func GetTopSellingProductsDateRange(limit int, startDate, endDate string) ([]models.RevenueReport, error) {
	if limit <= 0 {
		limit = 5 // Default to top 5
	}

	var params []interface{}
	var report []models.RevenueReport

	query := `
		SELECT 
			p.id,
			p.name,
			COALESCE(c.name, 'Uncategorized') as category_name,
			SUM(s.quantity) as units_sold,
			SUM(s.total) as revenue,
			COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as cost
		FROM 
			sales s
		JOIN 
			products p ON s.product_id = p.id
		LEFT JOIN
			categories c ON p.category_id = c.id
		LEFT JOIN
			product_batches pb ON s.product_id = pb.product_id
	`
	
	// Add date filters if provided
	if startDate != "" || endDate != "" {
		query += " WHERE "
		
		if startDate != "" {
			query += "date(s.sale_date) >= ? "
			params = append(params, startDate)
			
			if endDate != "" {
				query += "AND "
			}
		}
		
		if endDate != "" {
			query += "date(s.sale_date) <= ? "
			params = append(params, endDate)
		}
	}
	
	query += `
		GROUP BY 
			p.id, p.name, c.name
		ORDER BY 
			units_sold DESC
		LIMIT ?
	`
	params = append(params, limit)

	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RevenueReport
		var cost float64
		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.CategoryName,
			&item.UnitsSold,
			&item.Revenue,
			&cost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top selling product: %w", err)
		}
		
		// Calculate profit metrics
		item.Cost = cost
		item.Profit = item.Revenue - cost
		if item.Revenue > 0 {
			item.ProfitMargin = (item.Profit / item.Revenue) * 100
		}
		
		report = append(report, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top selling products: %w", err)
	}

	return report, nil
}

// GetSalesForDateRange returns sales data for a specific date range
func GetSalesForDateRange(startDate, endDate string) ([]models.SaleReport, error) {
	var params []interface{}
	var sales []models.SaleReport

	query := `
		SELECT 
			p.id,
			p.name,
			COALESCE(c.name, 'Uncategorized') as category_name,
			SUM(s.quantity) as units_sold,
			SUM(s.total) as revenue,
			COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as cost
		FROM 
			sales s
		JOIN 
			products p ON s.product_id = p.id
		LEFT JOIN
			categories c ON p.category_id = c.id
		LEFT JOIN
			product_batches pb ON s.product_id = pb.product_id
	`
	
	// Add date filters if provided
	if startDate != "" || endDate != "" {
		query += " WHERE "
		
		if startDate != "" {
			query += "date(s.sale_date) >= ? "
			params = append(params, startDate)
			
			if endDate != "" {
				query += "AND "
			}
		}
		
		if endDate != "" {
			query += "date(s.sale_date) <= ? "
			params = append(params, endDate)
		}
	}
	
	query += `
		GROUP BY 
			p.id, p.name, c.name
		ORDER BY 
			revenue DESC
	`

	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales for date range: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sale models.SaleReport
		err := rows.Scan(
			&sale.ProductID,
			&sale.ProductName,
			&sale.CategoryName,
			&sale.Quantity,
			&sale.Revenue,
			&sale.Cost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales report: %w", err)
		}
		
		// Calculate profit
		sale.Profit = sale.Revenue - sale.Cost
		
		sales = append(sales, sale)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales: %w", err)
	}

	return sales, nil
}

// GetProfitLossReport returns a profit and loss summary for a date range
func GetProfitLossReport(startDate, endDate string) (models.ProfitLossReport, error) {
	var report models.ProfitLossReport
	var params []interface{}
	
	// Build the query with optional date filters
	query := `
		SELECT 
			COALESCE(SUM(s.total), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as total_cost,
			COALESCE(SUM(s.quantity), 0) as total_sold,
			COUNT(*) as transactions
		FROM sales s
		LEFT JOIN product_batches pb ON s.product_id = pb.product_id
	`
	
	// Add date filters if provided
	if startDate != "" || endDate != "" {
		query += " WHERE "
		
		if startDate != "" {
			query += "date(s.sale_date) >= ? "
			params = append(params, startDate)
			
			if endDate != "" {
				query += "AND "
			}
		}
		
		if endDate != "" {
			query += "date(s.sale_date) <= ? "
			params = append(params, endDate)
		}
	}
	
	// Execute the query with parameters
	err := db.DB.QueryRow(query, params...).Scan(
		&report.TotalRevenue, 
		&report.TotalCost,
		&report.TotalSold,
		&report.Transactions,
	)
	
	if err != nil {
		return models.ProfitLossReport{}, fmt.Errorf("failed to get profit/loss report: %w", err)
	}
	
	// Calculate derived metrics
	report.GrossProfit = report.TotalRevenue - report.TotalCost
	
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.GrossProfit / report.TotalRevenue) * 100
	}
	
	if report.Transactions > 0 {
		report.AvgTransaction = report.TotalRevenue / float64(report.Transactions)
	}

	return report, nil
}

// GetSalesByCategory returns sales data grouped by category 
func GetSalesByCategory(startDate, endDate string) ([]models.CategoryReport, error) {
	var params []interface{}
	
	query := `
		SELECT 
			c.id,
			c.name,
			COUNT(DISTINCT p.id) as product_count,
			COALESCE(SUM(s.quantity), 0) as units_sold,
			COALESCE(SUM(s.total), 0) as revenue,
			COALESCE(SUM(s.total) - SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as profit
		FROM 
			categories c
		LEFT JOIN 
			products p ON c.id = p.category_id
		LEFT JOIN 
			sales s ON p.id = s.product_id
		LEFT JOIN
			product_batches pb ON p.id = pb.product_id
	`
	
	// Add date filters if provided
	if startDate != "" || endDate != "" {
		query += " WHERE "
		
		if startDate != "" {
			query += "date(s.sale_date) >= ? "
			params = append(params, startDate)
			
			if endDate != "" {
				query += "AND "
			}
		}
		
		if endDate != "" {
			query += "date(s.sale_date) <= ? "
			params = append(params, endDate)
		}
	}
	
	query += `
		GROUP BY 
			c.id, c.name
		ORDER BY 
			revenue DESC
	`

	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query category sales: %w", err)
	}
	defer rows.Close()

	var categories []models.CategoryReport
	for rows.Next() {
		var c models.CategoryReport
		if err := rows.Scan(
			&c.CategoryID,
			&c.CategoryName,
			&c.ProductCount,
			&c.UnitsSold,
			&c.Revenue,
			&c.Profit,
		); err != nil {
			return nil, fmt.Errorf("failed to scan category sale: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category sales: %w", err)
	}

	return categories, nil
}

// GetSalesTrends returns sales trend data grouped by period (day, week, month)
func GetSalesTrends(startDate, endDate, groupBy string) ([]models.SalesTrendReport, error) {
	var params []interface{}
	var dateFormat string
	
	// Determine grouping (day or month)
	switch groupBy {
	case "month":
		dateFormat = "%Y-%m"
	case "week":
		dateFormat = "%Y-%W"
	default:
		dateFormat = "%Y-%m-%d"  // Default to day
	}
	
	query := `
		SELECT 
			strftime(?, sale_date) as period,
			COUNT(*) as sale_count,
			SUM(total) as total_revenue,
			SUM(quantity) as total_items
		FROM 
			sales
	`
	params = append(params, dateFormat)
	
	// Add date filters if provided
	if startDate != "" || endDate != "" {
		query += " WHERE "
		
		if startDate != "" {
			query += "date(sale_date) >= ? "
			params = append(params, startDate)
			
			if endDate != "" {
				query += "AND "
			}
		}
		
		if endDate != "" {
			query += "date(sale_date) <= ? "
			params = append(params, endDate)
		}
	}
	
	query += `
		GROUP BY 
			period
		ORDER BY 
			sale_date
	`

	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales trends: %w", err)
	}
	defer rows.Close()

	var trends []models.SalesTrendReport
	for rows.Next() {
		var t models.SalesTrendReport
		if err := rows.Scan(
			&t.Period,
			&t.SaleCount,
			&t.TotalRevenue,
			&t.TotalItems,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sales trend: %w", err)
		}
		trends = append(trends, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales trends: %w", err)
	}

	return trends, nil
}