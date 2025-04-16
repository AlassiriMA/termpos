package db

// createCustomersTable creates the customers table
func createCustomersTable() error {
	query := `
	CREATE TABLE customers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE,
		phone TEXT UNIQUE,
		address TEXT,
		join_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_purchase_date TIMESTAMP,
		total_purchases REAL NOT NULL DEFAULT 0,
		notes TEXT,
		loyalty_points INTEGER NOT NULL DEFAULT 0,
		loyalty_tier TEXT NOT NULL DEFAULT 'Bronze',
		birthday TEXT,
		preferred_products TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create indexes for faster lookup
	CREATE INDEX idx_customers_email ON customers(email);
	CREATE INDEX idx_customers_phone ON customers(phone);
	CREATE INDEX idx_customers_loyalty_tier ON customers(loyalty_tier);
	`

	_, err := DB.Exec(query)
	return err
}

// createLoyaltyTiersTable creates the loyalty_tiers table
func createLoyaltyTiersTable() error {
	query := `
	CREATE TABLE loyalty_tiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		min_points INTEGER NOT NULL,
		discount_percentage REAL NOT NULL DEFAULT 0,
		points_multiplier REAL NOT NULL DEFAULT 1.0,
		benefits TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create default loyalty tiers
	INSERT INTO loyalty_tiers (name, min_points, discount_percentage, points_multiplier, benefits)
	VALUES 
		('Bronze', 0, 0.0, 1.0, 'Basic loyalty program membership'),
		('Silver', 200, 0.05, 1.2, 'Earn 20% more points, 5% discount on purchases'),
		('Gold', 500, 0.10, 1.5, 'Earn 50% more points, 10% discount on purchases'),
		('Platinum', 1000, 0.15, 2.0, 'Earn double points, 15% discount on purchases');
	`

	_, err := DB.Exec(query)
	return err
}

// createLoyaltyRewardsTable creates the loyalty_rewards table
func createLoyaltyRewardsTable() error {
	query := `
	CREATE TABLE loyalty_rewards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		points_cost INTEGER NOT NULL,
		discount_value REAL NOT NULL,
		is_percentage BOOLEAN NOT NULL DEFAULT 0,
		valid_days INTEGER NOT NULL DEFAULT 30,
		active BOOLEAN NOT NULL DEFAULT 1,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create some default rewards
	INSERT INTO loyalty_rewards (name, description, points_cost, discount_value, is_percentage, valid_days)
	VALUES 
		('$5 Off Coupon', 'Get $5 off your next purchase', 100, 5.0, 0, 30),
		('10% Discount', 'Get 10% off your entire purchase', 200, 10.0, 1, 30),
		('Free Drink', 'Get a free drink with any purchase', 150, 3.5, 0, 14),
		('Buy One Get One Free', 'Buy one item and get one free', 300, 100.0, 1, 7);
	`

	_, err := DB.Exec(query)
	return err
}

// createCustomerSalesTable creates the customer_sales table
func createCustomerSalesTable() error {
	query := `
	CREATE TABLE customer_sales (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sale_id INTEGER NOT NULL,
		customer_id INTEGER NOT NULL,
		points_earned INTEGER NOT NULL DEFAULT 0,
		points_used INTEGER NOT NULL DEFAULT 0,
		reward_id INTEGER DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sale_id) REFERENCES sales (id),
		FOREIGN KEY (customer_id) REFERENCES customers (id),
		FOREIGN KEY (reward_id) REFERENCES loyalty_rewards (id)
	);
	
	-- Create indexes for faster lookup
	CREATE INDEX idx_customer_sales_sale_id ON customer_sales(sale_id);
	CREATE INDEX idx_customer_sales_customer_id ON customer_sales(customer_id);
	`

	_, err := DB.Exec(query)
	return err
}

// createLoyaltyRedemptionsTable creates the loyalty_redemptions table
func createLoyaltyRedemptionsTable() error {
	query := `
	CREATE TABLE loyalty_redemptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer_id INTEGER NOT NULL,
		reward_id INTEGER NOT NULL,
		points_used INTEGER NOT NULL,
		redeemed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expiry_date TIMESTAMP NOT NULL,
		used BOOLEAN NOT NULL DEFAULT 0,
		used_at TIMESTAMP,
		used_sale_id INTEGER,
		FOREIGN KEY (customer_id) REFERENCES customers (id),
		FOREIGN KEY (reward_id) REFERENCES loyalty_rewards (id),
		FOREIGN KEY (used_sale_id) REFERENCES sales (id)
	);
	
	-- Create indexes for faster lookup
	CREATE INDEX idx_loyalty_redemptions_customer_id ON loyalty_redemptions(customer_id);
	CREATE INDEX idx_loyalty_redemptions_active ON loyalty_redemptions(used, expiry_date);
	`

	_, err := DB.Exec(query)
	return err
}