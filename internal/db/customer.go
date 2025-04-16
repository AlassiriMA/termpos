package db

import (
        "database/sql"
        "fmt"
        "time"

        "termpos/internal/models"
)

// AddCustomer adds a new customer to the database
func AddCustomer(customer models.Customer) (int, error) {
        var id int64
        now := time.Now()
        
        // Format birthday as string (empty if not provided)
        birthday := sql.NullString{
                String: customer.Birthday,
                Valid:  customer.Birthday != "",
        }

        // If join date not specified, use current time
        joinDate := customer.JoinDate
        if joinDate.IsZero() {
                joinDate = now
        }

        query := `
                INSERT INTO customers (
                        name, email, phone, address, join_date, notes, 
                        loyalty_points, loyalty_tier, birthday, preferred_products,
                        created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                RETURNING id
        `

        err := DB.QueryRow(
                query,
                customer.Name,
                customer.Email,
                customer.Phone,
                customer.Address,
                joinDate,
                customer.Notes,
                customer.LoyaltyPoints,
                customer.LoyaltyTier,
                birthday,
                customer.PreferredProducts,
                now,
                now,
        ).Scan(&id)

        if err != nil {
                return 0, fmt.Errorf("failed to add customer: %w", err)
        }

        // Add audit log
        AddAuditLog(
                "system", 
                ActionCreate, 
                "customers", 
                fmt.Sprintf("%d", id), 
                fmt.Sprintf("Created new customer: %s", customer.Name),
                "", 
                "", 
                "",
                "",
        )

        return int(id), nil
}

// GetCustomer retrieves a customer by ID
func GetCustomer(id int) (models.Customer, error) {
        var customer models.Customer
        var email, phone, address, notes, preferredProducts sql.NullString
        var birthday sql.NullString
        var lastPurchaseDate sql.NullTime

        query := `
                SELECT 
                        id, name, email, phone, address, join_date, last_purchase_date,
                        total_purchases, notes, loyalty_points, loyalty_tier, birthday, 
                        preferred_products, created_at, updated_at
                FROM customers
                WHERE id = ?
        `

        err := DB.QueryRow(query, id).Scan(
                &customer.ID,
                &customer.Name,
                &email,
                &phone,
                &address,
                &customer.JoinDate,
                &lastPurchaseDate,
                &customer.TotalPurchases,
                &notes,
                &customer.LoyaltyPoints,
                &customer.LoyaltyTier,
                &birthday,
                &preferredProducts,
                &customer.CreatedAt,
                &customer.UpdatedAt,
        )

        if err != nil {
                if err == sql.ErrNoRows {
                        return models.Customer{}, fmt.Errorf("customer not found")
                }
                return models.Customer{}, fmt.Errorf("failed to get customer: %w", err)
        }

        // Convert nullable fields
        if email.Valid {
                customer.Email = email.String
        }
        if phone.Valid {
                customer.Phone = phone.String
        }
        if address.Valid {
                customer.Address = address.String
        }
        if notes.Valid {
                customer.Notes = notes.String
        }
        if birthday.Valid {
                customer.Birthday = birthday.String
        }
        if preferredProducts.Valid {
                customer.PreferredProducts = preferredProducts.String
        }
        if lastPurchaseDate.Valid {
                customer.LastPurchaseDate = lastPurchaseDate.Time
        }

        return customer, nil
}

// GetCustomerByPhone retrieves a customer by phone number
func GetCustomerByPhone(phone string) (models.Customer, error) {
        var id int
        err := DB.QueryRow("SELECT id FROM customers WHERE phone = ?", phone).Scan(&id)
        if err != nil {
                if err == sql.ErrNoRows {
                        return models.Customer{}, fmt.Errorf("customer not found")
                }
                return models.Customer{}, fmt.Errorf("failed to find customer by phone: %w", err)
        }
        return GetCustomer(id)
}

// GetCustomerByEmail retrieves a customer by email
func GetCustomerByEmail(email string) (models.Customer, error) {
        var id int
        err := DB.QueryRow("SELECT id FROM customers WHERE email = ?", email).Scan(&id)
        if err != nil {
                if err == sql.ErrNoRows {
                        return models.Customer{}, fmt.Errorf("customer not found")
                }
                return models.Customer{}, fmt.Errorf("failed to find customer by email: %w", err)
        }
        return GetCustomer(id)
}

// UpdateCustomer updates customer information
func UpdateCustomer(customer models.Customer) error {
        now := time.Now()
        
        // Format birthday as string (empty if not provided)
        birthday := sql.NullString{
                String: customer.Birthday,
                Valid:  customer.Birthday != "",
        }

        query := `
                UPDATE customers SET
                        name = ?,
                        email = ?,
                        phone = ?,
                        address = ?,
                        notes = ?,
                        loyalty_points = ?,
                        loyalty_tier = ?,
                        birthday = ?,
                        preferred_products = ?,
                        updated_at = ?
                WHERE id = ?
        `

        _, err := DB.Exec(
                query,
                customer.Name,
                customer.Email,
                customer.Phone,
                customer.Address,
                customer.Notes,
                customer.LoyaltyPoints,
                customer.LoyaltyTier,
                birthday,
                customer.PreferredProducts,
                now,
                customer.ID,
        )

        if err != nil {
                return fmt.Errorf("failed to update customer: %w", err)
        }

        // Add audit log
        AddAuditLog(
                "system", 
                ActionUpdate, 
                "customers", 
                fmt.Sprintf("%d", customer.ID), 
                fmt.Sprintf("Updated customer: %s", customer.Name),
                "", 
                "", 
                "",
                "",
        )

        return nil
}

// DeleteCustomer deletes a customer
func DeleteCustomer(id int) error {
        // Get customer name for audit log
        var name string
        err := DB.QueryRow("SELECT name FROM customers WHERE id = ?", id).Scan(&name)
        if err != nil {
                return fmt.Errorf("failed to find customer: %w", err)
        }

        _, err = DB.Exec("DELETE FROM customers WHERE id = ?", id)
        if err != nil {
                return fmt.Errorf("failed to delete customer: %w", err)
        }

        // Add audit log
        AddAuditLog(
                "system", 
                ActionDelete, 
                "customers", 
                fmt.Sprintf("%d", id), 
                fmt.Sprintf("Deleted customer: %s", name),
                "", 
                "", 
                "",
                "",
        )

        return nil
}

// ListCustomers lists all customers with optional filtering
func ListCustomers(filter string, limit, offset int) ([]models.CustomerSummary, error) {
        var customers []models.CustomerSummary
        
        // Base query
        query := `
                SELECT 
                        id, name, phone, email, loyalty_points, loyalty_tier
                FROM customers
        `
        
        // Apply filters if provided
        params := []interface{}{}
        if filter != "" {
                query += " WHERE name LIKE ? OR phone LIKE ? OR email LIKE ?"
                filterParam := "%" + filter + "%"
                params = append(params, filterParam, filterParam, filterParam)
        }
        
        // Apply ordering
        query += " ORDER BY name ASC"
        
        // Apply pagination
        if limit > 0 {
                query += " LIMIT ?"
                params = append(params, limit)
                
                if offset > 0 {
                        query += " OFFSET ?"
                        params = append(params, offset)
                }
        }
        
        rows, err := DB.Query(query, params...)
        if err != nil {
                return nil, fmt.Errorf("failed to list customers: %w", err)
        }
        defer rows.Close()
        
        for rows.Next() {
                var customer models.CustomerSummary
                var email, phone sql.NullString
                
                err := rows.Scan(
                        &customer.ID,
                        &customer.Name,
                        &phone,
                        &email,
                        &customer.LoyaltyPoints,
                        &customer.LoyaltyTier,
                )
                
                if err != nil {
                        return nil, fmt.Errorf("failed to scan customer row: %w", err)
                }
                
                if email.Valid {
                        customer.Email = email.String
                }
                if phone.Valid {
                        customer.Phone = phone.String
                }
                
                customers = append(customers, customer)
        }
        
        return customers, nil
}

// LinkSaleToCustomer associates a sale with a customer and updates loyalty points
func LinkSaleToCustomer(saleID, customerID, pointsEarned, pointsUsed int, rewardID int) error {
        // Retry logic for handling database locks
        var err error
        maxRetries := 5
        retryDelay := 100 * time.Millisecond
        
        for attempt := 1; attempt <= maxRetries; attempt++ {
                // Try the operation
                err = linkSaleToCustomerWithTransaction(saleID, customerID, pointsEarned, pointsUsed, rewardID)
                
                // If successful or error is not a locked database, return
                if err == nil || !isDBLockedError(err) {
                        return err
                }
                
                // If this was the last attempt, return the error
                if attempt == maxRetries {
                        return fmt.Errorf("database still locked after %d attempts: %w", maxRetries, err)
                }
                
                // Wait before retrying with exponential backoff
                retryDelay = retryDelay * 2
                time.Sleep(retryDelay)
        }
        
        return err // Should never reach here
}

// linkSaleToCustomerWithTransaction performs the actual database operations within a transaction
func linkSaleToCustomerWithTransaction(saleID, customerID, pointsEarned, pointsUsed int, rewardID int) error {
        // Begin transaction
        tx, err := DB.Begin()
        if err != nil {
                return fmt.Errorf("failed to start transaction: %w", err)
        }
        
        // Record the customer sale link
        _, err = tx.Exec(
                "INSERT INTO customer_sales (sale_id, customer_id, points_earned, points_used, reward_id, created_at) VALUES (?, ?, ?, ?, ?, ?)",
                saleID, customerID, pointsEarned, pointsUsed, rewardID, time.Now(),
        )
        if err != nil {
                tx.Rollback()
                return fmt.Errorf("failed to link sale to customer: %w", err)
        }
        
        // Get current customer points
        var currentPoints int
        err = tx.QueryRow("SELECT loyalty_points FROM customers WHERE id = ?", customerID).Scan(&currentPoints)
        if err != nil {
                tx.Rollback()
                return fmt.Errorf("failed to get customer points: %w", err)
        }
        
        // Update customer points (add earned, subtract used)
        newPoints := currentPoints + pointsEarned - pointsUsed
        if newPoints < 0 {
                newPoints = 0 // Prevent negative points
        }
        
        // Get the sale amount for updating total purchases
        var saleAmount float64
        err = tx.QueryRow("SELECT total FROM sales WHERE id = ?", saleID).Scan(&saleAmount)
        if err != nil {
                tx.Rollback()
                return fmt.Errorf("failed to get sale amount: %w", err)
        }
        
        // Update customer record
        _, err = tx.Exec(
                `UPDATE customers SET 
                        loyalty_points = ?, 
                        loyalty_tier = ?,
                        last_purchase_date = ?,
                        total_purchases = total_purchases + ?
                 WHERE id = ?`,
                newPoints,
                models.GetLoyaltyTierName(newPoints),
                time.Now(),
                saleAmount,
                customerID,
        )
        if err != nil {
                tx.Rollback()
                return fmt.Errorf("failed to update customer points: %w", err)
        }
        
        // Commit transaction
        err = tx.Commit()
        if err != nil {
                return fmt.Errorf("failed to commit transaction: %w", err)
        }
        
        return nil
}

// isDBLockedError checks if an error is a database locked error
func isDBLockedError(err error) bool {
        if err == nil {
                return false
        }
        return err.Error() == "database is locked" || 
               err.Error() == "database table is locked" ||
               err.Error() == "database busy" ||
               err.Error() == "locked" ||
               err.Error() == "busy" ||
               err.Error() == "resource temporarily unavailable"
}

// GetLoyaltyRewards retrieves all available loyalty rewards
func GetLoyaltyRewards(activeOnly bool) ([]models.LoyaltyReward, error) {
        var rewards []models.LoyaltyReward
        
        query := "SELECT id, name, description, points_cost, discount_value, is_percentage, valid_days, active FROM loyalty_rewards"
        if activeOnly {
                query += " WHERE active = 1"
        }
        query += " ORDER BY points_cost ASC"
        
        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to get loyalty rewards: %w", err)
        }
        defer rows.Close()
        
        for rows.Next() {
                var reward models.LoyaltyReward
                
                err := rows.Scan(
                        &reward.ID,
                        &reward.Name,
                        &reward.Description,
                        &reward.PointsCost,
                        &reward.DiscountValue,
                        &reward.IsPercentage,
                        &reward.ValidDays,
                        &reward.Active,
                )
                
                if err != nil {
                        return nil, fmt.Errorf("failed to scan reward row: %w", err)
                }
                
                rewards = append(rewards, reward)
        }
        
        return rewards, nil
}

// GetLoyaltyReward retrieves a specific loyalty reward by ID
func GetLoyaltyReward(id int) (models.LoyaltyReward, error) {
        var reward models.LoyaltyReward
        
        query := "SELECT id, name, description, points_cost, discount_value, is_percentage, valid_days, active FROM loyalty_rewards WHERE id = ?"
        
        err := DB.QueryRow(query, id).Scan(
                &reward.ID,
                &reward.Name,
                &reward.Description,
                &reward.PointsCost,
                &reward.DiscountValue,
                &reward.IsPercentage,
                &reward.ValidDays,
                &reward.Active,
        )
        
        if err != nil {
                if err == sql.ErrNoRows {
                        return models.LoyaltyReward{}, fmt.Errorf("reward not found")
                }
                return models.LoyaltyReward{}, fmt.Errorf("failed to get loyalty reward: %w", err)
        }
        
        return reward, nil
}

// RedeemLoyaltyReward uses customer points to redeem a reward
func RedeemLoyaltyReward(customerID, rewardID int) (models.LoyaltyReward, error) {
        // Begin transaction
        tx, err := DB.Begin()
        if err != nil {
                return models.LoyaltyReward{}, fmt.Errorf("failed to start transaction: %w", err)
        }
        
        // Get reward details
        var reward models.LoyaltyReward
        err = tx.QueryRow(
                "SELECT id, name, description, points_cost, discount_value, is_percentage, valid_days, active FROM loyalty_rewards WHERE id = ?",
                rewardID,
        ).Scan(
                &reward.ID,
                &reward.Name,
                &reward.Description,
                &reward.PointsCost,
                &reward.DiscountValue,
                &reward.IsPercentage,
                &reward.ValidDays,
                &reward.Active,
        )
        
        if err != nil {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("failed to get reward: %w", err)
        }
        
        // Verify reward is active
        if !reward.Active {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("this reward is no longer active")
        }
        
        // Get customer points
        var currentPoints int
        err = tx.QueryRow("SELECT loyalty_points FROM customers WHERE id = ?", customerID).Scan(&currentPoints)
        if err != nil {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("failed to get customer points: %w", err)
        }
        
        // Check if customer has enough points
        if currentPoints < reward.PointsCost {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("insufficient loyalty points")
        }
        
        // Deduct points from customer
        newPoints := currentPoints - reward.PointsCost
        _, err = tx.Exec(
                "UPDATE customers SET loyalty_points = ?, updated_at = ? WHERE id = ?",
                newPoints, time.Now(), customerID,
        )
        if err != nil {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("failed to update customer points: %w", err)
        }
        
        // Record redemption in redemption history table
        _, err = tx.Exec(
                "INSERT INTO loyalty_redemptions (customer_id, reward_id, points_used, redeemed_at, expiry_date) VALUES (?, ?, ?, ?, ?)",
                customerID, rewardID, reward.PointsCost, time.Now(), time.Now().AddDate(0, 0, reward.ValidDays),
        )
        if err != nil {
                tx.Rollback()
                return models.LoyaltyReward{}, fmt.Errorf("failed to record redemption: %w", err)
        }
        
        // Commit transaction
        err = tx.Commit()
        if err != nil {
                return models.LoyaltyReward{}, fmt.Errorf("failed to commit transaction: %w", err)
        }
        
        return reward, nil
}

// GetCustomerPurchaseHistory gets a customer's purchase history
func GetCustomerPurchaseHistory(customerID int, limit int) ([]map[string]interface{}, error) {
        var history []map[string]interface{}
        
        query := `
                SELECT 
                        s.id, 
                        p.name as product_name, 
                        s.quantity, 
                        s.price_per_unit, 
                        s.total, 
                        s.sale_date,
                        cs.points_earned,
                        cs.points_used,
                        CASE WHEN cs.reward_id > 0 THEN lr.name ELSE NULL END as reward_name
                FROM sales s
                JOIN products p ON s.product_id = p.id
                JOIN customer_sales cs ON s.id = cs.sale_id
                LEFT JOIN loyalty_rewards lr ON cs.reward_id = lr.id
                WHERE cs.customer_id = ?
                ORDER BY s.sale_date DESC
        `
        
        if limit > 0 {
                query += fmt.Sprintf(" LIMIT %d", limit)
        }
        
        rows, err := DB.Query(query, customerID)
        if err != nil {
                return nil, fmt.Errorf("failed to get purchase history: %w", err)
        }
        defer rows.Close()
        
        for rows.Next() {
                var saleID, quantity, pointsEarned, pointsUsed int
                var productName, rewardName sql.NullString
                var pricePerUnit, total float64
                var saleDate time.Time
                
                err := rows.Scan(
                        &saleID,
                        &productName,
                        &quantity,
                        &pricePerUnit,
                        &total,
                        &saleDate,
                        &pointsEarned,
                        &pointsUsed,
                        &rewardName,
                )
                
                if err != nil {
                        return nil, fmt.Errorf("failed to scan history row: %w", err)
                }
                
                purchase := map[string]interface{}{
                        "sale_id":       saleID,
                        "product_name":  productName.String,
                        "quantity":      quantity,
                        "price_per_unit": pricePerUnit,
                        "total":         total,
                        "sale_date":     saleDate.Format("2006-01-02 15:04:05"),
                        "points_earned": pointsEarned,
                        "points_used":   pointsUsed,
                }
                
                if rewardName.Valid {
                        purchase["reward_name"] = rewardName.String
                }
                
                history = append(history, purchase)
        }
        
        return history, nil
}

// SearchCustomers searches for customers by name, email, or phone number
func SearchCustomers(query string, limit int) ([]models.CustomerSummary, error) {
        if query == "" {
                return nil, fmt.Errorf("search query cannot be empty")
        }
        
        searchTerm := "%" + query + "%"
        
        sqlQuery := `
                SELECT 
                        id, name, phone, email, loyalty_points, loyalty_tier
                FROM customers
                WHERE name LIKE ? OR phone LIKE ? OR email LIKE ?
                ORDER BY name ASC
        `
        
        if limit > 0 {
                sqlQuery += fmt.Sprintf(" LIMIT %d", limit)
        }
        
        rows, err := DB.Query(sqlQuery, searchTerm, searchTerm, searchTerm)
        if err != nil {
                return nil, fmt.Errorf("failed to search customers: %w", err)
        }
        defer rows.Close()
        
        var customers []models.CustomerSummary
        for rows.Next() {
                var customer models.CustomerSummary
                var email, phone sql.NullString
                
                err := rows.Scan(
                        &customer.ID,
                        &customer.Name,
                        &phone,
                        &email,
                        &customer.LoyaltyPoints,
                        &customer.LoyaltyTier,
                )
                
                if err != nil {
                        return nil, fmt.Errorf("failed to scan customer row: %w", err)
                }
                
                if email.Valid {
                        customer.Email = email.String
                }
                if phone.Valid {
                        customer.Phone = phone.String
                }
                
                customers = append(customers, customer)
        }
        
        return customers, nil
}

// GetLoyaltyTiers retrieves all loyalty tiers
func GetLoyaltyTiers() ([]models.LoyaltyTier, error) {
        var tiers []models.LoyaltyTier
        
        query := `
                SELECT 
                        id, name, min_points, discount_percentage, 
                        points_multiplier, benefits
                FROM loyalty_tiers
                ORDER BY min_points ASC
        `
        
        rows, err := DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to get loyalty tiers: %w", err)
        }
        defer rows.Close()
        
        for rows.Next() {
                var tier models.LoyaltyTier
                var benefits sql.NullString
                
                err := rows.Scan(
                        &tier.ID,
                        &tier.Name,
                        &tier.MinPoints,
                        &tier.DiscountPercentage,
                        &tier.PointsMultiplier,
                        &benefits,
                )
                
                if err != nil {
                        return nil, fmt.Errorf("failed to scan tier row: %w", err)
                }
                
                if benefits.Valid {
                        tier.Benefits = benefits.String
                }
                
                tiers = append(tiers, tier)
        }
        
        return tiers, nil
}

// CalculateLoyaltyDiscount calculates a discount based on a customer's loyalty tier
func CalculateLoyaltyDiscount(customerID int, amount float64) (float64, error) {
        // Get customer's loyalty tier
        var tierName string
        err := DB.QueryRow("SELECT loyalty_tier FROM customers WHERE id = ?", customerID).Scan(&tierName)
        if err != nil {
                return 0, fmt.Errorf("failed to get customer tier: %w", err)
        }
        
        // Get discount percentage for this tier
        var discountPercentage float64
        query := "SELECT discount_percentage FROM loyalty_tiers WHERE name = ?"
        err = DB.QueryRow(query, tierName).Scan(&discountPercentage)
        
        // If tier not found in database, use the default mapping
        if err != nil {
                if err == sql.ErrNoRows {
                        discountPercentage = models.GetLoyaltyTierDiscount(tierName)
                } else {
                        return 0, fmt.Errorf("failed to get tier discount: %w", err)
                }
        }
        
        // Calculate discount amount
        discountAmount := amount * discountPercentage
        
        return discountAmount, nil
}