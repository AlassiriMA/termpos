package db

import (
        "fmt"
        "time"
)

// RunMigrations applies all database migrations
func RunMigrations() error {
        // Create the schema_migrations table if it doesn't exist
        if err := createMigrationsTable(); err != nil {
                return fmt.Errorf("failed to create migrations table: %w", err)
        }

        // Run all migrations in order
        migrations := []struct {
                id       int
                name     string
                function func() error
        }{
                {1, "create_products_table", createProductsTable},
                {2, "create_sales_table", createSalesTable},
                {3, "create_users_table", createUsersTable},
                {4, "create_categories_table", createCategoriesTable},
                {5, "create_suppliers_table", createSuppliersTable},
                {6, "create_locations_table", createLocationsTable},
                {7, "create_product_batches_table", createProductBatchesTable},
                {8, "create_product_locations_table", createProductLocationsTable},
                {9, "alter_products_table", alterProductsTable},
                {10, "alter_sales_table", alterSalesTable},
                {11, "create_settings_table", createSettingsTable},
                {12, "create_audit_logs_table", createAuditLogsTable},
                {13, "create_sensitive_data_table", createSensitiveDataTable},
                {14, "create_customers_table", createCustomersTable},
                {15, "create_loyalty_tiers_table", createLoyaltyTiersTable},
                {16, "create_loyalty_rewards_table", createLoyaltyRewardsTable},
                {17, "create_customer_sales_table", createCustomerSalesTable},
                {18, "create_loyalty_redemptions_table", createLoyaltyRedemptionsTable},
                {19, "alter_sales_table_for_customers", alterSalesTableForCustomers},
        }

        for _, m := range migrations {
                applied, err := isMigrationApplied(m.id)
                if err != nil {
                        return fmt.Errorf("failed to check migration status: %w", err)
                }

                if !applied {
                        fmt.Printf("Applying migration %d: %s\n", m.id, m.name)
                        
                        if err := m.function(); err != nil {
                                return fmt.Errorf("failed to apply migration %d: %w", m.id, err)
                        }

                        if err := recordMigration(m.id, m.name); err != nil {
                                return fmt.Errorf("failed to record migration %d: %w", m.id, err)
                        }
                }
        }

        return nil
}

// createMigrationsTable creates the schema_migrations table
func createMigrationsTable() error {
        query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL,
                applied_at TIMESTAMP NOT NULL
        );`

        _, err := DB.Exec(query)
        return err
}

// isMigrationApplied checks if a migration has been applied
func isMigrationApplied(id int) (bool, error) {
        var count int
        query := "SELECT COUNT(*) FROM schema_migrations WHERE id = ?"
        err := DB.QueryRow(query, id).Scan(&count)
        if err != nil {
                return false, err
        }
        return count > 0, nil
}

// recordMigration records that a migration has been applied
func recordMigration(id int, name string) error {
        query := "INSERT INTO schema_migrations (id, name, applied_at) VALUES (?, ?, ?)"
        _, err := DB.Exec(query, id, name, time.Now())
        return err
}

// createProductsTable creates the products table
func createProductsTable() error {
        query := `
        CREATE TABLE products (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                price REAL NOT NULL,
                stock INTEGER NOT NULL DEFAULT 0,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`

        _, err := DB.Exec(query)
        return err
}

// createSalesTable creates the sales table
func createSalesTable() error {
        query := `
        CREATE TABLE sales (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                product_id INTEGER NOT NULL,
                quantity INTEGER NOT NULL,
                price_per_unit REAL NOT NULL,
                discount_amount REAL DEFAULT 0.0,
                discount_code TEXT,
                tax_rate REAL DEFAULT 0.0,
                tax_amount REAL DEFAULT 0.0,
                subtotal REAL NOT NULL,
                total REAL NOT NULL,
                payment_method TEXT DEFAULT 'cash',
                payment_reference TEXT,
                sale_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                receipt_number TEXT,
                customer_email TEXT,
                customer_phone TEXT,
                notes TEXT,
                FOREIGN KEY (product_id) REFERENCES products (id)
        );`

        _, err := DB.Exec(query)
        return err
}

// createUsersTable creates the users table
func createUsersTable() error {
        query := `
        CREATE TABLE users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL UNIQUE,
                password_hash TEXT NOT NULL,
                role TEXT NOT NULL CHECK(role IN ('admin', 'manager', 'cashier')),
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                last_login_at TIMESTAMP,
                active BOOLEAN NOT NULL DEFAULT 1
        );
        
        -- Create default admin user with password 'password123'
        INSERT INTO users (username, password_hash, role) 
        VALUES ('admin', '$2a$10$uArvRu7.jI3g.FbbkoYtLu1lmrWui9iAf1ivC6wQiEK95/t3wXaRu', 'admin');
        `

        _, err := DB.Exec(query)
        return err
}

// createCategoriesTable creates the categories table
func createCategoriesTable() error {
        query := `
        CREATE TABLE categories (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL UNIQUE,
                description TEXT,
                parent_id INTEGER,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (parent_id) REFERENCES categories (id)
        );

        -- Create a default 'Uncategorized' category
        INSERT INTO categories (name, description) 
        VALUES ('Uncategorized', 'Default category for products');
        `

        _, err := DB.Exec(query)
        return err
}

// createSuppliersTable creates the suppliers table
func createSuppliersTable() error {
        query := `
        CREATE TABLE suppliers (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL UNIQUE,
                contact TEXT,
                email TEXT,
                phone TEXT,
                address TEXT,
                notes TEXT,
                is_active BOOLEAN NOT NULL DEFAULT 1,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        -- Create a default supplier
        INSERT INTO suppliers (name, notes) 
        VALUES ('Default Supplier', 'Default supplier for products');
        `

        _, err := DB.Exec(query)
        return err
}

// createLocationsTable creates the locations table
func createLocationsTable() error {
        query := `
        CREATE TABLE locations (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL UNIQUE,
                address TEXT,
                description TEXT,
                is_active BOOLEAN NOT NULL DEFAULT 1,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        -- Create a default location
        INSERT INTO locations (name, description) 
        VALUES ('Main Warehouse', 'Default warehouse location');
        `

        _, err := DB.Exec(query)
        return err
}

// createProductBatchesTable creates the product_batches table
func createProductBatchesTable() error {
        query := `
        CREATE TABLE product_batches (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                product_id INTEGER NOT NULL,
                location_id INTEGER NOT NULL,
                supplier_id INTEGER NOT NULL,
                quantity INTEGER NOT NULL DEFAULT 0,
                batch_number TEXT,
                expiry_date TIMESTAMP,
                manufacture_date TIMESTAMP,
                cost_price REAL NOT NULL DEFAULT 0,
                receipt_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (product_id) REFERENCES products (id),
                FOREIGN KEY (location_id) REFERENCES locations (id),
                FOREIGN KEY (supplier_id) REFERENCES suppliers (id)
        );`

        _, err := DB.Exec(query)
        return err
}

// createProductLocationsTable creates the product_locations table
func createProductLocationsTable() error {
        query := `
        CREATE TABLE product_locations (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                product_id INTEGER NOT NULL,
                location_id INTEGER NOT NULL,
                quantity INTEGER NOT NULL DEFAULT 0,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (product_id) REFERENCES products (id),
                FOREIGN KEY (location_id) REFERENCES locations (id),
                UNIQUE(product_id, location_id)
        );`

        _, err := DB.Exec(query)
        return err
}

// alterProductsTable adds new columns to the products table
func alterProductsTable() error {
        // SQLite doesn't support adding multiple columns in a single ALTER TABLE statement
        // so we need to execute multiple statements
        queries := []string{
                "ALTER TABLE products ADD COLUMN category_id INTEGER DEFAULT 1 REFERENCES categories(id);",
                "ALTER TABLE products ADD COLUMN low_stock_alert INTEGER DEFAULT 0;",
                "ALTER TABLE products ADD COLUMN default_supplier_id INTEGER DEFAULT 1 REFERENCES suppliers(id);",
                "ALTER TABLE products ADD COLUMN sku TEXT;",
                "ALTER TABLE products ADD COLUMN description TEXT;",
        }

        for _, query := range queries {
                if _, err := DB.Exec(query); err != nil {
                        return err
                }
        }
        return nil
}

// alterSalesTable adds new columns to the sales table for enhanced features
func alterSalesTable() error {
        // SQLite doesn't support adding multiple columns in a single ALTER TABLE statement
        // so we need to execute multiple statements
        queries := []string{
                "ALTER TABLE sales ADD COLUMN discount_amount REAL DEFAULT 0.0;",
                "ALTER TABLE sales ADD COLUMN discount_code TEXT;",
                "ALTER TABLE sales ADD COLUMN tax_rate REAL DEFAULT 0.0;",
                "ALTER TABLE sales ADD COLUMN tax_amount REAL DEFAULT 0.0;",
                "ALTER TABLE sales ADD COLUMN subtotal REAL;",
                "ALTER TABLE sales ADD COLUMN payment_method TEXT DEFAULT 'cash';",
                "ALTER TABLE sales ADD COLUMN payment_reference TEXT;",
                "ALTER TABLE sales ADD COLUMN receipt_number TEXT;",
                "ALTER TABLE sales ADD COLUMN customer_email TEXT;",
                "ALTER TABLE sales ADD COLUMN customer_phone TEXT;",
                "ALTER TABLE sales ADD COLUMN notes TEXT;",
        }
        
        for _, query := range queries {
                if _, err := DB.Exec(query); err != nil {
                        // If column already exists, continue with the next column
                        if err.Error() == "duplicate column name: discount_amount" || 
                           err.Error() == "duplicate column name: discount_code" ||
                           err.Error() == "duplicate column name: tax_rate" ||
                           err.Error() == "duplicate column name: tax_amount" ||
                           err.Error() == "duplicate column name: subtotal" ||
                           err.Error() == "duplicate column name: payment_method" ||
                           err.Error() == "duplicate column name: payment_reference" ||
                           err.Error() == "duplicate column name: receipt_number" ||
                           err.Error() == "duplicate column name: customer_email" ||
                           err.Error() == "duplicate column name: customer_phone" ||
                           err.Error() == "duplicate column name: notes" {
                                continue
                        }
                        return err
                }
        }
        
        // Update existing sales records to set subtotal equal to total if it's NULL
        _, err := DB.Exec("UPDATE sales SET subtotal = total WHERE subtotal IS NULL")
        if err != nil {
                return err
        }
        
        return nil
}

// createSettingsTable creates the settings table for POS configuration
func createSettingsTable() error {
        query := `
        CREATE TABLE settings (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                settings_json TEXT NOT NULL,
                last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                last_updated_by TEXT,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        -- Create initial settings with default values
        INSERT INTO settings (settings_json, last_updated, last_updated_by) 
        VALUES (
                '{"store":{"name":"My Store","receipt_footer":"Thank you for your purchase!"},
                "tax":{"default_tax_rate":8.0,"tax_inclusive":false},
                "product":{"default_category_id":1,"default_supplier_id":1,"low_stock_threshold":5},
                "payment":{"enabled_payment_methods":["cash","card","mobile"],"default_payment_method":"cash"},
                "receipt":{"receipt_number_prefix":"RCP-","print_receipt_by_default":false,"email_receipt_by_default":false,"show_tax_details":true,"show_discount_details":true,"show_payment_details":true},
                "backup":{"auto_backup_enabled":false,"backup_interval_hours":24,"backup_path":"./backups","keep_backup_count":7},
                "system":{"language":"en","currency":"USD","currency_symbol":"$","date_format":"2006-01-02","time_format":"15:04:05","default_operating_mode":"classic"}}',
                CURRENT_TIMESTAMP,
                'system'
        );
        `

        _, err := DB.Exec(query)
        return err
}

// createAuditLogsTable creates the audit_logs table for tracking changes
func createAuditLogsTable() error {
        query := `
        CREATE TABLE audit_logs (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                username TEXT NOT NULL,
                action TEXT NOT NULL,
                resource_type TEXT NOT NULL,
                resource_id TEXT,
                description TEXT,
                previous_value TEXT,
                new_value TEXT,
                ip_address TEXT,
                additional_info TEXT
        );

        -- Create indexes for faster querying
        CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
        CREATE INDEX idx_audit_logs_username ON audit_logs(username);
        CREATE INDEX idx_audit_logs_action ON audit_logs(action);
        CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
        `

        _, err := DB.Exec(query)
        return err
}

// createSensitiveDataTable creates the sensitive_data table for storing encrypted values
func createSensitiveDataTable() error {
        query := `
        CREATE TABLE sensitive_data (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                resource_type TEXT NOT NULL,
                resource_id INTEGER NOT NULL,
                field_name TEXT NOT NULL,
                encrypted_value TEXT NOT NULL,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(resource_type, resource_id, field_name)
        );

        -- Create index for faster lookup
        CREATE INDEX idx_sensitive_data_resource ON sensitive_data(resource_type, resource_id);
        `

        _, err := DB.Exec(query)
        return err
}
