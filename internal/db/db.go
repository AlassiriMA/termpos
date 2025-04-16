package db

import (
        "database/sql"
        "fmt"
        "os"
        "path/filepath"
        "termpos/internal/models"
        "time"

        _ "github.com/mattn/go-sqlite3"
)

var (
        DB *sql.DB
)

// Initialize sets up the database connection and runs migrations
func Initialize(dbPath string) error {
        // Ensure the directory exists
        dir := filepath.Dir(dbPath)
        if dir != "." && dir != "" {
                if err := os.MkdirAll(dir, 0755); err != nil {
                        return fmt.Errorf("failed to create directory: %w", err)
                }
        }

        // Connect to the database
        db, err := sql.Open("sqlite3", dbPath)
        if err != nil {
                return fmt.Errorf("failed to open database: %w", err)
        }

        // Test the connection
        if err := db.Ping(); err != nil {
                db.Close()
                return fmt.Errorf("failed to connect to database: %w", err)
        }

        // Set pragmas for better performance
        if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
                db.Close()
                return fmt.Errorf("failed to set journal mode: %w", err)
        }

        if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
                db.Close()
                return fmt.Errorf("failed to set synchronous mode: %w", err)
        }

        // Set the global DB variable
        DB = db

        // Run migrations
        if err := RunMigrations(); err != nil {
                DB.Close()
                return fmt.Errorf("failed to run migrations: %w", err)
        }

        return nil
}

// Close closes the database connection
func Close() error {
        if DB != nil {
                return DB.Close()
        }
        return nil
}

// UpdateProductStock updates the stock of a product
func UpdateProductStock(id int, quantity int) error {
        if quantity < 0 {
                return models.ErrInvalidStock
        }

        query := `
                UPDATE products
                SET stock = ?, updated_at = ?
                WHERE id = ?
        `

        err := Transaction(func(tx *sql.Tx) error {
                // Check if the product exists
                var exists bool
                err := tx.QueryRow("SELECT 1 FROM products WHERE id = ?", id).Scan(&exists)
                if err != nil {
                        if err == sql.ErrNoRows {
                                return models.ErrProductNotFound
                        }
                        return err
                }

                // Update the stock
                _, err = tx.Exec(query, quantity, time.Now(), id)
                return err
        })

        if err != nil {
                return fmt.Errorf("failed to update product stock: %w", err)
        }

        return nil
}

// GetProductByID retrieves a product by its ID
func GetProductByID(id int) (models.Product, error) {
        var product models.Product

        query := `
                SELECT id, name, price, stock, created_at, updated_at
                FROM products
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &product.ID,
                &product.Name,
                &product.Price,
                &product.Stock,
                &product.CreatedAt,
                &product.UpdatedAt,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return product, models.ErrProductNotFound
                }
                return product, fmt.Errorf("failed to get product: %w", err)
        }

        return product, nil
}

// GetAllProducts retrieves all products from the database
func GetAllProducts() ([]models.Product, error) {
        var products []models.Product

        query := `
                SELECT id, name, price, stock, created_at, updated_at
                FROM products
                ORDER BY name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query products: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var product models.Product
                err := rows.Scan(
                        &product.ID,
                        &product.Name,
                        &product.Price,
                        &product.Stock,
                        &product.CreatedAt,
                        &product.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan product: %w", err)
                }
                products = append(products, product)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating products: %w", err)
        }

        return products, nil
}

// AddProduct adds a new product to the database
func AddProduct(product models.Product) (int, error) {
        // Validate the product
        if err := product.Validate(); err != nil {
                return 0, err
        }

        // Insert the product
        query := `
                INSERT INTO products (name, price, stock, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?)
        `
        now := time.Now()

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                result, err := tx.Exec(query, product.Name, product.Price, product.Stock, now, now)
                if err != nil {
                        return err
                }

                id, err = result.LastInsertId()
                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to add product: %w", err)
        }

        return int(id), nil
}

// RecordSale records a new sale
func RecordSale(sale models.Sale) (int, error) {
        // Validate the sale
        if err := sale.Validate(); err != nil {
                return 0, err
        }

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                // Get the product
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
                result, err := tx.Exec(
                        "INSERT INTO sales (product_id, quantity, price_per_unit, total, sale_date) VALUES (?, ?, ?, ?, ?)",
                        sale.ProductID, sale.Quantity, product.Price, total, time.Now(),
                )
                if err != nil {
                        return err
                }

                // Get the sale ID
                id, err = result.LastInsertId()
                if err != nil {
                        return err
                }

                // Update the product stock
                _, err = tx.Exec(
                        "UPDATE products SET stock = stock - ?, updated_at = ? WHERE id = ?",
                        sale.Quantity, time.Now(), sale.ProductID,
                )
                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to record sale: %w", err)
        }

        return int(id), nil
}

// Transaction wraps a database transaction
func Transaction(fn func(*sql.Tx) error) error {
        tx, err := DB.Begin()
        if err != nil {
                return fmt.Errorf("failed to begin transaction: %w", err)
        }

        defer func() {
                if p := recover(); p != nil {
                        tx.Rollback()
                        panic(p) // re-throw panic after rollback
                }
        }()

        if err := fn(tx); err != nil {
                if rbErr := tx.Rollback(); rbErr != nil {
                        return fmt.Errorf("error on rollback: %v (original error: %w)", rbErr, err)
                }
                return err
        }

        if err := tx.Commit(); err != nil {
                return fmt.Errorf("failed to commit transaction: %w", err)
        }

        return nil
}
