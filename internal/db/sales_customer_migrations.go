package db

// alterSalesTableForCustomers adds customer loyalty fields to the sales table
func alterSalesTableForCustomers() error {
        // SQLite doesn't support adding multiple columns in a single ALTER TABLE statement
        // so we need to execute multiple statements
        queries := []string{
                "ALTER TABLE sales ADD COLUMN customer_id INTEGER DEFAULT 0;",
                "ALTER TABLE sales ADD COLUMN customer_name TEXT;",
                "ALTER TABLE sales ADD COLUMN loyalty_discount REAL DEFAULT 0;",
                "ALTER TABLE sales ADD COLUMN points_earned INTEGER DEFAULT 0;",
                "ALTER TABLE sales ADD COLUMN points_used INTEGER DEFAULT 0;",
                "ALTER TABLE sales ADD COLUMN loyalty_tier TEXT;",
                "ALTER TABLE sales ADD COLUMN reward_id INTEGER DEFAULT 0;",
                "ALTER TABLE sales ADD COLUMN reward_name TEXT;",
        }

        for _, q := range queries {
                _, err := DB.Exec(q)
                if err != nil {
                        return err
                }
        }
        
        return nil
}