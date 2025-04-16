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
		total REAL NOT NULL,
		sale_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (product_id) REFERENCES products (id)
	);`

	_, err := DB.Exec(query)
	return err
}
