package main

import (
        "encoding/json"
        "fmt"
        "net/http"
        "strconv"
        "strings"
        "time"

        "github.com/spf13/cobra"

        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/handlers"
        "termpos/internal/models"
)

// initAgentCommand sets up the agent mode server command
func initAgentCommand() {
        var agentCmd = &cobra.Command{
                Use:   "agent",
                Short: "Start the agent mode server",
                Long:  `Start an HTTP server that receives POS commands remotely.`,
                RunE: func(cmd *cobra.Command, args []string) error {
                        fmt.Printf("Starting agent mode server on port %d...\n", port)
                        return startAgentServer(port)
                },
        }

        rootCmd.AddCommand(agentCmd)
}

// authMiddleware checks if the request has a valid authentication token
func authMiddleware(next http.HandlerFunc, permission string) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                // Get token from Authorization header
                authHeader := r.Header.Get("Authorization")
                if authHeader == "" {
                        http.Error(w, "Authorization header required", http.StatusUnauthorized)
                        return
                }

                // Expected format: "Bearer <token>"
                parts := strings.Split(authHeader, " ")
                if len(parts) != 2 || parts[0] != "Bearer" {
                        http.Error(w, "Invalid authorization format, expected 'Bearer <token>'", http.StatusUnauthorized)
                        return
                }

                // In a token-based system, we would parse the token from
                // Authorization header, verify and then use it to authenticate.
                // For now, we'll just check if the user is already authenticated
                // and has the required permission.
                if !auth.IsAuthenticated() {
                        http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
                        return
                }
                
                // Check permission
                if err := auth.RequirePermission(permission); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }

                // Call the next handler
                next(w, r)
        }
}

// Handle authentication routes
func handleLogin(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        var creds struct {
                Username string `json:"username"`
                Password string `json:"password"`
        }

        if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        token, err := auth.Login(creds.Username, creds.Password, db.GetUserByUsername, db.UpdateLastLogin)
        if err != nil {
                http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
                return
        }

        // Create a token string response
        tokenStr := fmt.Sprintf("%d:%s:%s", token.UserID, token.Username, token.Role)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}

// productHandler handles requests based on HTTP method and required permissions
func productHandler(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
                // View products requires inventory:view permission
                if err := auth.RequirePermission("inventory:view"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleGetProducts(w, r)
        case http.MethodPost:
                // Add products requires product:manage permission
                if err := auth.RequirePermission("product:manage"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleAddProduct(w, r)
        default:
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
}

// productByIDHandler handles requests based on HTTP method and required permissions
func productByIDHandler(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
                // View product requires inventory:view permission
                if err := auth.RequirePermission("inventory:view"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleGetProductByID(w, r)
        case http.MethodPut:
                // Update product requires product:manage permission
                if err := auth.RequirePermission("product:manage"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleUpdateProductStock(w, r)
        default:
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
}

// salesHandler handles requests based on HTTP method and required permissions
func salesHandler(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
                // View sales requires report:generate permission
                if err := auth.RequirePermission("report:generate"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleGetSales(w, r)
        case http.MethodPost:
                // Add sales requires sales:create permission
                if err := auth.RequirePermission("sales:create"); err != nil {
                        http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
                        return
                }
                handleAddSale(w, r)
        default:
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
}

func startAgentServer(port int) error {
        // Authentication endpoints
        http.HandleFunc("/auth/login", handleLogin)
        
        // Product, sales and inventory routes
        http.HandleFunc("/products", productHandler)
        http.HandleFunc("/products/", productByIDHandler)
        http.HandleFunc("/sales", salesHandler)
        
        // Report routes - all require report:generate permission
        http.HandleFunc("/reports/sales", authMiddleware(handleSalesReport, "report:generate"))
        http.HandleFunc("/reports/inventory", authMiddleware(handleInventoryReport, "report:generate"))
        http.HandleFunc("/reports/revenue", authMiddleware(handleRevenueReport, "report:generate"))
        http.HandleFunc("/reports/summary", authMiddleware(handleSummaryReport, "report:generate"))
        http.HandleFunc("/reports/top", authMiddleware(handleTopProductsReport, "report:generate"))
        http.HandleFunc("/reports/daily", authMiddleware(handleDailySalesReport, "report:generate"))

        // Start the server
        addr := fmt.Sprintf("0.0.0.0:%d", port)
        fmt.Printf("Server listening on %s\n", addr)
        return http.ListenAndServe(addr, nil)
}

// handleGetProducts returns all products
func handleGetProducts(w http.ResponseWriter, r *http.Request) {
        products, err := handlers.GetAllProducts()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get products: %v", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(products)
}

// handleAddProduct adds a new product
func handleAddProduct(w http.ResponseWriter, r *http.Request) {
        var product models.Product
        if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        id, err := handlers.AddProduct(product)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to add product: %v", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]int{"id": id})
}

// handleGetProductByID returns a specific product by ID
func handleGetProductByID(w http.ResponseWriter, r *http.Request) {
        // Extract product ID from the URL
        idStr := r.URL.Path[len("/products/"):]
        id, err := strconv.Atoi(idStr)
        if err != nil {
                http.Error(w, "Invalid product ID", http.StatusBadRequest)
                return
        }
        
        product, err := handlers.GetProductByID(id)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get product: %v", err), http.StatusInternalServerError)
                return
        }

        if product.ID == 0 {
                http.Error(w, "Product not found", http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(product)
}

// handleUpdateProductStock updates the stock of a product
func handleUpdateProductStock(w http.ResponseWriter, r *http.Request) {
        // Extract product ID from the URL
        idStr := r.URL.Path[len("/products/"):]
        id, err := strconv.Atoi(idStr)
        if err != nil {
                http.Error(w, "Invalid product ID", http.StatusBadRequest)
                return
        }
        
        var data struct {
                Stock int `json:"stock"`
        }
        if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        if err := handlers.UpdateProductStock(id, data.Stock); err != nil {
                http.Error(w, fmt.Sprintf("Failed to update stock: %v", err), http.StatusInternalServerError)
                return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleGetSales returns all sales records
func handleGetSales(w http.ResponseWriter, r *http.Request) {
        sales, err := handlers.GetAllSales()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get sales: %v", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(sales)
}

// handleAddSale records a new sale
func handleAddSale(w http.ResponseWriter, r *http.Request) {
        var sale models.Sale
        if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        id, err := handlers.RecordSale(sale)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to record sale: %v", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func handleSalesReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        sales, err := handlers.GetAllSales()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get sales data: %v", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(sales)
}

func handleInventoryReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        products, err := handlers.GetAllProducts()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get inventory data: %v", err), http.StatusInternalServerError)
                return
        }

        // Calculate inventory value
        var result []map[string]interface{}
        var totalValue float64

        for _, p := range products {
                value := p.Price * float64(p.Stock)
                totalValue += value

                item := map[string]interface{}{
                        "id":     p.ID,
                        "name":   p.Name,
                        "price":  p.Price,
                        "stock":  p.Stock,
                        "value":  value,
                }
                result = append(result, item)
        }

        response := map[string]interface{}{
                "items":       result,
                "totalValue":  totalValue,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

func handleRevenueReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        revenue, err := handlers.GetRevenueReport()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get revenue data: %v", err), http.StatusInternalServerError)
                return
        }

        // Calculate total revenue
        var totalRevenue float64
        for _, r := range revenue {
                totalRevenue += r.Revenue
        }

        response := map[string]interface{}{
                "items":        revenue,
                "totalRevenue": totalRevenue,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

func handleSummaryReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        summary, err := db.GetSalesSummary()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get sales summary: %v", err), http.StatusInternalServerError)
                return
        }

        // Calculate average transaction value if there are transactions
        var avgTransactionValue float64
        if summary.TotalTransactions > 0 {
                avgTransactionValue = summary.TotalRevenue / float64(summary.TotalTransactions)
        }

        response := map[string]interface{}{
                "totalRevenue":         summary.TotalRevenue,
                "totalItemsSold":       summary.TotalItemsSold,
                "totalTransactions":    summary.TotalTransactions,
                "avgTransactionValue":  avgTransactionValue,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

func handleTopProductsReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        // Get limit parameter from query string, default to 5
        limitStr := r.URL.Query().Get("limit")
        limit := 5
        if limitStr != "" {
                var err error
                limit, err = strconv.Atoi(limitStr)
                if err != nil || limit <= 0 {
                        limit = 5
                }
        }

        topProducts, err := db.GetTopSellingProducts(limit)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get top selling products: %v", err), http.StatusInternalServerError)
                return
        }

        response := map[string]interface{}{
                "items": topProducts,
                "limit": limit,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

func handleDailySalesReport(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        dailySales, err := db.GetDailySales()
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get daily sales: %v", err), http.StatusInternalServerError)
                return
        }

        // Calculate totals
        var totalUnits int
        var totalRevenue float64
        for _, s := range dailySales {
                totalUnits += s.Quantity
                totalRevenue += s.Revenue
        }

        response := map[string]interface{}{
                "date":          time.Now().Format("2006-01-02"),
                "items":         dailySales,
                "totalUnits":    totalUnits,
                "totalRevenue":  totalRevenue,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}
