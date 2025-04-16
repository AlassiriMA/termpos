package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

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

func startAgentServer(port int) error {
	// Set up HTTP routes
	http.HandleFunc("/products", handleProducts)
	http.HandleFunc("/products/", handleProductByID)
	http.HandleFunc("/sales", handleSales)
	http.HandleFunc("/reports/sales", handleSalesReport)
	http.HandleFunc("/reports/inventory", handleInventoryReport)
	http.HandleFunc("/reports/revenue", handleRevenueReport)

	// Start the server
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all products
		products, err := handlers.GetAllProducts()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get products: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)

	case http.MethodPost:
		// Add a new product
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

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from the URL
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get product by ID
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

	case http.MethodPut:
		// Update product stock
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

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSales(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all sales
		sales, err := handlers.GetAllSales()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get sales: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sales)

	case http.MethodPost:
		// Record a new sale
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

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
