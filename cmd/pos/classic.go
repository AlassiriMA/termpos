package main

import (
        "fmt"
        "strconv"
        "strings"
        "time"

        "github.com/olekukonko/tablewriter"
        "github.com/spf13/cobra"

        "termpos/internal/db"
        "termpos/internal/handlers"
        "termpos/internal/models"
)

// initClassicCommands sets up the commands for classic CLI mode
func initClassicCommands() {
        // Product commands
        var addCmd = &cobra.Command{
                Use:   "add [name] [price] [stock]",
                Short: "Add a new product",
                Long:  `Add a new product to the inventory with name, price, and optional stock quantity.`,
                Args:  cobra.MinimumNArgs(2),
                RunE: func(cmd *cobra.Command, args []string) error {
                        name := args[0]
                        price, err := strconv.ParseFloat(args[1], 64)
                        if err != nil {
                                return fmt.Errorf("invalid price: %w", err)
                        }

                        stock := 0
                        if len(args) > 2 {
                                stock, err = strconv.Atoi(args[2])
                                if err != nil {
                                        return fmt.Errorf("invalid stock quantity: %w", err)
                                }
                        }

                        product := models.Product{
                                Name:  name,
                                Price: price,
                                Stock: stock,
                        }

                        id, err := handlers.AddProduct(product)
                        if err != nil {
                                return fmt.Errorf("failed to add product: %w", err)
                        }

                        fmt.Printf("Product added successfully with ID: %d\n", id)
                        return nil
                },
        }

        var inventoryCmd = &cobra.Command{
                Use:   "inventory",
                Short: "List all products in inventory",
                RunE: func(cmd *cobra.Command, args []string) error {
                        products, err := handlers.GetAllProducts()
                        if err != nil {
                                return fmt.Errorf("failed to get products: %w", err)
                        }

                        if len(products) == 0 {
                                fmt.Println("No products found in inventory")
                                return nil
                        }

                        // Create a table for output
                        table := tablewriter.NewWriter(cmd.OutOrStdout())
                        table.SetHeader([]string{"ID", "Name", "Price", "Stock"})
                        table.SetBorder(false)

                        for _, p := range products {
                                table.Append([]string{
                                        fmt.Sprintf("%d", p.ID),
                                        p.Name,
                                        fmt.Sprintf("$%.2f", p.Price),
                                        fmt.Sprintf("%d", p.Stock),
                                })
                        }

                        table.Render()
                        return nil
                },
        }

        var updateStockCmd = &cobra.Command{
                Use:   "update-stock [product_id] [quantity]",
                Short: "Update product stock",
                Long:  `Update the stock quantity for a product by ID.`,
                Args:  cobra.ExactArgs(2),
                RunE: func(cmd *cobra.Command, args []string) error {
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                return fmt.Errorf("invalid product ID: %w", err)
                        }

                        quantity, err := strconv.Atoi(args[1])
                        if err != nil {
                                return fmt.Errorf("invalid quantity: %w", err)
                        }

                        if err := handlers.UpdateProductStock(id, quantity); err != nil {
                                return fmt.Errorf("failed to update stock: %w", err)
                        }

                        fmt.Printf("Stock updated successfully for product ID: %d\n", id)
                        return nil
                },
        }

        // Sale commands
        var sellCmd = &cobra.Command{
                Use:   "sell [product_id] [quantity]",
                Short: "Sell a product",
                Long:  `Record a sale of a product with the specified quantity.`,
                Args:  cobra.ExactArgs(2),
                RunE: func(cmd *cobra.Command, args []string) error {
                        productID, err := strconv.Atoi(args[0])
                        if err != nil {
                                return fmt.Errorf("invalid product ID: %w", err)
                        }

                        quantity, err := strconv.Atoi(args[1])
                        if err != nil {
                                return fmt.Errorf("invalid quantity: %w", err)
                        }

                        sale := models.Sale{
                                ProductID: productID,
                                Quantity:  quantity,
                        }

                        id, err := handlers.RecordSale(sale)
                        if err != nil {
                                return fmt.Errorf("failed to record sale: %w", err)
                        }

                        fmt.Printf("Sale recorded successfully with ID: %d\n", id)
                        return nil
                },
        }

        // Report commands
        var reportCmd = &cobra.Command{
                Use:   "report [type]",
                Short: "Generate a report",
                Long:  `Generate various reports: "sales", "inventory", "revenue", "summary", "top", "daily"`,
                Args:  cobra.ExactArgs(1),
                RunE: func(cmd *cobra.Command, args []string) error {
                        reportType := strings.ToLower(args[0])

                        switch reportType {
                        case "sales":
                                return generateSalesReport(cmd)
                        case "inventory":
                                return generateInventoryReport(cmd)
                        case "revenue":
                                return generateRevenueReport(cmd)
                        case "summary":
                                return generateSummaryReport(cmd)
                        case "top":
                                return generateTopProductsReport(cmd)
                        case "daily":
                                return generateDailySalesReport(cmd)
                        default:
                                return fmt.Errorf("unknown report type: %s", reportType)
                        }
                },
        }

        // Add commands to the root command
        rootCmd.AddCommand(addCmd)
        rootCmd.AddCommand(inventoryCmd)
        rootCmd.AddCommand(updateStockCmd)
        rootCmd.AddCommand(sellCmd)
        rootCmd.AddCommand(reportCmd)
}

func generateSalesReport(cmd *cobra.Command) error {
        sales, err := handlers.GetAllSales()
        if err != nil {
                return fmt.Errorf("failed to get sales data: %w", err)
        }

        if len(sales) == 0 {
                fmt.Println("No sales data available")
                return nil
        }

        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Sale ID", "Product", "Quantity", "Total", "Date"})
        table.SetBorder(false)

        for _, s := range sales {
                table.Append([]string{
                        fmt.Sprintf("%d", s.ID),
                        s.ProductName,
                        fmt.Sprintf("%d", s.Quantity),
                        fmt.Sprintf("$%.2f", s.Total),
                        s.SaleDate.Format("2006-01-02 15:04:05"),
                })
        }

        fmt.Println("Sales Report:")
        table.Render()
        return nil
}

func generateInventoryReport(cmd *cobra.Command) error {
        products, err := handlers.GetAllProducts()
        if err != nil {
                return fmt.Errorf("failed to get inventory data: %w", err)
        }

        if len(products) == 0 {
                fmt.Println("No products found in inventory")
                return nil
        }

        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Product ID", "Name", "Price", "Stock", "Value"})
        table.SetBorder(false)

        var totalValue float64
        for _, p := range products {
                value := p.Price * float64(p.Stock)
                totalValue += value
                table.Append([]string{
                        fmt.Sprintf("%d", p.ID),
                        p.Name,
                        fmt.Sprintf("$%.2f", p.Price),
                        fmt.Sprintf("%d", p.Stock),
                        fmt.Sprintf("$%.2f", value),
                })
        }

        fmt.Println("Inventory Report:")
        table.Render()
        fmt.Printf("Total Inventory Value: $%.2f\n", totalValue)
        return nil
}

func generateRevenueReport(cmd *cobra.Command) error {
        revenue, err := handlers.GetRevenueReport()
        if err != nil {
                return fmt.Errorf("failed to get revenue data: %w", err)
        }

        if len(revenue) == 0 {
                fmt.Println("No revenue data available")
                return nil
        }

        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Product", "Units Sold", "Revenue"})
        table.SetBorder(false)

        var totalRevenue float64
        for _, r := range revenue {
                totalRevenue += r.Revenue
                table.Append([]string{
                        r.ProductName,
                        fmt.Sprintf("%d", r.UnitsSold),
                        fmt.Sprintf("$%.2f", r.Revenue),
                })
        }

        fmt.Println("Revenue Report:")
        table.Render()
        fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
        return nil
}

// generateSummaryReport generates a summary report with total revenue and total items sold
func generateSummaryReport(cmd *cobra.Command) error {
        summary, err := db.GetSalesSummary()
        if err != nil {
                return fmt.Errorf("failed to get sales summary: %w", err)
        }

        if summary.TotalTransactions == 0 {
                fmt.Println("No sales data available for summary report")
                return nil
        }

        fmt.Println("Sales Summary Report:")
        fmt.Println("--------------------")
        fmt.Printf("Total Revenue: $%.2f\n", summary.TotalRevenue)
        fmt.Printf("Total Items Sold: %d\n", summary.TotalItemsSold)
        fmt.Printf("Total Transactions: %d\n", summary.TotalTransactions)
        
        // Calculate average transaction value if there are transactions
        if summary.TotalTransactions > 0 {
                avg := summary.TotalRevenue / float64(summary.TotalTransactions)
                fmt.Printf("Average Transaction Value: $%.2f\n", avg)
        }

        return nil
}

// generateTopProductsReport generates a report of top-selling products by quantity
func generateTopProductsReport(cmd *cobra.Command) error {
        // Default to top 5 products
        topProducts, err := db.GetTopSellingProducts(5)
        if err != nil {
                return fmt.Errorf("failed to get top selling products: %w", err)
        }

        if len(topProducts) == 0 {
                fmt.Println("No sales data available for top products report")
                return nil
        }

        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Rank", "Product", "Units Sold", "Revenue"})
        table.SetBorder(false)

        for i, p := range topProducts {
                table.Append([]string{
                        fmt.Sprintf("%d", i+1),
                        p.ProductName,
                        fmt.Sprintf("%d", p.Quantity),
                        fmt.Sprintf("$%.2f", p.Revenue),
                })
        }

        fmt.Println("Top Selling Products Report:")
        table.Render()
        return nil
}

// generateDailySalesReport generates a report of sales for today grouped by product
func generateDailySalesReport(cmd *cobra.Command) error {
        dailySales, err := db.GetDailySales()
        if err != nil {
                return fmt.Errorf("failed to get daily sales: %w", err)
        }

        if len(dailySales) == 0 {
                fmt.Printf("No sales data available for today (%s)\n", time.Now().Format("2006-01-02"))
                return nil
        }

        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Product", "Units Sold", "Revenue"})
        table.SetBorder(false)

        var totalUnits int
        var totalRevenue float64

        for _, s := range dailySales {
                table.Append([]string{
                        s.ProductName,
                        fmt.Sprintf("%d", s.Quantity),
                        fmt.Sprintf("$%.2f", s.Revenue),
                })
                totalUnits += s.Quantity
                totalRevenue += s.Revenue
        }

        fmt.Printf("Daily Sales Report for %s:\n", time.Now().Format("2006-01-02"))
        table.Render()
        fmt.Printf("Total Units Sold Today: %d\n", totalUnits)
        fmt.Printf("Total Revenue Today: $%.2f\n", totalRevenue)
        return nil
}
