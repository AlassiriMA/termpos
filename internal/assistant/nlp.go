package assistant

import (
        "fmt"
        "regexp"
        "strconv"
        "strings"
        "time"

        "github.com/sahilm/fuzzy"
        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/models"
)

// Intent represents the user's intention
type Intent string

const (
        IntentUnknown      Intent = "unknown"
        IntentAddProduct   Intent = "add_product"
        IntentSellProduct  Intent = "sell_product"
        IntentGetInventory Intent = "get_inventory"
        IntentUpdateStock  Intent = "update_stock"
        IntentGetReport    Intent = "get_report"
        IntentHelp         Intent = "help"
        IntentFeedback     Intent = "feedback"
        IntentGreeting     Intent = "greeting"
        IntentThanks       Intent = "thanks"
)

// Entity represents extracted parameters from user input
type Entity struct {
        Type  string
        Value interface{}
}

// ConversationContext maintains the state of the conversation
type ConversationContext struct {
        PreviousIntent       Intent
        Entities             map[string]Entity
        ConversationHistory  []string
        LastProductsAccessed []models.Product
        LastTime             time.Time
}

// Global conversation context
var context = ConversationContext{
        PreviousIntent:      IntentUnknown,
        Entities:            make(map[string]Entity),
        ConversationHistory: []string{},
        LastTime:            time.Now(),
}

// ProcessNaturalLanguageWithContext uses context to improve NLP understanding
func ProcessNaturalLanguageWithContext(input string) (string, error) {
        // Trim and normalize input
        input = strings.TrimSpace(input)
        normalizedInput := strings.ToLower(input)
        
        // Add to conversation history
        context.ConversationHistory = append(context.ConversationHistory, input)
        if len(context.ConversationHistory) > 10 {
                // Keep only the last 10 interactions
                context.ConversationHistory = context.ConversationHistory[1:]
        }
        
        // Detect intent
        intent := detectIntent(normalizedInput)
        
        // Extract entities based on intent
        entities := extractEntities(normalizedInput, intent)
        
        // Update context with extracted entities
        for k, v := range entities {
                context.Entities[k] = v
        }
        
        // Generate response based on intent and entities
        response, err := generateResponse(intent, entities)
        if err != nil {
                // If we got an error, try to recover with context if possible
                if recoveryResponse, recoveryErr := tryToRecoverWithContext(intent, err); recoveryErr == nil {
                        return recoveryResponse, nil
                }
                
                // If error is about unknown command, suggest alternatives
                if strings.Contains(err.Error(), "not sure what you want") {
                        suggestions := suggestCommands(normalizedInput)
                        if len(suggestions) > 0 {
                                return fmt.Sprintf("I'm not sure what you want to do. Did you mean:\n%s\n\nTry 'help' for instructions.", 
                                        strings.Join(suggestions, "\n")), nil
                        }
                }
                
                return "", err
        }
        
        // Update previous intent for next interaction
        context.PreviousIntent = intent
        context.LastTime = time.Now()
        
        return response, nil
}

// detectIntent determines the user's intention from the input
func detectIntent(input string) Intent {
        // Check for greeting patterns
        greetingPatterns := []string{"hello", "hi ", "hey", "greetings", "good morning", "good afternoon", "good evening", "howdy"}
        for _, pattern := range greetingPatterns {
                if strings.Contains(input, pattern) {
                        return IntentGreeting
                }
        }
        
        // Check for thanks patterns
        thanksPatterns := []string{"thank", "thanks", "thx", "appreciate"}
        for _, pattern := range thanksPatterns {
                if strings.Contains(input, pattern) {
                        return IntentThanks
                }
        }
        
        // Check for feedback patterns
        feedbackPatterns := []string{"feedback", "suggestion", "improve", "issue", "problem", "bug", "feature", "request"}
        for _, pattern := range feedbackPatterns {
                if strings.Contains(input, pattern) {
                        return IntentFeedback
                }
        }
        
        // Check for help intent
        if strings.Contains(input, "help") || strings.Contains(input, "commands") || 
           strings.Contains(input, "instructions") || strings.Contains(input, "guide") {
                return IntentHelp
        }
        
        // Check for product addition intent
        if strings.HasPrefix(input, "add") || 
           strings.Contains(input, "new product") || 
           strings.Contains(input, "create product") ||
           (strings.Contains(input, "product") && strings.Contains(input, "add")) {
                return IntentAddProduct
        }
        
        // Check for sell intent
        if strings.HasPrefix(input, "sell") || 
           strings.Contains(input, "purchase") || 
           strings.Contains(input, "buy") ||
           strings.Contains(input, "order") {
                return IntentSellProduct
        }
        
        // Check for inventory intent
        if strings.Contains(input, "inventory") || 
           strings.Contains(input, "stock") || 
           (strings.Contains(input, "list") && strings.Contains(input, "product")) || 
           strings.Contains(input, "products") {
                
                // Further differentiate between viewing inventory and updating stock
                if strings.Contains(input, "update") || 
                   strings.Contains(input, "change") || 
                   strings.Contains(input, "set") ||
                   strings.Contains(input, "modify") {
                        return IntentUpdateStock
                }
                
                return IntentGetInventory
        }
        
        // Check for report intent
        if strings.Contains(input, "report") || 
           strings.Contains(input, "sales") || 
           strings.Contains(input, "revenue") ||
           strings.Contains(input, "summary") ||
           strings.Contains(input, "analytics") {
                return IntentGetReport
        }
        
        // If context suggests a previous intent, and this message is short or ambiguous,
        // it might be a continuation of the previous conversation
        if context.PreviousIntent != IntentUnknown && 
           (len(strings.Fields(input)) <= 3 || strings.HasPrefix(input, "yes") || strings.HasPrefix(input, "no")) {
                return context.PreviousIntent
        }
        
        return IntentUnknown
}

// extractEntities pulls relevant information from the user input based on intent
func extractEntities(input string, intent Intent) map[string]Entity {
        entities := make(map[string]Entity)
        
        switch intent {
        case IntentAddProduct:
                extractAddProductEntities(input, entities)
        case IntentSellProduct:
                extractSellProductEntities(input, entities)
        case IntentUpdateStock:
                extractUpdateStockEntities(input, entities)
        case IntentGetReport:
                extractReportEntities(input, entities)
        }
        
        return entities
}

// extractAddProductEntities extracts information for adding a product
func extractAddProductEntities(input string, entities map[string]Entity) {
        // Regex for: "add X product at $Y" or "add X product for $Y" or "add X product at Y dollars"
        addPattern := regexp.MustCompile(`add\s+(\d+)?\s*([a-zA-Z\s]+)\s+(at|for)\s+\$?(\d+(\.\d+)?)`)
        matches := addPattern.FindStringSubmatch(input)
        
        if len(matches) >= 5 {
                // Extract the quantity, product name and price
                quantity := 1
                if matches[1] != "" {
                        if q, err := strconv.Atoi(matches[1]); err == nil {
                                quantity = q
                        }
                }
                
                name := strings.TrimSpace(matches[2])
                price, _ := strconv.ParseFloat(matches[4], 64)
                
                entities["product_name"] = Entity{Type: "string", Value: name}
                entities["product_price"] = Entity{Type: "float", Value: price}
                entities["product_quantity"] = Entity{Type: "int", Value: quantity}
                return
        }
        
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
                if err == nil {
                        nameEnd--
                        
                        // Everything in between is the product name
                        name := strings.Join(parts[1:nameEnd+1], " ")
                        
                        entities["product_name"] = Entity{Type: "string", Value: name}
                        entities["product_price"] = Entity{Type: "float", Value: price}
                        entities["product_quantity"] = Entity{Type: "int", Value: stock}
                }
        }
}

// extractSellProductEntities extracts information for selling a product
func extractSellProductEntities(input string, entities map[string]Entity) {
        // First try to extract product ID directly
        idPattern := regexp.MustCompile(`sell\s+(\d+)\s+of\s+product\s+(\d+)`)
        idMatches := idPattern.FindStringSubmatch(input)
        
        if len(idMatches) >= 3 {
                quantity, _ := strconv.Atoi(idMatches[1])
                productID, _ := strconv.Atoi(idMatches[2])
                
                entities["product_id"] = Entity{Type: "int", Value: productID}
                entities["quantity"] = Entity{Type: "int", Value: quantity}
                return
        }
        
        // Try to match by product name
        namePattern := regexp.MustCompile(`sell\s+(\d+)?\s*(of|units|items)?\s*([a-zA-Z\s]+)`)
        nameMatches := namePattern.FindStringSubmatch(input)
        
        if len(nameMatches) >= 4 {
                quantity := 1
                if nameMatches[1] != "" {
                        if q, err := strconv.Atoi(nameMatches[1]); err == nil {
                                quantity = q
                        }
                }
                
                productName := strings.TrimSpace(nameMatches[3])
                entities["product_name"] = Entity{Type: "string", Value: productName}
                entities["quantity"] = Entity{Type: "int", Value: quantity}
                return
        }
        
        // Try simpler pattern: "sell product"
        parts := strings.Fields(input)
        if len(parts) >= 2 {
                quantity := 1
                var productName string
                
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
                
                entities["product_name"] = Entity{Type: "string", Value: productName}
                entities["quantity"] = Entity{Type: "int", Value: quantity}
        }
}

// extractUpdateStockEntities extracts information for updating stock
func extractUpdateStockEntities(input string, entities map[string]Entity) {
        // Try to match: "update stock of product X to Y" or "set stock of X to Y"
        idPattern := regexp.MustCompile(`(update|set|change)\s+stock\s+(of|for)\s+product\s+(\d+)\s+to\s+(\d+)`)
        idMatches := idPattern.FindStringSubmatch(input)
        
        if len(idMatches) >= 5 {
                productID, _ := strconv.Atoi(idMatches[3])
                stock, _ := strconv.Atoi(idMatches[4])
                
                entities["product_id"] = Entity{Type: "int", Value: productID}
                entities["stock"] = Entity{Type: "int", Value: stock}
                return
        }
        
        // Try to match by product name: "update stock of X to Y"
        namePattern := regexp.MustCompile(`(update|set|change)\s+stock\s+(of|for)\s+([a-zA-Z\s]+)\s+to\s+(\d+)`)
        nameMatches := namePattern.FindStringSubmatch(input)
        
        if len(nameMatches) >= 5 {
                productName := strings.TrimSpace(nameMatches[3])
                stock, _ := strconv.Atoi(nameMatches[4])
                
                entities["product_name"] = Entity{Type: "string", Value: productName}
                entities["stock"] = Entity{Type: "int", Value: stock}
        }
}

// extractReportEntities extracts information for generating reports
func extractReportEntities(input string, entities map[string]Entity) {
        if strings.Contains(input, "sales") || strings.Contains(input, "transaction") {
                entities["report_type"] = Entity{Type: "string", Value: "sales"}
        } else if strings.Contains(input, "inventory") || strings.Contains(input, "stock") {
                entities["report_type"] = Entity{Type: "string", Value: "inventory"}
        } else if strings.Contains(input, "revenue") || strings.Contains(input, "profit") {
                entities["report_type"] = Entity{Type: "string", Value: "revenue"}
        } else if strings.Contains(input, "daily") || strings.Contains(input, "today") {
                entities["report_type"] = Entity{Type: "string", Value: "daily"}
        } else if strings.Contains(input, "top") || strings.Contains(input, "best") {
                entities["report_type"] = Entity{Type: "string", Value: "top"}
        } else {
                // Default to "summary" report
                entities["report_type"] = Entity{Type: "string", Value: "summary"}
        }
}

// generateResponse creates a response based on intent and entities
func generateResponse(intent Intent, entities map[string]Entity) (string, error) {
        switch intent {
        case IntentGreeting:
                return "Hello! I'm the TermPOS assistant. How can I help you with your point of sale needs today?", nil
        
        case IntentThanks:
                return "You're welcome! Is there anything else I can help you with?", nil
        
        case IntentFeedback:
                return "Thank you for your feedback! I'll make note of that to improve the system.", nil
        
        case IntentHelp:
                return getHelpText(), nil
        
        case IntentAddProduct:
                return handleAddProduct(entities)
        
        case IntentSellProduct:
                return handleSellProduct(entities)
        
        case IntentGetInventory:
                return handleGetInventory()
        
        case IntentUpdateStock:
                return handleUpdateStock(entities)
        
        case IntentGetReport:
                return handleGetReport(entities)
        
        default:
                return "", fmt.Errorf("I'm not sure what you want to do. Try 'help' for instructions")
        }
}

// handleAddProduct processes the add product intent
func handleAddProduct(entities map[string]Entity) (string, error) {
        // Check permissions - only admin and manager can add products
        if err := auth.RequirePermission("product:manage"); err != nil {
                return "", fmt.Errorf("you don't have permission to add products: %w", err)
        }

        var productName string
        var productPrice float64
        var productQuantity int
        
        // Extract product name
        if nameEntity, ok := entities["product_name"]; ok {
                if name, ok := nameEntity.Value.(string); ok {
                        productName = name
                } else {
                        return "", fmt.Errorf("invalid product name")
                }
        } else {
                return "", fmt.Errorf("product name is required")
        }
        
        // Extract product price
        if priceEntity, ok := entities["product_price"]; ok {
                if price, ok := priceEntity.Value.(float64); ok {
                        productPrice = price
                } else {
                        return "", fmt.Errorf("invalid product price")
                }
        } else {
                return "", fmt.Errorf("product price is required")
        }
        
        // Extract product quantity (optional)
        if quantityEntity, ok := entities["product_quantity"]; ok {
                if quantity, ok := quantityEntity.Value.(int); ok {
                        productQuantity = quantity
                }
        }
        
        // Create the product
        product := models.Product{
                Name:  productName,
                Price: productPrice,
                Stock: productQuantity,
        }
        
        id, err := db.AddProduct(product)
        if err != nil {
                return "", err
        }
        
        return fmt.Sprintf("Added %s at $%.2f with stock %d (ID: %d)", productName, productPrice, productQuantity, id), nil
}

// handleSellProduct processes the sell product intent
func handleSellProduct(entities map[string]Entity) (string, error) {
        // Check permissions - anyone with sales:create can sell products
        if err := auth.RequirePermission("sales:create"); err != nil {
                return "", fmt.Errorf("you don't have permission to record sales: %w", err)
        }
        
        var productID int
        var productName string
        var quantity int = 1
        
        // Extract product ID
        if idEntity, ok := entities["product_id"]; ok {
                if id, ok := idEntity.Value.(int); ok {
                        productID = id
                }
        }
        
        // Extract product name if ID not provided
        if productID == 0 {
                if nameEntity, ok := entities["product_name"]; ok {
                        if name, ok := nameEntity.Value.(string); ok {
                                productName = name
                                
                                // Find product by name with fuzzy matching
                                matchedProduct, err := findProductByName(productName)
                                if err != nil {
                                        return "", err
                                }
                                
                                productID = matchedProduct.ID
                                
                                // Cache the product for context
                                context.LastProductsAccessed = []models.Product{matchedProduct}
                        }
                }
        }
        
        // Validate product ID
        if productID == 0 {
                return "", fmt.Errorf("could not determine which product to sell")
        }
        
        // Extract quantity
        if quantityEntity, ok := entities["quantity"]; ok {
                if q, ok := quantityEntity.Value.(int); ok {
                        quantity = q
                }
        }
        
        // Validate quantity
        if quantity <= 0 {
                return "", fmt.Errorf("quantity must be greater than zero")
        }
        
        // Record the sale
        sale := models.Sale{
                ProductID: productID,
                Quantity:  quantity,
        }
        
        id, err := db.RecordSale(sale)
        if err != nil {
                return "", err
        }
        
        // Get product info for feedback
        product, err := db.GetProductByID(productID)
        if err != nil {
                return "", err
        }
        
        return fmt.Sprintf("Sold %d of %s for $%.2f (Sale ID: %d)", quantity, product.Name, product.Price*float64(quantity), id), nil
}

// handleGetInventory processes the get inventory intent
func handleGetInventory() (string, error) {
        // Check permissions - anyone with inventory:view can view inventory
        if err := auth.RequirePermission("inventory:view"); err != nil {
                return "", fmt.Errorf("you don't have permission to view inventory: %w", err)
        }
        products, err := db.GetAllProducts()
        if err != nil {
                return "", err
        }
        
        if len(products) == 0 {
                return "Inventory is empty. Add products with the 'add' command.", nil
        }
        
        // Update context with products
        context.LastProductsAccessed = products
        
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

// handleUpdateStock processes the update stock intent
func handleUpdateStock(entities map[string]Entity) (string, error) {
        // Check permissions - only admin and manager can update stock
        if err := auth.RequirePermission("product:manage"); err != nil {
                return "", fmt.Errorf("you don't have permission to update product stock: %w", err)
        }
        var productID int
        var productName string
        var stock int
        
        // Extract product ID
        if idEntity, ok := entities["product_id"]; ok {
                if id, ok := idEntity.Value.(int); ok {
                        productID = id
                }
        }
        
        // Extract product name if ID not provided
        if productID == 0 {
                if nameEntity, ok := entities["product_name"]; ok {
                        if name, ok := nameEntity.Value.(string); ok {
                                productName = name
                                
                                // Find product by name with fuzzy matching
                                matchedProduct, err := findProductByName(productName)
                                if err != nil {
                                        return "", err
                                }
                                
                                productID = matchedProduct.ID
                                
                                // Cache the product for context
                                context.LastProductsAccessed = []models.Product{matchedProduct}
                        }
                }
        }
        
        // Validate product ID
        if productID == 0 {
                return "", fmt.Errorf("could not determine which product to update")
        }
        
        // Extract stock
        if stockEntity, ok := entities["stock"]; ok {
                if s, ok := stockEntity.Value.(int); ok {
                        stock = s
                }
        }
        
        // Validate stock
        if stock < 0 {
                return "", fmt.Errorf("stock cannot be negative")
        }
        
        if err := db.UpdateProductStock(productID, stock); err != nil {
                return "", err
        }
        
        product, err := db.GetProductByID(productID)
        if err != nil {
                return "", err
        }
        
        return fmt.Sprintf("Updated stock of %s to %d", product.Name, stock), nil
}

// handleGetReport processes the get report intent
func handleGetReport(entities map[string]Entity) (string, error) {
        // Check permissions - only admin and manager can generate reports
        if err := auth.RequirePermission("report:generate"); err != nil {
                return "", fmt.Errorf("you don't have permission to generate reports: %w", err)
        }
        var reportType string
        
        // Extract report type
        if typeEntity, ok := entities["report_type"]; ok {
                if rType, ok := typeEntity.Value.(string); ok {
                        reportType = rType
                }
        }
        
        switch reportType {
        case "sales":
                return generateSalesReport()
        case "inventory":
                return generateInventoryReport()
        case "revenue":
                return generateRevenueReport()
        case "daily":
                sales, err := db.GetDailySales()
                if err != nil {
                        return "", err
                }
                
                if len(sales) == 0 {
                        return "No sales recorded today.", nil
                }
                
                // Format the daily sales report
                sb := strings.Builder{}
                sb.WriteString(fmt.Sprintf("Daily Sales Report for %s:\n", time.Now().Format("2006-01-02")))
                
                totalUnits := 0
                totalRevenue := 0.0
                
                for _, sale := range sales {
                        sb.WriteString(fmt.Sprintf("  %s | Units Sold: %d | Revenue: $%.2f\n", 
                                sale.ProductName, sale.Quantity, sale.Revenue))
                        totalUnits += sale.Quantity
                        totalRevenue += sale.Revenue
                }
                
                sb.WriteString(fmt.Sprintf("Total Units Sold Today: %d\n", totalUnits))
                sb.WriteString(fmt.Sprintf("Total Revenue Today: $%.2f\n", totalRevenue))
                
                return sb.String(), nil
                
        case "top":
                topProducts, err := db.GetTopSellingProducts(5)
                if err != nil {
                        return "", err
                }
                
                if len(topProducts) == 0 {
                        return "No sales data available.", nil
                }
                
                // Format the top selling products report
                sb := strings.Builder{}
                sb.WriteString("Top Selling Products Report:\n")
                
                for i, product := range topProducts {
                        sb.WriteString(fmt.Sprintf("  %d. %s | Units Sold: %d | Revenue: $%.2f\n", 
                                i+1, product.ProductName, product.Quantity, product.Revenue))
                }
                
                return sb.String(), nil
                
        default: // "summary"
                summary, err := db.GetSalesSummary()
                if err != nil {
                        return "", err
                }
                
                // Format the summary report
                sb := strings.Builder{}
                sb.WriteString("Sales Summary Report:\n")
                sb.WriteString("--------------------\n")
                sb.WriteString(fmt.Sprintf("Total Revenue: $%.2f\n", summary.TotalRevenue))
                sb.WriteString(fmt.Sprintf("Total Items Sold: %d\n", summary.TotalItemsSold))
                sb.WriteString(fmt.Sprintf("Total Transactions: %d\n", summary.TotalTransactions))
                
                // Calculate average transaction value
                avgTransaction := 0.0
                if summary.TotalTransactions > 0 {
                        avgTransaction = summary.TotalRevenue / float64(summary.TotalTransactions)
                }
                sb.WriteString(fmt.Sprintf("Average Transaction Value: $%.2f\n", avgTransaction))
                
                return sb.String(), nil
        }
}

// tryToRecoverWithContext attempts to use context to handle errors
func tryToRecoverWithContext(intent Intent, originalError error) (string, error) {
        errMsg := originalError.Error()
        
        // Product not found error handling with fuzzy search
        if strings.Contains(errMsg, "no product found") || 
           strings.Contains(errMsg, "product not found") {
                
                // Get all products for fuzzy search
                products, err := db.GetAllProducts()
                if err != nil {
                        return "", originalError
                }
                
                // Extract the name that wasn't found
                nameStart := strings.Index(errMsg, "'")
                nameEnd := strings.LastIndex(errMsg, "'")
                if nameStart >= 0 && nameEnd > nameStart {
                        searchName := errMsg[nameStart+1:nameEnd]
                        
                        // Get a list of all product names
                        var productNames []string
                        for _, p := range products {
                                productNames = append(productNames, p.Name)
                        }
                        
                        // Perform fuzzy search
                        matches := fuzzy.Find(searchName, productNames)
                        if len(matches) > 0 {
                                suggestions := make([]string, 0, len(matches))
                                for _, match := range matches[:min(3, len(matches))] {
                                        suggestions = append(suggestions, fmt.Sprintf("- %s", match.Str))
                                }
                                
                                return fmt.Sprintf("Product '%s' not found. Did you mean one of these?\n%s", 
                                        searchName, strings.Join(suggestions, "\n")), nil
                        }
                }
        }
        
        // If there's context about products, suggest them
        if len(context.LastProductsAccessed) > 0 && 
           (intent == IntentSellProduct || intent == IntentUpdateStock) {
                suggestions := make([]string, 0, len(context.LastProductsAccessed))
                for _, p := range context.LastProductsAccessed {
                        suggestions = append(suggestions, fmt.Sprintf("- %s (ID: %d, Price: $%.2f, Stock: %d)", 
                                p.Name, p.ID, p.Price, p.Stock))
                }
                
                return fmt.Sprintf("I'm having trouble understanding which product you mean. Here are some products in inventory:\n%s", 
                        strings.Join(suggestions, "\n")), nil
        }
        
        return "", originalError
}

// findProductByName finds a product by name with fuzzy matching
func findProductByName(name string) (models.Product, error) {
        // Get all products
        products, err := db.GetAllProducts()
        if err != nil {
                return models.Product{}, err
        }
        
        // First try exact match
        for _, p := range products {
                if strings.EqualFold(p.Name, name) {
                        return p, nil
                }
        }
        
        // Then try contains match
        for _, p := range products {
                if strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
                        return p, nil
                }
        }
        
        // Finally try fuzzy match
        var productNames []string
        nameToProduct := make(map[string]models.Product)
        
        for _, p := range products {
                productNames = append(productNames, p.Name)
                nameToProduct[p.Name] = p
        }
        
        matches := fuzzy.Find(name, productNames)
        if len(matches) > 0 {
                return nameToProduct[matches[0].Str], nil
        }
        
        return models.Product{}, fmt.Errorf("no product found matching '%s'", name)
}

// suggestCommands suggests possible commands based on user input
func suggestCommands(input string) []string {
        commands := []string{
                "add coffee at $3.50",
                "sell 2 coffees",
                "show inventory",
                "update stock of coffee to 20",
                "show sales report",
                "help",
        }
        
        // Find which commands are most similar to the input
        matches := fuzzy.Find(input, commands)
        
        suggestions := make([]string, 0, 3)
        for _, match := range matches {
                if len(suggestions) >= 3 {
                        break
                }
                suggestions = append(suggestions, match.Str)
        }
        
        return suggestions
}

// min returns the smaller of two integers
func min(a, b int) int {
        if a < b {
                return a
        }
        return b
}

// analyzeSentiment performs basic sentiment analysis
func analyzeSentiment(input string) string {
        positiveWords := []string{"good", "great", "excellent", "awesome", "wonderful", "fantastic", "amazing", "love", "happy"}
        negativeWords := []string{"bad", "terrible", "awful", "horrible", "poor", "hate", "dislike", "problem", "issue", "error"}
        
        input = strings.ToLower(input)
        posCount := 0
        negCount := 0
        
        for _, word := range positiveWords {
                if strings.Contains(input, word) {
                        posCount++
                }
        }
        
        for _, word := range negativeWords {
                if strings.Contains(input, word) {
                        negCount++
                }
        }
        
        if posCount > negCount && posCount > 0 {
                return "positive"
        } else if negCount > posCount && negCount > 0 {
                return "negative"
        }
        
        return "neutral"
}

// ClearContext resets the conversation context
func ClearContext() {
        context = ConversationContext{
                PreviousIntent:      IntentUnknown,
                Entities:            make(map[string]Entity),
                ConversationHistory: []string{},
                LastTime:            time.Now(),
        }
}