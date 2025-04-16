package db

import (
        "database/sql"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "sort"
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

// GetDB returns the database connection
func GetDB() (*sql.DB, error) {
        if DB == nil {
                return nil, fmt.Errorf("database not initialized")
        }
        return DB, nil
}

// CloseDB closes the database connection and sets DB to nil
func CloseDB() error {
        if DB != nil {
                err := DB.Close()
                DB = nil
                return err
        }
        return nil
}

// BackupDatabase creates a backup of the database to the specified directory
func BackupDatabase(backupDir string) error {
        // Use default backup location if not specified
        if backupDir == "" {
                backupDir = "./backups"
        }

        // Ensure the backup directory exists
        if err := os.MkdirAll(backupDir, 0755); err != nil {
                return fmt.Errorf("failed to create backup directory: %w", err)
        }

        // Create a timestamp-based backup filename
        timestamp := time.Now().Format("20060102_150405")
        backupPath := filepath.Join(backupDir, fmt.Sprintf("pos_backup_%s.db", timestamp))

        // Get the current database path
        // For SQLite, we need to open the file directly and copy it
        srcFile, err := os.Open(GetDatabasePath())
        if err != nil {
                return fmt.Errorf("failed to open source database file: %w", err)
        }
        defer srcFile.Close()

        // Create the destination file
        dstFile, err := os.Create(backupPath)
        if err != nil {
                return fmt.Errorf("failed to create backup file: %w", err)
        }
        defer dstFile.Close()

        // Copy the file
        if _, err := io.Copy(dstFile, srcFile); err != nil {
                return fmt.Errorf("failed to copy database: %w", err)
        }

        // Save backup information in settings
        settings, err := GetSettings()
        if err == nil { // Only update if we can get settings
                settings.Backup.LastBackupTime = time.Now().Format(time.RFC3339)
                // Use admin as the updater since this is a system operation
                if err := SaveSettings(settings, "system"); err != nil {
                        fmt.Printf("Warning: Could not update backup timestamp: %v\n", err)
                        // Non-fatal, continue anyway
                }
        }

        fmt.Printf("Backup created successfully at %s\n", backupPath)
        return nil
}

// CleanupBackups removes old backups keeping only the most recent ones
func CleanupBackups(backupDir string, keepCount int) error {
        if keepCount <= 0 {
                keepCount = 7 // Default to keeping 7 backups
        }

        // Use default backup location if not specified
        if backupDir == "" {
                backupDir = "./backups"
        }

        // List backup files
        files, err := filepath.Glob(filepath.Join(backupDir, "pos_backup_*.db"))
        if err != nil {
                return fmt.Errorf("failed to list backup files: %w", err)
        }

        // If we have fewer files than we want to keep, return
        if len(files) <= keepCount {
                return nil
        }

        // Get file info for sorting
        type fileInfo struct {
                path    string
                modTime time.Time
        }

        fileInfos := make([]fileInfo, 0, len(files))
        for _, file := range files {
                info, err := os.Stat(file)
                if err != nil {
                        continue // Skip files we can't stat
                }
                fileInfos = append(fileInfos, fileInfo{
                        path:    file,
                        modTime: info.ModTime(),
                })
        }

        // Sort by modification time (newest first)
        sort.Slice(fileInfos, func(i, j int) bool {
                return fileInfos[i].modTime.After(fileInfos[j].modTime)
        })

        // Delete older files beyond the keep count
        for i := keepCount; i < len(fileInfos); i++ {
                file := fileInfos[i].path
                if err := os.Remove(file); err != nil {
                        fmt.Printf("Warning: Could not remove old backup %s: %v\n", file, err)
                        continue // Try to remove others
                }
                fmt.Printf("Removed old backup: %s\n", filepath.Base(file))
        }

        return nil
}

// GetDatabasePath returns the current database file path
func GetDatabasePath() string {
        // Return the path to the database file 
        // For now, we use a hardcoded default
        return "./pos.db"
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
                SELECT id, name, price, stock, category_id, low_stock_alert, default_supplier_id, sku, description, created_at, updated_at
                FROM products
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &product.ID,
                &product.Name,
                &product.Price,
                &product.Stock,
                &product.CategoryID,
                &product.LowStockAlert,
                &product.DefaultSupplierID,
                &product.SKU,
                &product.Description,
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

// GetProductWithDetails retrieves a product with its related details
func GetProductWithDetails(id int) (models.ProductWithDetails, error) {
        var product models.ProductWithDetails

        query := `
                SELECT p.id, p.name, p.price, p.stock, p.category_id, p.low_stock_alert, 
                       p.default_supplier_id, p.sku, p.description, p.created_at, p.updated_at,
                       c.name AS category_name, s.name AS supplier_name,
                       (SELECT COUNT(*) FROM product_batches WHERE product_id = p.id) AS batch_count,
                       (SELECT COUNT(*) FROM product_locations WHERE product_id = p.id) AS locations_count,
                       CASE WHEN EXISTS (
                           SELECT 1 FROM product_batches 
                           WHERE product_id = p.id AND expiry_date IS NOT NULL AND expiry_date < date('now')
                       ) THEN 1 ELSE 0 END AS has_expired_batches,
                       CASE WHEN p.low_stock_alert > 0 AND p.stock <= p.low_stock_alert THEN 1 ELSE 0 END AS is_low_stock
                FROM products p
                LEFT JOIN categories c ON p.category_id = c.id
                LEFT JOIN suppliers s ON p.default_supplier_id = s.id
                WHERE p.id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &product.ID,
                &product.Name,
                &product.Price,
                &product.Stock,
                &product.CategoryID,
                &product.LowStockAlert,
                &product.DefaultSupplierID,
                &product.SKU,
                &product.Description,
                &product.CreatedAt,
                &product.UpdatedAt,
                &product.CategoryName,
                &product.SupplierName,
                &product.BatchCount,
                &product.LocationsCount,
                &product.HasExpiredBatches,
                &product.IsLowStock,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return product, models.ErrProductNotFound
                }
                return product, fmt.Errorf("failed to get product with details: %w", err)
        }

        return product, nil
}

// GetAllProducts retrieves all products from the database
func GetAllProducts() ([]models.Product, error) {
        var products []models.Product

        query := `
                SELECT id, name, price, stock, category_id, low_stock_alert, 
                       default_supplier_id, sku, description, created_at, updated_at
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
                        &product.CategoryID,
                        &product.LowStockAlert,
                        &product.DefaultSupplierID,
                        &product.SKU,
                        &product.Description,
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

// GetAllProductsWithDetails retrieves all products with category and supplier details
func GetAllProductsWithDetails() ([]models.ProductWithDetails, error) {
        var products []models.ProductWithDetails

        query := `
                SELECT p.id, p.name, p.price, p.stock, p.category_id, p.low_stock_alert, 
                       p.default_supplier_id, p.sku, p.description, p.created_at, p.updated_at,
                       c.name AS category_name, s.name AS supplier_name,
                       (SELECT COUNT(*) FROM product_batches WHERE product_id = p.id) AS batch_count,
                       (SELECT COUNT(*) FROM product_locations WHERE product_id = p.id) AS locations_count,
                       CASE WHEN EXISTS (
                           SELECT 1 FROM product_batches 
                           WHERE product_id = p.id AND expiry_date IS NOT NULL AND expiry_date < date('now')
                       ) THEN 1 ELSE 0 END AS has_expired_batches,
                       CASE WHEN p.low_stock_alert > 0 AND p.stock <= p.low_stock_alert THEN 1 ELSE 0 END AS is_low_stock
                FROM products p
                LEFT JOIN categories c ON p.category_id = c.id
                LEFT JOIN suppliers s ON p.default_supplier_id = s.id
                ORDER BY p.name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query products with details: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var product models.ProductWithDetails
                err := rows.Scan(
                        &product.ID,
                        &product.Name,
                        &product.Price,
                        &product.Stock,
                        &product.CategoryID,
                        &product.LowStockAlert,
                        &product.DefaultSupplierID,
                        &product.SKU,
                        &product.Description,
                        &product.CreatedAt,
                        &product.UpdatedAt,
                        &product.CategoryName,
                        &product.SupplierName,
                        &product.BatchCount,
                        &product.LocationsCount,
                        &product.HasExpiredBatches,
                        &product.IsLowStock,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan product with details: %w", err)
                }
                products = append(products, product)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating products with details: %w", err)
        }

        return products, nil
}

// AddProduct adds a new product to the database
func AddProduct(product models.Product) (int, error) {
        // Validate the product
        if err := product.Validate(); err != nil {
                return 0, err
        }

        // Insert the product with all new fields
        query := `
                INSERT INTO products (
                        name, price, stock, category_id, low_stock_alert, 
                        default_supplier_id, sku, description, created_at, updated_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        now := time.Now()

        // Set default values if not provided
        if product.CategoryID == 0 {
                product.CategoryID = 1 // Default category
        }
        
        if product.DefaultSupplierID == 0 {
                product.DefaultSupplierID = 1 // Default supplier
        }

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                result, err := tx.Exec(
                        query, 
                        product.Name,
                        product.Price,
                        product.Stock,
                        product.CategoryID,
                        product.LowStockAlert,
                        product.DefaultSupplierID,
                        product.SKU,
                        product.Description,
                        now,
                        now,
                )
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

//
// Category Management Functions
//

// GetAllCategories retrieves all product categories
func GetAllCategories() ([]models.Category, error) {
        var categories []models.Category

        query := `
                SELECT id, name, description, parent_id, created_at, updated_at
                FROM categories
                ORDER BY name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query categories: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var category models.Category
                err := rows.Scan(
                        &category.ID,
                        &category.Name,
                        &category.Description,
                        &category.ParentID,
                        &category.CreatedAt,
                        &category.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan category: %w", err)
                }
                categories = append(categories, category)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating categories: %w", err)
        }

        return categories, nil
}

// GetCategoryByID retrieves a category by its ID
func GetCategoryByID(id int) (models.Category, error) {
        var category models.Category

        query := `
                SELECT id, name, description, parent_id, created_at, updated_at
                FROM categories
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &category.ID,
                &category.Name,
                &category.Description,
                &category.ParentID,
                &category.CreatedAt,
                &category.UpdatedAt,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return category, fmt.Errorf("category not found")
                }
                return category, fmt.Errorf("failed to get category: %w", err)
        }

        return category, nil
}

// AddCategory adds a new product category
func AddCategory(category models.Category) (int, error) {
        if category.Name == "" {
                return 0, fmt.Errorf("category name is required")
        }

        query := `
                INSERT INTO categories (name, description, parent_id, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?)
        `
        now := time.Now()

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                // Check if parent category exists if provided
                if category.ParentID > 0 {
                        var exists bool
                        err := tx.QueryRow("SELECT 1 FROM categories WHERE id = ?", category.ParentID).Scan(&exists)
                        if err != nil {
                                if err == sql.ErrNoRows {
                                        return fmt.Errorf("parent category not found")
                                }
                                return err
                        }
                }

                result, err := tx.Exec(
                        query,
                        category.Name,
                        category.Description,
                        category.ParentID,
                        now,
                        now,
                )
                if err != nil {
                        return err
                }

                id, err = result.LastInsertId()
                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to add category: %w", err)
        }

        return int(id), nil
}

//
// Supplier Management Functions
//

// GetAllSuppliers retrieves all suppliers
func GetAllSuppliers() ([]models.Supplier, error) {
        var suppliers []models.Supplier

        query := `
                SELECT id, name, contact, email, phone, address, notes, is_active, created_at, updated_at
                FROM suppliers
                ORDER BY name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query suppliers: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var supplier models.Supplier
                err := rows.Scan(
                        &supplier.ID,
                        &supplier.Name,
                        &supplier.Contact,
                        &supplier.Email,
                        &supplier.Phone,
                        &supplier.Address,
                        &supplier.Notes,
                        &supplier.IsActive,
                        &supplier.CreatedAt,
                        &supplier.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan supplier: %w", err)
                }
                suppliers = append(suppliers, supplier)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating suppliers: %w", err)
        }

        return suppliers, nil
}

// GetSupplierByID retrieves a supplier by its ID
func GetSupplierByID(id int) (models.Supplier, error) {
        var supplier models.Supplier

        query := `
                SELECT id, name, contact, email, phone, address, notes, is_active, created_at, updated_at
                FROM suppliers
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &supplier.ID,
                &supplier.Name,
                &supplier.Contact,
                &supplier.Email,
                &supplier.Phone,
                &supplier.Address,
                &supplier.Notes,
                &supplier.IsActive,
                &supplier.CreatedAt,
                &supplier.UpdatedAt,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return supplier, fmt.Errorf("supplier not found")
                }
                return supplier, fmt.Errorf("failed to get supplier: %w", err)
        }

        return supplier, nil
}

// AddSupplier adds a new supplier
func AddSupplier(supplier models.Supplier) (int, error) {
        if supplier.Name == "" {
                return 0, fmt.Errorf("supplier name is required")
        }

        query := `
                INSERT INTO suppliers (name, contact, email, phone, address, notes, is_active, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        now := time.Now()

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                result, err := tx.Exec(
                        query,
                        supplier.Name,
                        supplier.Contact,
                        supplier.Email,
                        supplier.Phone,
                        supplier.Address,
                        supplier.Notes,
                        supplier.IsActive,
                        now,
                        now,
                )
                if err != nil {
                        return err
                }

                id, err = result.LastInsertId()
                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to add supplier: %w", err)
        }

        return int(id), nil
}

//
// Location Management Functions
//

// GetAllLocations retrieves all locations
func GetAllLocations() ([]models.Location, error) {
        var locations []models.Location

        query := `
                SELECT id, name, address, description, is_active, created_at, updated_at
                FROM locations
                ORDER BY name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query locations: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var location models.Location
                err := rows.Scan(
                        &location.ID,
                        &location.Name,
                        &location.Address,
                        &location.Description,
                        &location.IsActive,
                        &location.CreatedAt,
                        &location.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan location: %w", err)
                }
                locations = append(locations, location)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating locations: %w", err)
        }

        return locations, nil
}

// GetLocationByID retrieves a location by its ID
func GetLocationByID(id int) (models.Location, error) {
        var location models.Location

        query := `
                SELECT id, name, address, description, is_active, created_at, updated_at
                FROM locations
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &location.ID,
                &location.Name,
                &location.Address,
                &location.Description,
                &location.IsActive,
                &location.CreatedAt,
                &location.UpdatedAt,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return location, fmt.Errorf("location not found")
                }
                return location, fmt.Errorf("failed to get location: %w", err)
        }

        return location, nil
}

// AddLocation adds a new location
func AddLocation(location models.Location) (int, error) {
        if location.Name == "" {
                return 0, fmt.Errorf("location name is required")
        }

        query := `
                INSERT INTO locations (name, address, description, is_active, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?)
        `
        now := time.Now()

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                result, err := tx.Exec(
                        query,
                        location.Name,
                        location.Address,
                        location.Description,
                        location.IsActive,
                        now,
                        now,
                )
                if err != nil {
                        return err
                }

                id, err = result.LastInsertId()
                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to add location: %w", err)
        }

        return int(id), nil
}

//
// Product Batch Management Functions
//

// GetProductBatches retrieves all batches for a product
func GetProductBatches(productID int) ([]models.ProductBatch, error) {
        var batches []models.ProductBatch

        query := `
                SELECT id, product_id, location_id, supplier_id, quantity, batch_number, 
                       expiry_date, manufacture_date, cost_price, receipt_date,
                       created_at, updated_at
                FROM product_batches
                WHERE product_id = ?
                ORDER BY expiry_date, created_at
        `

        rows, err := DB.Query(query, productID)
        if err != nil {
                return nil, fmt.Errorf("failed to query product batches: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var batch models.ProductBatch
                err := rows.Scan(
                        &batch.ID,
                        &batch.ProductID,
                        &batch.LocationID,
                        &batch.SupplierID,
                        &batch.Quantity,
                        &batch.BatchNumber,
                        &batch.ExpiryDate,
                        &batch.ManufactureDate,
                        &batch.CostPrice,
                        &batch.ReceiptDate,
                        &batch.CreatedAt,
                        &batch.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan product batch: %w", err)
                }
                batches = append(batches, batch)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating product batches: %w", err)
        }

        return batches, nil
}

// AddProductBatch adds a new batch for a product
func AddProductBatch(batch models.ProductBatch) (int, error) {
        // Validate required fields
        if batch.ProductID <= 0 {
                return 0, fmt.Errorf("product ID is required")
        }
        if batch.LocationID <= 0 {
                return 0, fmt.Errorf("location ID is required")
        }
        if batch.SupplierID <= 0 {
                return 0, fmt.Errorf("supplier ID is required")
        }
        if batch.Quantity <= 0 {
                return 0, fmt.Errorf("quantity must be greater than zero")
        }

        query := `
                INSERT INTO product_batches (
                        product_id, location_id, supplier_id, quantity, batch_number,
                        expiry_date, manufacture_date, cost_price, receipt_date,
                        created_at, updated_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        now := time.Now()

        var id int64
        err := Transaction(func(tx *sql.Tx) error {
                // Check if product exists
                var exists bool
                err := tx.QueryRow("SELECT 1 FROM products WHERE id = ?", batch.ProductID).Scan(&exists)
                if err != nil {
                        if err == sql.ErrNoRows {
                                return fmt.Errorf("product not found")
                        }
                        return err
                }

                // Check if location exists
                err = tx.QueryRow("SELECT 1 FROM locations WHERE id = ?", batch.LocationID).Scan(&exists)
                if err != nil {
                        if err == sql.ErrNoRows {
                                return fmt.Errorf("location not found")
                        }
                        return err
                }

                // Check if supplier exists
                err = tx.QueryRow("SELECT 1 FROM suppliers WHERE id = ?", batch.SupplierID).Scan(&exists)
                if err != nil {
                        if err == sql.ErrNoRows {
                                return fmt.Errorf("supplier not found")
                        }
                        return err
                }

                // Insert the batch
                result, err := tx.Exec(
                        query,
                        batch.ProductID,
                        batch.LocationID,
                        batch.SupplierID,
                        batch.Quantity,
                        batch.BatchNumber,
                        batch.ExpiryDate,
                        batch.ManufactureDate,
                        batch.CostPrice,
                        batch.ReceiptDate,
                        now,
                        now,
                )
                if err != nil {
                        return err
                }

                id, err = result.LastInsertId()
                if err != nil {
                        return err
                }

                // Update the product's total stock
                _, err = tx.Exec(
                        "UPDATE products SET stock = stock + ?, updated_at = ? WHERE id = ?",
                        batch.Quantity, now, batch.ProductID,
                )
                if err != nil {
                        return err
                }

                // Update or insert product_locations entry
                _, err = tx.Exec(`
                        INSERT INTO product_locations (product_id, location_id, quantity, created_at, updated_at)
                        VALUES (?, ?, ?, ?, ?)
                        ON CONFLICT(product_id, location_id) 
                        DO UPDATE SET quantity = quantity + ?, updated_at = ?
                `, batch.ProductID, batch.LocationID, batch.Quantity, now, now, batch.Quantity, now)

                return err
        })

        if err != nil {
                return 0, fmt.Errorf("failed to add product batch: %w", err)
        }

        return int(id), nil
}

// GetExpiredBatches retrieves all expired batches
func GetExpiredBatches() ([]models.ProductBatch, error) {
        var batches []models.ProductBatch

        query := `
                SELECT id, product_id, location_id, supplier_id, quantity, batch_number, 
                       expiry_date, manufacture_date, cost_price, receipt_date,
                       created_at, updated_at
                FROM product_batches
                WHERE expiry_date IS NOT NULL AND expiry_date < date('now')
                ORDER BY expiry_date
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query expired batches: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var batch models.ProductBatch
                err := rows.Scan(
                        &batch.ID,
                        &batch.ProductID,
                        &batch.LocationID,
                        &batch.SupplierID,
                        &batch.Quantity,
                        &batch.BatchNumber,
                        &batch.ExpiryDate,
                        &batch.ManufactureDate,
                        &batch.CostPrice,
                        &batch.ReceiptDate,
                        &batch.CreatedAt,
                        &batch.UpdatedAt,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan expired batch: %w", err)
                }
                batches = append(batches, batch)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating expired batches: %w", err)
        }

        return batches, nil
}

// GetLowStockProducts retrieves all products with stock below the alert threshold
func GetLowStockProducts() ([]models.ProductWithDetails, error) {
        var products []models.ProductWithDetails

        query := `
                SELECT p.id, p.name, p.price, p.stock, p.category_id, p.low_stock_alert, 
                       p.default_supplier_id, p.sku, p.description, p.created_at, p.updated_at,
                       c.name AS category_name, s.name AS supplier_name,
                       (SELECT COUNT(*) FROM product_batches WHERE product_id = p.id) AS batch_count,
                       (SELECT COUNT(*) FROM product_locations WHERE product_id = p.id) AS locations_count,
                       CASE WHEN EXISTS (
                           SELECT 1 FROM product_batches 
                           WHERE product_id = p.id AND expiry_date IS NOT NULL AND expiry_date < date('now')
                       ) THEN 1 ELSE 0 END AS has_expired_batches,
                       1 AS is_low_stock
                FROM products p
                LEFT JOIN categories c ON p.category_id = c.id
                LEFT JOIN suppliers s ON p.default_supplier_id = s.id
                WHERE p.low_stock_alert > 0 AND p.stock <= p.low_stock_alert
                ORDER BY p.name
        `

        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query low stock products: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var product models.ProductWithDetails
                err := rows.Scan(
                        &product.ID,
                        &product.Name,
                        &product.Price,
                        &product.Stock,
                        &product.CategoryID,
                        &product.LowStockAlert,
                        &product.DefaultSupplierID,
                        &product.SKU,
                        &product.Description,
                        &product.CreatedAt,
                        &product.UpdatedAt,
                        &product.CategoryName,
                        &product.SupplierName,
                        &product.BatchCount,
                        &product.LocationsCount,
                        &product.HasExpiredBatches,
                        &product.IsLowStock,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan low stock product: %w", err)
                }
                products = append(products, product)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating low stock products: %w", err)
        }

        return products, nil
}

// GetProductsByLocation retrieves all products at a specific location
func GetProductsByLocation(locationID int) ([]models.ProductWithDetails, error) {
        var products []models.ProductWithDetails

        query := `
                SELECT p.id, p.name, p.price, pl.quantity AS stock, p.category_id, p.low_stock_alert, 
                       p.default_supplier_id, p.sku, p.description, p.created_at, p.updated_at,
                       c.name AS category_name, s.name AS supplier_name,
                       (SELECT COUNT(*) FROM product_batches WHERE product_id = p.id AND location_id = ?) AS batch_count,
                       1 AS locations_count,
                       CASE WHEN EXISTS (
                           SELECT 1 FROM product_batches 
                           WHERE product_id = p.id AND location_id = ? AND expiry_date IS NOT NULL AND expiry_date < date('now')
                       ) THEN 1 ELSE 0 END AS has_expired_batches,
                       CASE WHEN p.low_stock_alert > 0 AND pl.quantity <= p.low_stock_alert THEN 1 ELSE 0 END AS is_low_stock
                FROM product_locations pl
                JOIN products p ON pl.product_id = p.id
                LEFT JOIN categories c ON p.category_id = c.id
                LEFT JOIN suppliers s ON p.default_supplier_id = s.id
                WHERE pl.location_id = ?
                ORDER BY p.name
        `

        rows, err := DB.Query(query, locationID, locationID, locationID)
        if err != nil {
                return nil, fmt.Errorf("failed to query products by location: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var product models.ProductWithDetails
                err := rows.Scan(
                        &product.ID,
                        &product.Name,
                        &product.Price,
                        &product.Stock,
                        &product.CategoryID,
                        &product.LowStockAlert,
                        &product.DefaultSupplierID,
                        &product.SKU,
                        &product.Description,
                        &product.CreatedAt,
                        &product.UpdatedAt,
                        &product.CategoryName,
                        &product.SupplierName,
                        &product.BatchCount,
                        &product.LocationsCount,
                        &product.HasExpiredBatches,
                        &product.IsLowStock,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan product by location: %w", err)
                }
                products = append(products, product)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating products by location: %w", err)
        }

        return products, nil
}

// GetProductsByCategory retrieves all products in a specific category
func GetProductsByCategory(categoryID int) ([]models.ProductWithDetails, error) {
        var products []models.ProductWithDetails

        query := `
                SELECT p.id, p.name, p.price, p.stock, p.category_id, p.low_stock_alert, 
                       p.default_supplier_id, p.sku, p.description, p.created_at, p.updated_at,
                       c.name AS category_name, s.name AS supplier_name,
                       (SELECT COUNT(*) FROM product_batches WHERE product_id = p.id) AS batch_count,
                       (SELECT COUNT(*) FROM product_locations WHERE product_id = p.id) AS locations_count,
                       CASE WHEN EXISTS (
                           SELECT 1 FROM product_batches 
                           WHERE product_id = p.id AND expiry_date IS NOT NULL AND expiry_date < date('now')
                       ) THEN 1 ELSE 0 END AS has_expired_batches,
                       CASE WHEN p.low_stock_alert > 0 AND p.stock <= p.low_stock_alert THEN 1 ELSE 0 END AS is_low_stock
                FROM products p
                LEFT JOIN categories c ON p.category_id = c.id
                LEFT JOIN suppliers s ON p.default_supplier_id = s.id
                WHERE p.category_id = ?
                ORDER BY p.name
        `

        rows, err := DB.Query(query, categoryID)
        if err != nil {
                return nil, fmt.Errorf("failed to query products by category: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var product models.ProductWithDetails
                err := rows.Scan(
                        &product.ID,
                        &product.Name,
                        &product.Price,
                        &product.Stock,
                        &product.CategoryID,
                        &product.LowStockAlert,
                        &product.DefaultSupplierID,
                        &product.SKU,
                        &product.Description,
                        &product.CreatedAt,
                        &product.UpdatedAt,
                        &product.CategoryName,
                        &product.SupplierName,
                        &product.BatchCount,
                        &product.LocationsCount,
                        &product.HasExpiredBatches,
                        &product.IsLowStock,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan product by category: %w", err)
                }
                products = append(products, product)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating products by category: %w", err)
        }

        return products, nil
}
