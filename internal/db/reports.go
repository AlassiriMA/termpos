package db

import (
        "fmt"
        "time"
)

// SalesSummary represents a summary of sales data
type SalesSummary struct {
        TotalRevenue  float64
        TotalItemsSold int
        TotalTransactions int
}

// DailySale represents a sales data for a specific day and product
type DailySale struct {
        ProductID   int
        ProductName string
        Quantity    int
        Revenue     float64
}

// TopSellingProduct represents a product with its sales performance
type TopSellingProduct struct {
        ProductID   int
        ProductName string
        Quantity    int
        Revenue     float64
}

// GetSalesSummary returns overall sales metrics
func GetSalesSummary() (SalesSummary, error) {
        var summary SalesSummary

        // Get total revenue and items sold
        query := `
                SELECT 
                        COALESCE(SUM(total), 0) as total_revenue,
                        COALESCE(SUM(quantity), 0) as total_items_sold,
                        COUNT(*) as total_transactions
                FROM sales
        `
        
        err := DB.QueryRow(query).Scan(
                &summary.TotalRevenue, 
                &summary.TotalItemsSold,
                &summary.TotalTransactions,
        )
        
        if err != nil {
                return SalesSummary{}, fmt.Errorf("failed to get sales summary: %w", err)
        }

        return summary, nil
}

// GetTopSellingProducts returns the top selling products ordered by quantity
func GetTopSellingProducts(limit int) ([]TopSellingProduct, error) {
        if limit <= 0 {
                limit = 5 // Default to top 5
        }

        query := `
                SELECT 
                        p.id,
                        p.name,
                        SUM(s.quantity) as quantity,
                        SUM(s.total) as revenue
                FROM 
                        sales s
                JOIN 
                        products p ON s.product_id = p.id
                GROUP BY 
                        p.id, p.name
                ORDER BY 
                        quantity DESC
                LIMIT ?
        `

        rows, err := DB.Query(query, limit)
        if err != nil {
                return nil, fmt.Errorf("failed to query top selling products: %w", err)
        }
        defer rows.Close()

        var products []TopSellingProduct
        for rows.Next() {
                var p TopSellingProduct
                if err := rows.Scan(&p.ProductID, &p.ProductName, &p.Quantity, &p.Revenue); err != nil {
                        return nil, fmt.Errorf("failed to scan top selling product: %w", err)
                }
                products = append(products, p)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating top selling products: %w", err)
        }

        return products, nil
}

// GetDailySales returns sales data for today grouped by product
func GetDailySales() ([]DailySale, error) {
        // Get today's date in SQLite format (YYYY-MM-DD)
        today := time.Now().Format("2006-01-02")
        
        query := `
                SELECT 
                        p.id,
                        p.name,
                        SUM(s.quantity) as quantity,
                        SUM(s.total) as revenue
                FROM 
                        sales s
                JOIN 
                        products p ON s.product_id = p.id
                WHERE 
                        date(s.sale_date) = ?
                GROUP BY 
                        p.id, p.name
                ORDER BY 
                        quantity DESC
        `

        rows, err := DB.Query(query, today)
        if err != nil {
                return nil, fmt.Errorf("failed to query daily sales: %w", err)
        }
        defer rows.Close()

        var sales []DailySale
        for rows.Next() {
                var s DailySale
                if err := rows.Scan(&s.ProductID, &s.ProductName, &s.Quantity, &s.Revenue); err != nil {
                        return nil, fmt.Errorf("failed to scan daily sale: %w", err)
                }
                sales = append(sales, s)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating daily sales: %w", err)
        }

        return sales, nil
}

// GenerateCSVReport generates a CSV report for the specified report type
// This is a stub for future implementation
func GenerateCSVReport(reportType string, filepath string) error {
        // This is a stub function that will be implemented in the future
        // It would generate different CSV reports based on the reportType
        // and save them to the specified filepath
        
        // For now, just return a not implemented error
        return fmt.Errorf("CSV report generation not yet implemented for: %s", reportType)
}