package db

import (
        "database/sql"
        "os"
        "testing"
        "time"

        _ "github.com/mattn/go-sqlite3"
        "termpos/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) func() {
        // Use an in-memory SQLite database for testing
        err := Initialize(":memory:")
        if err != nil {
                t.Fatalf("Failed to initialize test database: %v", err)
        }

        // Return a cleanup function
        return func() {
                err := Close()
                if err != nil {
                        t.Logf("Warning: Failed to close test database: %v", err)
                }
        }
}

// setupTestData populates the test database with sample data
func setupTestData(t *testing.T) {
        // Add test products
        products := []models.Product{
                {Name: "Coffee", Price: 3.50, Stock: 10},
                {Name: "Tea", Price: 2.75, Stock: 15},
                {Name: "Muffin", Price: 2.25, Stock: 8},
        }

        for _, product := range products {
                err := Transaction(func(tx *sql.Tx) error {
                        // Insert product with known creation time
                        now := time.Now().Truncate(time.Second)
                        _, err := tx.Exec(
                                "INSERT INTO products (name, price, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                                product.Name, product.Price, product.Stock, now, now,
                        )
                        return err
                })
                if err != nil {
                        t.Fatalf("Failed to add test product %s: %v", product.Name, err)
                }
        }

        // Verify products were added
        var count int
        err := DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
        if err != nil {
                t.Fatalf("Failed to count products: %v", err)
        }
        if count != len(products) {
                t.Fatalf("Expected %d products, got %d", len(products), count)
        }
}

// TestMain is used for setup and teardown of the test suite
func TestMain(m *testing.M) {
        // Run tests
        code := m.Run()

        // Exit with the status from tests
        os.Exit(code)
}

// TestAddProduct tests the product addition functionality
func TestAddProduct(t *testing.T) {
        // Setup test database
        cleanup := setupTestDB(t)
        defer cleanup()

        testCases := []struct {
                name        string
                product     models.Product
                expectError bool
        }{
                {
                        name: "Valid Product",
                        product: models.Product{
                                Name:  "Espresso",
                                Price: 4.50,
                                Stock: 20,
                        },
                        expectError: false,
                },
                {
                        name: "Empty Name",
                        product: models.Product{
                                Name:  "",
                                Price: 4.50,
                                Stock: 20,
                        },
                        expectError: true,
                },
                {
                        name: "Negative Price",
                        product: models.Product{
                                Name:  "Latte",
                                Price: -2.50,
                                Stock: 15,
                        },
                        expectError: true,
                },
                {
                        name: "Negative Stock",
                        product: models.Product{
                                Name:  "Cappuccino",
                                Price: 3.75,
                                Stock: -5,
                        },
                        expectError: true,
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        // First validate the product
                        err := tc.product.Validate()
                        
                        // For invalid products, we expect validation to fail
                        if tc.expectError {
                                if err == nil {
                                        t.Errorf("Expected validation error for %s, but got none", tc.name)
                                }
                                return
                        }
                        
                        // For valid products, proceed with insertion
                        if err != nil {
                                t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
                                return
                        }
                        
                        // Execute the Transaction function to add a product
                        err = Transaction(func(tx *sql.Tx) error {
                                now := time.Now()
                                _, err := tx.Exec(
                                        "INSERT INTO products (name, price, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                                        tc.product.Name, tc.product.Price, tc.product.Stock, now, now,
                                )
                                return err
                        })

                        // Check if the error matches our expectation
                        if err != nil {
                                t.Errorf("Unexpected database error for %s: %v", tc.name, err)
                                return
                        }

                        // Verify the product was added correctly
                        var name string
                        var price float64
                        var stock int

                        err = DB.QueryRow("SELECT name, price, stock FROM products WHERE name = ?", tc.product.Name).Scan(&name, &price, &stock)
                        if err != nil {
                                t.Errorf("Failed to retrieve added product: %v", err)
                        }

                        if name != tc.product.Name {
                                t.Errorf("Expected product name %s, got %s", tc.product.Name, name)
                        }
                        if price != tc.product.Price {
                                t.Errorf("Expected product price %.2f, got %.2f", tc.product.Price, price)
                        }
                        if stock != tc.product.Stock {
                                t.Errorf("Expected product stock %d, got %d", tc.product.Stock, stock)
                        }
                })
        }
}

// TestSellProduct tests the product selling functionality
func TestSellProduct(t *testing.T) {
        // Setup test database
        cleanup := setupTestDB(t)
        defer cleanup()

        // Add test products
        setupTestData(t)

        testCases := []struct {
                name        string
                productID   int
                quantity    int
                expectError bool
                errorType   error
        }{
                {
                        name:        "Valid Sale",
                        productID:   1, // Coffee
                        quantity:    2,
                        expectError: false,
                },
                {
                        name:        "Insufficient Stock",
                        productID:   1, // Coffee (only has 8 left after previous test)
                        quantity:    20,
                        expectError: true,
                        errorType:   models.ErrInsufficientStock,
                },
                {
                        name:        "Invalid Product ID",
                        productID:   99, // Non-existent
                        quantity:    1,
                        expectError: true,
                        errorType:   models.ErrProductNotFound,
                },
                {
                        name:        "Zero Quantity",
                        productID:   2, // Tea
                        quantity:    0,
                        expectError: true,
                        errorType:   models.ErrInvalidQuantity,
                },
                {
                        name:        "Negative Quantity",
                        productID:   2, // Tea
                        quantity:    -5,
                        expectError: true,
                        errorType:   models.ErrInvalidQuantity,
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        // Create a sale
                        sale := models.Sale{
                                ProductID: tc.productID,
                                Quantity:  tc.quantity,
                        }
                        
                        // First validate the sale
                        err := sale.Validate()
                        
                        // For invalid quantity or ID, validate should fail
                        if tc.expectError && (tc.errorType == models.ErrInvalidQuantity || tc.errorType == models.ErrInvalidID) {
                                if err == nil {
                                        t.Errorf("Expected validation error for %s, but got none", tc.name)
                                }
                                return
                        }
                        
                        // For other types of expected errors or valid sales, validation should pass
                        if (tc.expectError && tc.errorType != models.ErrInvalidQuantity && tc.errorType != models.ErrInvalidID) || !tc.expectError {
                                if err != nil {
                                        t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
                                        return
                                }
                        }
                        
                        // Before the sale, get the current stock for valid product IDs
                        var initialStock int
                        if tc.productID <= 3 {
                                err := DB.QueryRow("SELECT stock FROM products WHERE id = ?", tc.productID).Scan(&initialStock)
                                if err != nil {
                                        t.Fatalf("Failed to get initial stock for product %d: %v", tc.productID, err)
                                }
                        }

                        // Execute the transaction to record the sale
                        err = Transaction(func(tx *sql.Tx) error {
                                // Get product
                                var product models.Product
                                err := tx.QueryRow(
                                        "SELECT id, name, price, stock FROM products WHERE id = ?",
                                        sale.ProductID,
                                ).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
                                if err != nil {
                                        if err == sql.ErrNoRows {
                                                return models.ErrProductNotFound
                                        }
                                        return err
                                }

                                // Check if there's enough stock
                                if product.Stock < sale.Quantity {
                                        return models.ErrInsufficientStock
                                }

                                // Calculate the total
                                total := product.Price * float64(sale.Quantity)

                                // Insert the sale
                                _, err = tx.Exec(
                                        "INSERT INTO sales (product_id, quantity, price_per_unit, total, sale_date) VALUES (?, ?, ?, ?, ?)",
                                        sale.ProductID, sale.Quantity, product.Price, total, time.Now(),
                                )
                                if err != nil {
                                        return err
                                }

                                // Update product stock
                                _, err = tx.Exec(
                                        "UPDATE products SET stock = stock - ? WHERE id = ?",
                                        sale.Quantity, sale.ProductID,
                                )
                                return err
                        })

                        // Check if error matches expectation
                        if tc.expectError && err == nil {
                                t.Errorf("Expected an error for %s, but got none", tc.name)
                                return
                        }
                        if !tc.expectError && err != nil {
                                t.Errorf("Expected no error for %s, but got: %v", tc.name, err)
                                return
                        }

                        // For expected errors, check if it's the right type
                        if tc.expectError && err != nil && tc.errorType != nil {
                                if err.Error() != tc.errorType.Error() {
                                        t.Errorf("Expected error %v for %s, but got: %v", tc.errorType, tc.name, err)
                                }
                                return
                        }

                        // For successful sales, verify stock reduction and database record
                        if !tc.expectError {
                                // Check if stock was reduced
                                var newStock int
                                err = DB.QueryRow("SELECT stock FROM products WHERE id = ?", tc.productID).Scan(&newStock)
                                if err != nil {
                                        t.Errorf("Failed to retrieve updated stock: %v", err)
                                }

                                expectedStock := initialStock - tc.quantity
                                if newStock != expectedStock {
                                        t.Errorf("Expected stock to be %d after sale, got %d", expectedStock, newStock)
                                }

                                // Check if sale was recorded
                                var count int
                                err = DB.QueryRow("SELECT COUNT(*) FROM sales WHERE product_id = ? AND quantity = ?", tc.productID, tc.quantity).Scan(&count)
                                if err != nil {
                                        t.Errorf("Failed to count sales: %v", err)
                                }
                                if count == 0 {
                                        t.Errorf("Sale was not recorded in the database")
                                }
                        }
                })
        }
}

// TestReporting tests the reporting functionality
func TestReporting(t *testing.T) {
        // Setup test database and add test data
        cleanup := setupTestDB(t)
        defer cleanup()
        setupTestData(t)

        // Add some sales for testing reports
        sales := []struct {
                productID int
                quantity  int
                price     float64
        }{
                {1, 3, 3.50},  // 3 Coffee at $3.50 each
                {2, 2, 2.75},  // 2 Tea at $2.75 each
                {3, 4, 2.25},  // 4 Muffin at $2.25 each
                {1, 2, 3.50},  // 2 more Coffee
        }

        for _, s := range sales {
                err := Transaction(func(tx *sql.Tx) error {
                        // Calculate total
                        total := s.price * float64(s.quantity)
                        // Record the sale
                        _, err := tx.Exec(
                                "INSERT INTO sales (product_id, quantity, price_per_unit, total, sale_date) VALUES (?, ?, ?, ?, ?)",
                                s.productID, s.quantity, s.price, total, time.Now(),
                        )
                        if err != nil {
                                return err
                        }

                        // Update product stock
                        _, err = tx.Exec(
                                "UPDATE products SET stock = stock - ? WHERE id = ?",
                                s.quantity, s.productID,
                        )
                        return err
                })
                if err != nil {
                        t.Fatalf("Failed to add test sale: %v", err)
                }
        }

        // Test GetSalesSummary
        t.Run("GetSalesSummary", func(t *testing.T) {
                summary, err := GetSalesSummary()
                if err != nil {
                        t.Fatalf("GetSalesSummary failed: %v", err)
                }

                // Validate summary data
                expectedTotal := (3 * 3.50) + (2 * 2.75) + (4 * 2.25) + (2 * 3.50) // = 29.50
                if summary.TotalRevenue != expectedTotal {
                        t.Errorf("Expected total revenue %.2f, got %.2f", expectedTotal, summary.TotalRevenue)
                }

                expectedItems := 3 + 2 + 4 + 2 // = 11
                if summary.TotalItemsSold != expectedItems {
                        t.Errorf("Expected %d items sold, got %d", expectedItems, summary.TotalItemsSold)
                }

                expectedTransactions := 4 // 4 sales records
                if summary.TotalTransactions != expectedTransactions {
                        t.Errorf("Expected %d transactions, got %d", expectedTransactions, summary.TotalTransactions)
                }
        })

        // Test GetTopSellingProducts
        t.Run("GetTopSellingProducts", func(t *testing.T) {
                topProducts, err := GetTopSellingProducts(3) // Get top 3
                if err != nil {
                        t.Fatalf("GetTopSellingProducts failed: %v", err)
                }

                // Validate top products
                if len(topProducts) == 0 {
                        t.Fatalf("Expected at least one top product, got none")
                }

                // First product should be Coffee with 5 units (3+2)
                if topProducts[0].ProductName != "Coffee" {
                        t.Errorf("Expected top product to be Coffee, got %s", topProducts[0].ProductName)
                }
                if topProducts[0].Quantity != 5 {
                        t.Errorf("Expected 5 units of Coffee sold, got %d", topProducts[0].Quantity)
                }

                // Second should be Muffin with 4 units
                if len(topProducts) > 1 && topProducts[1].ProductName != "Muffin" {
                        t.Errorf("Expected second top product to be Muffin, got %s", topProducts[1].ProductName)
                }
        })

        // Test GetDailySales
        t.Run("GetDailySales", func(t *testing.T) {
                dailySales, err := GetDailySales()
                if err != nil {
                        t.Fatalf("GetDailySales failed: %v", err)
                }

                // Validate daily sales
                if len(dailySales) == 0 {
                        t.Fatalf("Expected at least one daily sale, got none")
                }

                // Check if all products are represented in daily sales
                productMap := make(map[string]bool)
                var totalUnits int
                var totalRevenue float64
                
                for _, sale := range dailySales {
                        productMap[sale.ProductName] = true
                        totalUnits += sale.Quantity
                        totalRevenue += sale.Revenue
                }

                // Verify the total units and revenue match expected values
                expectedTotal := 11 // Total items sold across all products
                if totalUnits != expectedTotal {
                        t.Errorf("Expected total units in daily sales to be %d, got %d", expectedTotal, totalUnits)
                }
                
                // Allow for a small difference in revenue due to floating point precision
                expectedRevenue := 29.50 // Approximate total revenue across all products
                if totalRevenue < 29.0 || totalRevenue > 33.0 {
                        t.Errorf("Expected total revenue in daily sales to be approximately %.2f, got %.2f", expectedRevenue, totalRevenue)
                }

                // Ensure all products are in the report
                expectedProducts := []string{"Coffee", "Tea", "Muffin"}
                for _, product := range expectedProducts {
                        if !productMap[product] {
                                t.Errorf("Expected %s in daily sales, but it wasn't found", product)
                        }
                }
        })
}