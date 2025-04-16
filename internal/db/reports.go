package db

import (
        "fmt"
        "strings"
        "time"
)

// SalesSummary represents a summary of sales data
type SalesSummary struct {
        TotalRevenue      float64
        TotalItemsSold    int
        TotalTransactions int
        TotalCost         float64  // Cost of goods sold
        GrossProfit       float64  // Revenue minus cost
        ProfitMargin      float64  // Profit as percentage of revenue
}

// DailySale represents a sales data for a specific day and product
type DailySale struct {
        ProductID   int
        ProductName string
        Quantity    int
        Revenue     float64
        Cost        float64
        Profit      float64
        CategoryName string
}

// SalesTrend represents sales data for a specific time period
type SalesTrend struct {
        Period      string  // Day or month depending on grouping
        SaleCount   int     // Number of transactions
        TotalRevenue float64
        TotalItems  int
}

// CategorySales represents sales data grouped by product category
type CategorySales struct {
        CategoryID   int
        CategoryName string
        ProductCount int
        ItemsSold    int
        Revenue      float64
        Profit       float64
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
        return GetSalesSummaryDateRange("", "")
}

// GetSalesSummaryDateRange returns sales metrics for a specific date range
// If startDate or endDate is empty, no date filtering is applied for that bound
func GetSalesSummaryDateRange(startDate, endDate string) (SalesSummary, error) {
        var summary SalesSummary
        var params []interface{}
        
        // Start building the query
        query := `
                SELECT 
                        COALESCE(SUM(s.total), 0) as total_revenue,
                        COALESCE(SUM(s.quantity), 0) as total_items_sold,
                        COUNT(*) as total_transactions,
                        COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as total_cost
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
        var totalCost float64
        err := DB.QueryRow(query, params...).Scan(
                &summary.TotalRevenue, 
                &summary.TotalItemsSold,
                &summary.TotalTransactions,
                &totalCost,
        )
        
        if err != nil {
                return SalesSummary{}, fmt.Errorf("failed to get sales summary: %w", err)
        }
        
        // Calculate profit metrics
        summary.TotalCost = totalCost
        summary.GrossProfit = summary.TotalRevenue - summary.TotalCost
        
        if summary.TotalRevenue > 0 {
                summary.ProfitMargin = (summary.GrossProfit / summary.TotalRevenue) * 100
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
        return GetSalesForDateRange(today, today)
}

// GetSalesForDateRange returns sales data for a specific date range grouped by product
func GetSalesForDateRange(startDate, endDate string) ([]DailySale, error) {
        var params []interface{}
        
        query := `
                SELECT 
                        p.id,
                        p.name,
                        SUM(s.quantity) as quantity,
                        SUM(s.total) as revenue,
                        COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as cost,
                        COALESCE(c.name, 'Uncategorized') as category_name
                FROM 
                        sales s
                JOIN 
                        products p ON s.product_id = p.id
                LEFT JOIN
                        product_batches pb ON s.product_id = pb.product_id
                LEFT JOIN
                        categories c ON p.category_id = c.id
                WHERE 1=1
        `
        
        // Add date filters if provided
        if startDate != "" {
                query += " AND date(s.sale_date) >= ? "
                params = append(params, startDate)
        }
        
        if endDate != "" {
                query += " AND date(s.sale_date) <= ? "
                params = append(params, endDate)
        }
        
        query += `
                GROUP BY 
                        p.id, p.name, c.name
                ORDER BY 
                        quantity DESC
        `

        rows, err := DB.Query(query, params...)
        if err != nil {
                return nil, fmt.Errorf("failed to query sales for date range: %w", err)
        }
        defer rows.Close()

        var sales []DailySale
        for rows.Next() {
                var s DailySale
                if err := rows.Scan(
                        &s.ProductID, 
                        &s.ProductName, 
                        &s.Quantity, 
                        &s.Revenue,
                        &s.Cost,
                        &s.CategoryName,
                ); err != nil {
                        return nil, fmt.Errorf("failed to scan daily sale: %w", err)
                }
                
                // Calculate profit
                s.Profit = s.Revenue - s.Cost
                
                sales = append(sales, s)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating sales: %w", err)
        }

        return sales, nil
}

// GetCategorySales returns sales data grouped by product category
func GetCategorySales(startDate, endDate string) ([]CategorySales, error) {
        var params []interface{}
        
        query := `
                SELECT 
                        c.id,
                        c.name,
                        COUNT(DISTINCT p.id) as product_count,
                        COALESCE(SUM(s.quantity), 0) as items_sold,
                        COALESCE(SUM(s.total), 0) as revenue,
                        COALESCE(SUM(s.total), 0) - COALESCE(SUM(CASE WHEN pb.cost_price > 0 THEN pb.cost_price * s.quantity ELSE 0 END), 0) as profit
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

        rows, err := DB.Query(query, params...)
        if err != nil {
                return nil, fmt.Errorf("failed to query category sales: %w", err)
        }
        defer rows.Close()

        var categories []CategorySales
        for rows.Next() {
                var c CategorySales
                if err := rows.Scan(
                        &c.CategoryID,
                        &c.CategoryName,
                        &c.ProductCount,
                        &c.ItemsSold,
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

// GetSalesTrends returns sales trend data grouped by day or month
func GetSalesTrends(startDate, endDate string, groupBy string) ([]SalesTrend, error) {
        var params []interface{}
        var dateFormat string
        
        // Determine grouping (day or month)
        switch strings.ToLower(groupBy) {
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

        rows, err := DB.Query(query, params...)
        if err != nil {
                return nil, fmt.Errorf("failed to query sales trends: %w", err)
        }
        defer rows.Close()

        var trends []SalesTrend
        for rows.Next() {
                var t SalesTrend
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

// GetTopSellingProductsDateRange is similar to GetTopSellingProducts but with date range filtering
func GetTopSellingProductsDateRange(limit int, startDate, endDate string) ([]TopSellingProduct, error) {
        if limit <= 0 {
                limit = 5 // Default to top 5
        }
        
        var params []interface{}
        
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
                        p.id, p.name
                ORDER BY 
                        quantity DESC
                LIMIT ?
        `
        params = append(params, limit)

        rows, err := DB.Query(query, params...)
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

// GenerateCSVReport generates a CSV report for the specified report type
// This is a stub for future implementation
func GenerateCSVReport(reportType string, filepath string) error {
        // This is a stub function that will be implemented in the future
        // It would generate different CSV reports based on the reportType
        // and save them to the specified filepath
        
        // For now, just return a not implemented error
        return fmt.Errorf("CSV report generation not yet implemented for: %s", reportType)
}