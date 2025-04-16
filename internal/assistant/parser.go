package assistant

import (
        "fmt"
        "regexp"
        "strconv"
        "strings"

        "termpos/internal/handlers"
        "termpos/internal/models"
)

// Command types recognized by the parser
const (
        CommandAdd         = "add"
        CommandSell        = "sell"
        CommandInventory   = "inventory"
        CommandReport      = "report"
        CommandUpdateStock = "update-stock"
)

// ProcessNaturalLanguage parses and processes natural language input
func ProcessNaturalLanguage(input string) (string, error) {
        // For backwards compatibility, route to the new function that uses context
        return ProcessNaturalLanguageWithContext(input)
}

// handleAddCommand processes "add" commands
func handleAddCommand(input string) (string, error) {
        // Regex for: "add X product at $Y" or "add X product for $Y" or "add X product at Y dollars"
        addPattern := regexp.MustCompile(`add\s+(\d+)?\s*([a-zA-Z\s]+)\s+(at|for)\s+\$?(\d+(\.\d+)?)`)
        matches := addPattern.FindStringSubmatch(input)

        if len(matches) < 5 {
                // Try a simpler pattern: "add product name price stock"
                parts := strings.Fields(input)
                if len(parts) >= 3 && parts[0] == "add" {
                        nameEnd := len(parts) - 1
                        stock := 0
                        
                        // Check if last parameter is stock
                        lastNum, err := strconv.Atoi(parts[len(parts)-1])
                        if err == nil {
                                stock = lastNum
                                nameEnd--
                        }
                        
                        // Check if second-to-last parameter is price
                        price, err := strconv.ParseFloat(strings.TrimPrefix(parts[nameEnd], "$"), 64)
                        if err != nil {
                                return "", fmt.Errorf("invalid price format: %s", parts[nameEnd])
                        }
                        nameEnd--
                        
                        // Everything in between is the product name
                        name := strings.Join(parts[1:nameEnd+1], " ")
                        
                        product := models.Product{
                                Name:  name,
                                Price: price,
                                Stock: stock,
                        }
                        
                        id, err := handlers.AddProduct(product)
                        if err != nil {
                                return "", err
                        }
                        
                        return fmt.Sprintf("Added %s at $%.2f with stock %d (ID: %d)", name, price, stock, id), nil
                }
                
                return "", fmt.Errorf("invalid add command format. Try 'add coffee at $3.50' or 'add 10 mugs at $5'")
        }

        // Extract the quantity, product name and price
        quantity := 1
        if matches[1] != "" {
                var err error
                quantity, err = strconv.Atoi(matches[1])
                if err != nil {
                        return "", fmt.Errorf("invalid quantity: %s", matches[1])
                }
        }

        name := strings.TrimSpace(matches[2])
        price, err := strconv.ParseFloat(matches[4], 64)
        if err != nil {
                return "", fmt.Errorf("invalid price: %s", matches[4])
        }

        // Create the product
        product := models.Product{
                Name:  name,
                Price: price,
                Stock: quantity,
        }

        id, err := handlers.AddProduct(product)
        if err != nil {
                return "", err
        }

        return fmt.Sprintf("Added %s at $%.2f with stock %d (ID: %d)", name, price, quantity, id), nil
}

// handleSellCommand processes "sell" commands
func handleSellCommand(input string) (string, error) {
        // Try to match pattern: "sell X of product Y" or "sell X product" or "sell product"
        var productName string
        var quantity int = 1

        // First try to extract product ID directly
        idPattern := regexp.MustCompile(`sell\s+(\d+)\s+of\s+product\s+(\d+)`)
        idMatches := idPattern.FindStringSubmatch(input)
        if len(idMatches) >= 3 {
                var err error
                quantity, err = strconv.Atoi(idMatches[1])
                if err != nil {
                        return "", fmt.Errorf("invalid quantity: %s", idMatches[1])
                }

                productID, err := strconv.Atoi(idMatches[2])
                if err != nil {
                        return "", fmt.Errorf("invalid product ID: %s", idMatches[2])
                }

                // Record the sale
                sale := models.Sale{
                        ProductID: productID,
                        Quantity:  quantity,
                }

                id, err := handlers.RecordSale(sale)
                if err != nil {
                        return "", err
                }

                product, err := handlers.GetProductByID(productID)
                if err != nil {
                        return "", err
                }

                return fmt.Sprintf("Sold %d of %s for $%.2f (Sale ID: %d)", quantity, product.Name, product.Price*float64(quantity), id), nil
        }

        // Try to match by product name
        namePattern := regexp.MustCompile(`sell\s+(\d+)?\s*(of|units|items)?\s*([a-zA-Z\s]+)`)
        nameMatches := namePattern.FindStringSubmatch(input)
        if len(nameMatches) >= 4 {
                if nameMatches[1] != "" {
                        var err error
                        quantity, err = strconv.Atoi(nameMatches[1])
                        if err != nil {
                                return "", fmt.Errorf("invalid quantity: %s", nameMatches[1])
                        }
                }

                productName = strings.TrimSpace(nameMatches[3])
        } else {
                // Try simpler pattern: "sell product"
                parts := strings.Fields(input)
                if len(parts) < 2 {
                        return "", fmt.Errorf("invalid sell command format. Try 'sell 2 coffee' or 'sell mug'")
                }
                
                if len(parts) > 2 {
                        // Check if the second parameter is a number (quantity)
                        if q, err := strconv.Atoi(parts[1]); err == nil {
                                quantity = q
                                productName = strings.Join(parts[2:], " ")
                        } else {
                                productName = strings.Join(parts[1:], " ")
                        }
                } else {
                        productName = parts[1]
                }
        }

        // Get all products and find a match
        products, err := handlers.GetAllProducts()
        if err != nil {
                return "", err
        }

        var matchedProduct models.Product
        for _, p := range products {
                if strings.Contains(strings.ToLower(p.Name), strings.ToLower(productName)) {
                        matchedProduct = p
                        break
                }
        }

        if matchedProduct.ID == 0 {
                return "", fmt.Errorf("no product found matching '%s'", productName)
        }

        // Record the sale
        sale := models.Sale{
                ProductID: matchedProduct.ID,
                Quantity:  quantity,
        }

        id, err := handlers.RecordSale(sale)
        if err != nil {
                return "", err
        }

        return fmt.Sprintf("Sold %d of %s for $%.2f (Sale ID: %d)", quantity, matchedProduct.Name, matchedProduct.Price*float64(quantity), id), nil
}

// handleInventoryCommand processes inventory commands
func handleInventoryCommand(input string) (string, error) {
        products, err := handlers.GetAllProducts()
        if err != nil {
                return "", err
        }

        if len(products) == 0 {
                return "Inventory is empty. Add products with the 'add' command.", nil
        }

        // Generate the response
        sb := strings.Builder{}
        sb.WriteString("Current Inventory:\n")
        sb.WriteString("------------------\n")
        
        for _, p := range products {
                sb.WriteString(fmt.Sprintf("ID: %d | %s | Price: $%.2f | Stock: %d\n", 
                        p.ID, p.Name, p.Price, p.Stock))
        }

        return sb.String(), nil
}

// handleUpdateStockCommand processes update stock commands
func handleUpdateStockCommand(input string) (string, error) {
        // Try to match: "update stock of product X to Y" or "set stock of X to Y"
        idPattern := regexp.MustCompile(`(update|set|change)\s+stock\s+(of|for)\s+product\s+(\d+)\s+to\s+(\d+)`)
        idMatches := idPattern.FindStringSubmatch(input)
        
        if len(idMatches) >= 5 {
                productID, err := strconv.Atoi(idMatches[3])
                if err != nil {
                        return "", fmt.Errorf("invalid product ID: %s", idMatches[3])
                }

                stock, err := strconv.Atoi(idMatches[4])
                if err != nil {
                        return "", fmt.Errorf("invalid stock quantity: %s", idMatches[4])
                }

                if err := handlers.UpdateProductStock(productID, stock); err != nil {
                        return "", err
                }

                product, err := handlers.GetProductByID(productID)
                if err != nil {
                        return "", err
                }

                return fmt.Sprintf("Updated stock of %s to %d", product.Name, stock), nil
        }

        // Try to match by product name: "update stock of X to Y"
        namePattern := regexp.MustCompile(`(update|set|change)\s+stock\s+(of|for)\s+([a-zA-Z\s]+)\s+to\s+(\d+)`)
        nameMatches := namePattern.FindStringSubmatch(input)
        
        if len(nameMatches) >= 5 {
                productName := strings.TrimSpace(nameMatches[3])
                stock, err := strconv.Atoi(nameMatches[4])
                if err != nil {
                        return "", fmt.Errorf("invalid stock quantity: %s", nameMatches[4])
                }

                // Get all products and find a match
                products, err := handlers.GetAllProducts()
                if err != nil {
                        return "", err
                }

                var matchedProduct models.Product
                for _, p := range products {
                        if strings.Contains(strings.ToLower(p.Name), strings.ToLower(productName)) {
                                matchedProduct = p
                                break
                        }
                }

                if matchedProduct.ID == 0 {
                        return "", fmt.Errorf("no product found matching '%s'", productName)
                }

                if err := handlers.UpdateProductStock(matchedProduct.ID, stock); err != nil {
                        return "", err
                }

                return fmt.Sprintf("Updated stock of %s to %d", matchedProduct.Name, stock), nil
        }

        return "", fmt.Errorf("invalid update stock command format. Try 'update stock of coffee to 20'")
}

// handleReportCommand processes report commands
func handleReportCommand(input string) (string, error) {
        if strings.Contains(input, "sales") {
                return generateSalesReport()
        } else if strings.Contains(input, "inventory") {
                return generateInventoryReport()
        } else if strings.Contains(input, "revenue") {
                return generateRevenueReport()
        } else {
                // Default to revenue report
                return generateRevenueReport()
        }
}

// generateSalesReport generates a sales report
func generateSalesReport() (string, error) {
        sales, err := handlers.GetAllSales()
        if err != nil {
                return "", err
        }

        if len(sales) == 0 {
                return "No sales data available.", nil
        }

        sb := strings.Builder{}
        sb.WriteString("Sales Report:\n")
        sb.WriteString("-------------\n")
        
        totalSales := 0.0
        for _, s := range sales {
                sb.WriteString(fmt.Sprintf("ID: %d | %s | Quantity: %d | Total: $%.2f | Date: %s\n", 
                        s.ID, s.ProductName, s.Quantity, s.Total, s.SaleDate.Format("2006-01-02 15:04:05")))
                totalSales += s.Total
        }
        
        sb.WriteString(fmt.Sprintf("\nTotal Sales: $%.2f\n", totalSales))

        return sb.String(), nil
}

// generateInventoryReport generates an inventory report
func generateInventoryReport() (string, error) {
        products, err := handlers.GetAllProducts()
        if err != nil {
                return "", err
        }

        if len(products) == 0 {
                return "Inventory is empty.", nil
        }

        sb := strings.Builder{}
        sb.WriteString("Inventory Report:\n")
        sb.WriteString("-----------------\n")
        
        totalValue := 0.0
        for _, p := range products {
                value := p.Price * float64(p.Stock)
                totalValue += value
                sb.WriteString(fmt.Sprintf("ID: %d | %s | Price: $%.2f | Stock: %d | Value: $%.2f\n", 
                        p.ID, p.Name, p.Price, p.Stock, value))
        }
        
        sb.WriteString(fmt.Sprintf("\nTotal Inventory Value: $%.2f\n", totalValue))

        return sb.String(), nil
}

// generateRevenueReport generates a revenue report
func generateRevenueReport() (string, error) {
        revenue, err := handlers.GetRevenueReport()
        if err != nil {
                return "", err
        }

        if len(revenue) == 0 {
                return "No revenue data available.", nil
        }

        sb := strings.Builder{}
        sb.WriteString("Revenue Report:\n")
        sb.WriteString("---------------\n")
        
        totalRevenue := 0.0
        for _, r := range revenue {
                totalRevenue += r.Revenue
                sb.WriteString(fmt.Sprintf("%s | Units Sold: %d | Revenue: $%.2f\n", 
                        r.ProductName, r.UnitsSold, r.Revenue))
        }
        
        sb.WriteString(fmt.Sprintf("\nTotal Revenue: $%.2f\n", totalRevenue))

        return sb.String(), nil
}

// getHelpText returns help information for the assistant
func getHelpText() string {
        return `
TermPOS Assistant Help:
======================

Product Management:
- Add products: "add coffee at $3.50" or "add 10 mugs at $5"
- View inventory: "show inventory" or "list products"
- Update stock: "update stock of coffee to 20" or "set stock of product 1 to 15"

Sales:
- Record sales: "sell 2 coffees" or "sell mug" or "sell 3 of product 2"

Reports:
- Sales report: "show sales report" or "view sales"
- Inventory report: "show inventory report" 
- Revenue report: "show revenue report" or "revenue"

General:
- Help: "help" or "show commands"
`
}
