package main

import (
        "fmt"
        "strconv"
        "strings"
        "time"

        "github.com/olekukonko/tablewriter"
        "github.com/spf13/cobra"

        "termpos/internal/auth"
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
                        // Check if user is authorized to add products
                        if err := auth.RequirePermission("product:manage"); err != nil {
                                return err
                        }
                        
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
                                Name:              name,
                                Price:             price,
                                Stock:             stock,
                                CategoryID:        productCategory,
                                DefaultSupplierID: productSupplier,
                                LowStockAlert:     productLowStock,
                                SKU:               productSKU,
                                Description:       productDescription,
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
                        // Check if user is authorized to view inventory
                        if err := auth.RequirePermission("inventory:view"); err != nil {
                                return err
                        }
                        
                        products, err := handlers.GetAllProducts()
                        if err != nil {
                                return fmt.Errorf("failed to get products: %w", err)
                        }

                        if len(products) == 0 {
                                fmt.Println("No products found in inventory")
                                return nil
                        }

                        // Create a table for output - use detailed view with advanced inventory fields
                        detailedProducts, err := db.GetAllProductsWithDetails()
                        if err == nil && len(detailedProducts) > 0 {
                                // If we have detailed products, show enhanced view
                                table := tablewriter.NewWriter(cmd.OutOrStdout())
                                table.SetHeader([]string{"ID", "Name", "Price", "Stock", "Category", "Low Stock", "Status"})
                                table.SetBorder(false)

                                for _, p := range detailedProducts {
                                        // Format low stock status
                                        lowStockStatus := "-"
                                        status := "OK"
                                        
                                        if p.LowStockAlert > 0 {
                                                lowStockStatus = fmt.Sprintf("%d", p.LowStockAlert)
                                                if p.IsLowStock {
                                                        status = "LOW STOCK"
                                                }
                                        }
                                        
                                        if p.HasExpiredBatches {
                                                status = "EXPIRED BATCHES"
                                        }

                                        table.Append([]string{
                                                fmt.Sprintf("%d", p.ID),
                                                p.Name,
                                                fmt.Sprintf("$%.2f", p.Price),
                                                fmt.Sprintf("%d", p.Stock),
                                                p.CategoryName,
                                                lowStockStatus,
                                                status,
                                        })
                                }
                                table.Render()
                        } else {
                                // Fall back to basic view
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
                        }
                        return nil
                },
        }

        var updateStockCmd = &cobra.Command{
                Use:   "update-stock [product_id] [quantity]",
                Short: "Update product stock",
                Long:  `Update the stock quantity for a product by ID.`,
                Args:  cobra.ExactArgs(2),
                RunE: func(cmd *cobra.Command, args []string) error {
                        // Check if user is authorized to manage products
                        if err := auth.RequirePermission("product:manage"); err != nil {
                                return err
                        }
                        
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
                        // Check if user is authorized to create sales
                        if err := auth.RequirePermission("sales:create"); err != nil {
                                return err
                        }
                        
                        productID, err := strconv.Atoi(args[0])
                        if err != nil {
                                return fmt.Errorf("invalid product ID: %w", err)
                        }

                        quantity, err := strconv.Atoi(args[1])
                        if err != nil {
                                return fmt.Errorf("invalid quantity: %w", err)
                        }

                        // Get flags for enhanced sales functionality
                        discountAmount, _ := cmd.Flags().GetFloat64("discount")
                        discountCode, _ := cmd.Flags().GetString("discount-code")
                        taxRate, _ := cmd.Flags().GetFloat64("tax-rate") 
                        paymentMethod, _ := cmd.Flags().GetString("payment-method")
                        paymentRef, _ := cmd.Flags().GetString("payment-ref")
                        customerEmail, _ := cmd.Flags().GetString("email")
                        customerPhone, _ := cmd.Flags().GetString("phone")
                        notes, _ := cmd.Flags().GetString("notes")
                        printReceipt, _ := cmd.Flags().GetBool("print-receipt")
                        emailReceipt, _ := cmd.Flags().GetBool("email-receipt")
                        
                        // Get customer loyalty flags
                        customerID, _ := cmd.Flags().GetInt("customer-id")
                        applyLoyalty, _ := cmd.Flags().GetBool("apply-loyalty")
                        pointsUsed, _ := cmd.Flags().GetInt("points-used")
                        rewardID, _ := cmd.Flags().GetInt("reward-id")
                        
                        // If we want to disable loyalty features for this transaction
                        if !applyLoyalty {
                                customerID = 0
                                pointsUsed = 0
                                rewardID = 0
                        }
                        
                        // Convert tax rate from percentage to decimal if provided
                        if taxRate > 0 {
                                taxRate = taxRate / 100.0
                        }

                        sale := models.Sale{
                                ProductID:         productID,
                                Quantity:          quantity,
                                DiscountAmount:    discountAmount,
                                DiscountCode:      discountCode,
                                TaxRate:           taxRate,
                                PaymentMethod:     paymentMethod,
                                PaymentReference:  paymentRef,
                                CustomerEmail:     customerEmail,
                                CustomerPhone:     customerPhone,
                                Notes:             notes,
                                CustomerID:        customerID,
                                PointsUsed:        pointsUsed,
                                RewardID:          rewardID,
                        }

                        id, err := handlers.RecordSale(sale)
                        if err != nil {
                                return fmt.Errorf("failed to record sale: %w", err)
                        }

                        fmt.Printf("Sale recorded successfully with ID: %d\n", id)
                        
                        // Print receipt if requested
                        if printReceipt {
                                receipt, err := handlers.GenerateReceipt(id)
                                if err != nil {
                                        return fmt.Errorf("failed to generate receipt: %w", err)
                                }
                                fmt.Println("\n" + receipt)
                        }
                        
                        // Email receipt if email provided
                        if emailReceipt && customerEmail != "" {
                                if err := handlers.EmailReceipt(id, customerEmail); err != nil {
                                        return fmt.Errorf("failed to email receipt: %w", err)
                                }
                                fmt.Printf("Receipt emailed to %s\n", customerEmail)
                        }
                        
                        return nil
                },
        }

        // Report commands
        var reportCmd = &cobra.Command{
                Use:   "report [type]",
                Short: "Generate a report",
                Long:  `Generate various reports: "sales", "inventory", "revenue", "summary", "top", "daily", "profit", "category", "trends"`,
                Args:  cobra.ExactArgs(1),
                RunE: func(cmd *cobra.Command, args []string) error {
                        // Check if user is authorized to generate reports
                        if err := auth.RequirePermission("report:generate"); err != nil {
                                return err
                        }
                        
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
                        case "profit", "profitloss", "profit-loss":
                                return generateProfitLossReport(cmd)
                        case "category", "categories":
                                return generateCategorySalesReport(cmd)
                        case "trends", "trend":
                                return generateSalesTrendsReport(cmd)
                        default:
                                return fmt.Errorf("unknown report type: %s", reportType)
                        }
                },
        }

        // Add sales-related flags to the sell command
        sellCmd.Flags().Float64("discount", 0.0, "Discount amount to apply to the sale")
        sellCmd.Flags().String("discount-code", "", "Discount code to apply")
        sellCmd.Flags().Float64("tax-rate", 8.0, "Tax rate percentage (default 8%)")
        sellCmd.Flags().String("payment-method", "cash", "Payment method (cash, card, mobile)")
        sellCmd.Flags().String("payment-ref", "", "Payment reference or transaction ID")
        sellCmd.Flags().String("email", "", "Customer email for receipt")
        sellCmd.Flags().String("phone", "", "Customer phone number")
        sellCmd.Flags().String("notes", "", "Additional notes for the sale")
        sellCmd.Flags().Bool("print-receipt", false, "Print receipt after sale")
        sellCmd.Flags().Bool("email-receipt", false, "Email receipt to customer")
        
        // Add customer loyalty related flags
        sellCmd.Flags().Int("customer-id", 0, "Customer ID for loyalty program")
        sellCmd.Flags().Bool("apply-loyalty", true, "Apply loyalty discount if eligible")
        sellCmd.Flags().Int("points-used", 0, "Loyalty points to apply to this purchase")
        sellCmd.Flags().Int("reward-id", 0, "Loyalty reward ID to redeem with this purchase")
        
        // Add report-related flags to the report command
        reportCmd.Flags().Bool("detailed", false, "Show detailed report with discount and tax information")
        reportCmd.Flags().Bool("receipts", false, "Include full receipts in the report")
        reportCmd.Flags().String("start-date", "", "Start date for report range (YYYY-MM-DD)")
        reportCmd.Flags().String("end-date", "", "End date for report range (YYYY-MM-DD)")
        reportCmd.Flags().Int("limit", 5, "Limit number of items in certain reports (like top products)")
        reportCmd.Flags().String("group-by", "day", "Group sales trends by 'day', 'week', or 'month'")

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

        // Check if detailed report is requested
        detailed, _ := cmd.Flags().GetBool("detailed")
        
        if detailed {
                // Enhanced sales report with discount, tax, and payment info
                table := tablewriter.NewWriter(cmd.OutOrStdout())
                table.SetHeader([]string{"ID", "Product", "Qty", "Subtotal", "Discount", "Tax", "Total", "Payment", "Receipt", "Date"})
                table.SetBorder(false)

                for _, s := range sales {
                        discountStr := "-"
                        if s.DiscountAmount > 0 {
                                discountStr = fmt.Sprintf("$%.2f", s.DiscountAmount)
                                if s.DiscountCode != "" {
                                        discountStr += fmt.Sprintf(" (%s)", s.DiscountCode)
                                }
                        }
                        
                        taxStr := fmt.Sprintf("$%.2f", s.TaxAmount)
                        if s.TaxRate > 0 {
                                taxStr += fmt.Sprintf(" (%.1f%%)", s.TaxRate*100)
                        }
                        
                        table.Append([]string{
                                fmt.Sprintf("%d", s.ID),
                                s.ProductName,
                                fmt.Sprintf("%d", s.Quantity),
                                fmt.Sprintf("$%.2f", s.Subtotal),
                                discountStr,
                                taxStr,
                                fmt.Sprintf("$%.2f", s.Total),
                                s.PaymentMethod,
                                s.ReceiptNumber,
                                s.SaleDate.Format("2006-01-02 15:04"),
                        })
                }

                fmt.Println("Detailed Sales Report:")
                table.Render()
        } else {
                // Basic sales report
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
        }
        
        // Add a flag to print individual receipts
        receipts, _ := cmd.Flags().GetBool("receipts")
        if receipts {
                fmt.Println("\nIndividual Receipts:")
                fmt.Println("=====================")
                
                for _, s := range sales {
                        receipt, err := handlers.GenerateReceipt(s.ID)
                        if err == nil {
                                fmt.Println(receipt)
                                fmt.Println() // Add spacing between receipts
                        }
                }
        }
        
        return nil
}

func generateInventoryReport(cmd *cobra.Command) error {
        // Try to get enhanced product details
        detailedProducts, err := db.GetAllProductsWithDetails()
        if err == nil && len(detailedProducts) > 0 {
                if len(detailedProducts) == 0 {
                        fmt.Println("No products found in inventory")
                        return nil
                }

                // Enhanced inventory report with category and supplier information
                table := tablewriter.NewWriter(cmd.OutOrStdout())
                table.SetHeader([]string{"ID", "Name", "Category", "Price", "Stock", "Value", "Supplier", "Status"})
                table.SetBorder(false)

                var totalValue float64
                for _, p := range detailedProducts {
                        value := p.Price * float64(p.Stock)
                        totalValue += value
                        
                        // Format status
                        status := "OK"
                        if p.IsLowStock {
                                status = "LOW STOCK"
                        }
                        if p.HasExpiredBatches {
                                status = "EXPIRED BATCHES"
                        }

                        table.Append([]string{
                                fmt.Sprintf("%d", p.ID),
                                p.Name,
                                p.CategoryName,
                                fmt.Sprintf("$%.2f", p.Price),
                                fmt.Sprintf("%d", p.Stock),
                                fmt.Sprintf("$%.2f", value),
                                p.SupplierName,
                                status,
                        })
                }
                
                fmt.Println("Enhanced Inventory Report:")
                table.Render()
                fmt.Printf("Total Inventory Value: $%.2f\n", totalValue)
                
                // Show inventory stats
                var lowStockCount, withBatchesCount, expiredBatchesCount int
                for _, p := range detailedProducts {
                        if p.IsLowStock {
                                lowStockCount++
                        }
                        if p.BatchCount > 0 {
                                withBatchesCount++
                        }
                        if p.HasExpiredBatches {
                                expiredBatchesCount++
                        }
                }
                
                fmt.Println("\nInventory Statistics:")
                fmt.Printf("Products with Low Stock: %d\n", lowStockCount)
                fmt.Printf("Products with Batches: %d\n", withBatchesCount)
                fmt.Printf("Products with Expired Batches: %d\n", expiredBatchesCount)
                
                return nil
        }
        
        // Fall back to basic inventory report if detailed view is not available
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
        // Get date range and limit flags
        startDate, _ := cmd.Flags().GetString("start-date")
        endDate, _ := cmd.Flags().GetString("end-date")
        limit, _ := cmd.Flags().GetInt("limit")
        
        // Ensure we have a valid limit
        if limit <= 0 {
                limit = 5 // Default to top 5
        }
        
        // Get top selling products with date range
        topProducts, err := handlers.GetTopSellingProductsDateRange(limit, startDate, endDate)
        if err != nil {
                return fmt.Errorf("failed to get top selling products: %w", err)
        }

        if len(topProducts) == 0 {
                fmt.Println("No sales data available for top products report")
                return nil
        }
        
        // Create report title based on date range
        if startDate != "" && endDate != "" {
                if startDate == endDate {
                        fmt.Printf("Top %d Products Report for %s:\n", limit, startDate)
                } else {
                        fmt.Printf("Top %d Products Report for period %s to %s:\n", limit, startDate, endDate)
                }
        } else {
                fmt.Printf("Top %d Products Report (All Time):\n", limit)
        }

        detailed, _ := cmd.Flags().GetBool("detailed")
        table := tablewriter.NewWriter(cmd.OutOrStdout())
        
        if detailed {
                table.SetHeader([]string{"Rank", "Product", "Category", "Units Sold", "Revenue", "Profit", "Margin"})
        } else {
                table.SetHeader([]string{"Rank", "Product", "Units Sold", "Revenue"})
        }
        table.SetBorder(false)

        var totalUnits int
        var totalRevenue, totalProfit float64
        for i, p := range topProducts {
                totalUnits += p.UnitsSold
                totalRevenue += p.Revenue
                totalProfit += p.Profit
                
                if detailed {
                        table.Append([]string{
                                fmt.Sprintf("%d", i+1),
                                p.ProductName,
                                p.CategoryName,
                                fmt.Sprintf("%d", p.UnitsSold),
                                fmt.Sprintf("$%.2f", p.Revenue),
                                fmt.Sprintf("$%.2f", p.Profit),
                                fmt.Sprintf("%.1f%%", p.ProfitMargin),
                        })
                } else {
                        table.Append([]string{
                                fmt.Sprintf("%d", i+1),
                                p.ProductName,
                                fmt.Sprintf("%d", p.UnitsSold),
                                fmt.Sprintf("$%.2f", p.Revenue),
                        })
                }
        }

        table.Render()
        
        // Show summary
        fmt.Printf("Total Units Sold: %d\n", totalUnits)
        fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
        
        if detailed {
                fmt.Printf("Total Profit: $%.2f\n", totalProfit)
                if totalRevenue > 0 {
                        fmt.Printf("Overall Profit Margin: %.1f%%\n", (totalProfit/totalRevenue)*100)
                }
        }
        
        return nil
}

// generateDailySalesReport generates a report of sales for today grouped by product
func generateDailySalesReport(cmd *cobra.Command) error {
        // Get date range flags
        startDate, _ := cmd.Flags().GetString("start-date")
        endDate, _ := cmd.Flags().GetString("end-date")
        
        // If no date is specified, use today
        if startDate == "" && endDate == "" {
                today := time.Now().Format("2006-01-02")
                startDate = today
                endDate = today
        }
        
        // Get sales data for the specified date range
        dailySales, err := handlers.GetSalesForDateRange(startDate, endDate)
        if err != nil {
                return fmt.Errorf("failed to get daily sales: %w", err)
        }

        if len(dailySales) == 0 {
                fmt.Printf("No sales data available for the specified date range\n")
                return nil
        }

        // Define whether to show detailed report with category and profit
        detailed, _ := cmd.Flags().GetBool("detailed")
        
        table := tablewriter.NewWriter(cmd.OutOrStdout())
        if detailed {
                table.SetHeader([]string{"Product", "Category", "Qty", "Revenue", "Cost", "Profit", "Margin"})
        } else {
                table.SetHeader([]string{"Product", "Units Sold", "Revenue"})
        }
        table.SetBorder(false)

        var totalUnits int
        var totalRevenue, totalCost, totalProfit float64

        for _, s := range dailySales {
                totalUnits += s.Quantity
                totalRevenue += s.Revenue
                totalCost += s.Cost
                totalProfit += s.Profit
                
                if detailed {
                        margin := 0.0
                        if s.Revenue > 0 {
                                margin = (s.Profit / s.Revenue) * 100
                        }
                        table.Append([]string{
                                s.ProductName,
                                s.CategoryName,
                                fmt.Sprintf("%d", s.Quantity),
                                fmt.Sprintf("$%.2f", s.Revenue),
                                fmt.Sprintf("$%.2f", s.Cost),
                                fmt.Sprintf("$%.2f", s.Profit),
                                fmt.Sprintf("%.1f%%", margin),
                        })
                } else {
                        table.Append([]string{
                                s.ProductName,
                                fmt.Sprintf("%d", s.Quantity),
                                fmt.Sprintf("$%.2f", s.Revenue),
                        })
                }
        }

        // Create report title based on date range
        if startDate == endDate {
                fmt.Printf("Sales Report for %s:\n", startDate)
        } else {
                fmt.Printf("Sales Report for period %s to %s:\n", startDate, endDate)
        }
        
        table.Render()
        
        // Show summary
        fmt.Printf("Total Units Sold: %d\n", totalUnits)
        fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
        
        // Show profit metrics in detailed view
        if detailed {
                fmt.Printf("Total Cost: $%.2f\n", totalCost)
                fmt.Printf("Total Profit: $%.2f\n", totalProfit)
                if totalRevenue > 0 {
                        fmt.Printf("Overall Profit Margin: %.1f%%\n", (totalProfit/totalRevenue)*100)
                }
        }
        
        return nil
}

// generateProfitLossReport generates a profit and loss report
func generateProfitLossReport(cmd *cobra.Command) error {
        // Get date range flags
        startDate, _ := cmd.Flags().GetString("start-date")
        endDate, _ := cmd.Flags().GetString("end-date")
        
        // Get profit and loss data
        report, err := handlers.GetProfitLossReport(startDate, endDate)
        if err != nil {
                return fmt.Errorf("failed to get profit/loss report: %w", err)
        }

        if report.Transactions == 0 {
                fmt.Println("No sales data available for profit/loss report")
                return nil
        }

        // Create report title based on date range
        if startDate != "" && endDate != "" {
                if startDate == endDate {
                        fmt.Printf("Profit & Loss Report for %s:\n", startDate)
                } else {
                        fmt.Printf("Profit & Loss Report for period %s to %s:\n", startDate, endDate)
                }
        } else {
                fmt.Println("Profit & Loss Report (All Time):")
        }
        
        fmt.Println("========================================")
        fmt.Printf("Total Revenue:         $%.2f\n", report.TotalRevenue)
        fmt.Printf("Cost of Goods Sold:    $%.2f\n", report.TotalCost)
        fmt.Printf("Gross Profit:          $%.2f\n", report.GrossProfit)
        fmt.Printf("Profit Margin:         %.1f%%\n", report.ProfitMargin)
        fmt.Println("----------------------------------------")
        fmt.Printf("Total Items Sold:      %d\n", report.TotalSold)
        fmt.Printf("Total Transactions:    %d\n", report.Transactions)
        fmt.Printf("Average Transaction:   $%.2f\n", report.AvgTransaction)
        fmt.Println("========================================")
        
        return nil
}

// generateCategorySalesReport generates a report of sales grouped by product category
func generateCategorySalesReport(cmd *cobra.Command) error {
        // Get date range flags
        startDate, _ := cmd.Flags().GetString("start-date")
        endDate, _ := cmd.Flags().GetString("end-date")
        
        // Get category sales data
        categories, err := handlers.GetSalesByCategory(startDate, endDate)
        if err != nil {
                return fmt.Errorf("failed to get category sales report: %w", err)
        }

        if len(categories) == 0 {
                fmt.Println("No category sales data available")
                return nil
        }

        // Create report title based on date range
        if startDate != "" && endDate != "" {
                if startDate == endDate {
                        fmt.Printf("Category Sales Report for %s:\n", startDate)
                } else {
                        fmt.Printf("Category Sales Report for period %s to %s:\n", startDate, endDate)
                }
        } else {
                fmt.Println("Category Sales Report (All Time):")
        }
        
        table := tablewriter.NewWriter(cmd.OutOrStdout())
        detailed, _ := cmd.Flags().GetBool("detailed")
        
        if detailed {
                table.SetHeader([]string{"Category", "Products", "Units Sold", "Revenue", "Profit", "Margin"})
        } else {
                table.SetHeader([]string{"Category", "Products", "Units Sold", "Revenue"})
        }
        table.SetBorder(false)
        
        var totalProducts int
        var totalUnits int
        var totalRevenue, totalProfit float64
        
        for _, c := range categories {
                totalProducts += c.ProductCount
                totalUnits += c.UnitsSold
                totalRevenue += c.Revenue
                totalProfit += c.Profit
                
                if detailed {
                        margin := 0.0
                        if c.Revenue > 0 {
                                margin = (c.Profit / c.Revenue) * 100
                        }
                        
                        table.Append([]string{
                                c.CategoryName,
                                fmt.Sprintf("%d", c.ProductCount),
                                fmt.Sprintf("%d", c.UnitsSold),
                                fmt.Sprintf("$%.2f", c.Revenue),
                                fmt.Sprintf("$%.2f", c.Profit),
                                fmt.Sprintf("%.1f%%", margin),
                        })
                } else {
                        table.Append([]string{
                                c.CategoryName,
                                fmt.Sprintf("%d", c.ProductCount),
                                fmt.Sprintf("%d", c.UnitsSold),
                                fmt.Sprintf("$%.2f", c.Revenue),
                        })
                }
        }
        
        table.Render()
        
        // Show summary
        fmt.Printf("Total Categories: %d\n", len(categories))
        fmt.Printf("Total Products: %d\n", totalProducts)
        fmt.Printf("Total Units Sold: %d\n", totalUnits)
        fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
        
        if detailed {
                fmt.Printf("Total Profit: $%.2f\n", totalProfit)
                if totalRevenue > 0 {
                        fmt.Printf("Overall Profit Margin: %.1f%%\n", (totalProfit/totalRevenue)*100)
                }
        }
        
        return nil
}

// generateSalesTrendsReport generates a sales trends report by time period
func generateSalesTrendsReport(cmd *cobra.Command) error {
        // Get date range flags
        startDate, _ := cmd.Flags().GetString("start-date")
        endDate, _ := cmd.Flags().GetString("end-date")
        groupBy, _ := cmd.Flags().GetString("group-by")
        
        // Validate groupBy parameter
        if groupBy != "day" && groupBy != "week" && groupBy != "month" {
                groupBy = "day" // Default to daily
        }
        
        // Get sales trends data
        trends, err := handlers.GetSalesTrends(startDate, endDate, groupBy)
        if err != nil {
                return fmt.Errorf("failed to get sales trends report: %w", err)
        }

        if len(trends) == 0 {
                fmt.Println("No sales trend data available for the specified period")
                return nil
        }

        // Create report title based on date range and grouping
        var periodText string
        switch groupBy {
        case "week":
                periodText = "Weekly"
        case "month":
                periodText = "Monthly"
        default:
                periodText = "Daily"
        }
        
        if startDate != "" && endDate != "" {
                if startDate == endDate {
                        fmt.Printf("%s Sales Trends Report for %s:\n", periodText, startDate)
                } else {
                        fmt.Printf("%s Sales Trends Report for period %s to %s:\n", periodText, startDate, endDate)
                }
        } else {
                fmt.Printf("%s Sales Trends Report (All Time):\n", periodText)
        }
        
        table := tablewriter.NewWriter(cmd.OutOrStdout())
        table.SetHeader([]string{"Period", "Transactions", "Items Sold", "Revenue", "Avg Transaction"})
        table.SetBorder(false)
        
        var totalSales int
        var totalItems int
        var totalRevenue float64
        
        for _, t := range trends {
                avgTransaction := 0.0
                if t.SaleCount > 0 {
                        avgTransaction = t.TotalRevenue / float64(t.SaleCount)
                }
                
                table.Append([]string{
                        t.Period,
                        fmt.Sprintf("%d", t.SaleCount),
                        fmt.Sprintf("%d", t.TotalItems),
                        fmt.Sprintf("$%.2f", t.TotalRevenue),
                        fmt.Sprintf("$%.2f", avgTransaction),
                })
                
                totalSales += t.SaleCount
                totalItems += t.TotalItems
                totalRevenue += t.TotalRevenue
        }
        
        table.Render()
        
        // Show summary
        fmt.Printf("Total Periods: %d\n", len(trends))
        fmt.Printf("Total Transactions: %d\n", totalSales)
        fmt.Printf("Total Items Sold: %d\n", totalItems)
        fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
        
        if totalSales > 0 {
                fmt.Printf("Overall Average Transaction: $%.2f\n", totalRevenue/float64(totalSales))
        }
        
        // Show average per period
        if len(trends) > 0 {
                avgSalesPerPeriod := float64(totalSales) / float64(len(trends))
                avgRevenuePerPeriod := totalRevenue / float64(len(trends))
                fmt.Printf("Average Transactions per %s: %.1f\n", strings.ToLower(periodText), avgSalesPerPeriod)
                fmt.Printf("Average Revenue per %s: $%.2f\n", strings.ToLower(periodText), avgRevenuePerPeriod)
        }
        
        return nil
}
