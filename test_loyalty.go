package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./pos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if we have customers
	var customerCount int
	err = db.QueryRow("SELECT COUNT(*) FROM customers").Scan(&customerCount)
	if err != nil {
		log.Fatalf("Error counting customers: %v", err)
	}

	fmt.Printf("Found %d customers in the database\n", customerCount)

	// If no customers, create a test customer
	if customerCount == 0 {
		// Add three test customers with different loyalty tiers
		addTestCustomer(db, "Bronze Customer", "bronze@example.com", "555-1000", 150, "Bronze")
		addTestCustomer(db, "Silver Customer", "silver@example.com", "555-2000", 300, "Silver")
		addTestCustomer(db, "Gold Customer", "gold@example.com", "555-3000", 700, "Gold")
		fmt.Println("Added 3 test customers with different loyalty tiers")
	}

	// List all customers
	rows, err := db.Query(`
		SELECT id, name, email, phone, loyalty_points, loyalty_tier 
		FROM customers 
		ORDER BY id
	`)
	if err != nil {
		log.Fatalf("Error querying customers: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nCustomer List:")
	fmt.Println("----------------------------------------------------")
	fmt.Printf("%-3s | %-20s | %-10s | %-12s\n", "ID", "NAME", "POINTS", "TIER")
	fmt.Println("----------------------------------------------------")

	for rows.Next() {
		var id int
		var name, email, phone, tier string
		var points int
		if err := rows.Scan(&id, &name, &email, &phone, &points, &tier); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}
		fmt.Printf("%-3d | %-20s | %-10d | %-12s\n", id, name, points, tier)
	}
	fmt.Println("----------------------------------------------------")

	// Create a test sale linked to the first customer
	if customerCount > 0 {
		var firstCustomerID int
		err = db.QueryRow("SELECT id FROM customers ORDER BY id LIMIT 1").Scan(&firstCustomerID)
		if err != nil {
			log.Fatalf("Error getting first customer ID: %v", err)
		}

		// Check if we have any products to sell
		var productID, productStock int
		var productPrice float64
		var productName string
		err = db.QueryRow(`
			SELECT id, name, price, stock FROM products 
			WHERE stock > 0 ORDER BY id LIMIT 1
		`).Scan(&productID, &productName, &productPrice, &productStock)

		if err != nil {
			log.Printf("No products with stock available: %v", err)
		} else {
			// Create a test sale
			quantity := 1
			if productStock > 2 {
				quantity = 2 // Use 2 if we have enough stock
			}

			// Get customer tier
			var customerTier string
			err = db.QueryRow("SELECT loyalty_tier FROM customers WHERE id = ?", firstCustomerID).Scan(&customerTier)
			if err != nil {
				log.Fatalf("Error getting customer tier: %v", err)
			}

			// Calculate sale details
			subtotal := float64(quantity) * productPrice
			
			// Check for tier discount
			var tierDiscount float64
			switch customerTier {
			case "Platinum":
				tierDiscount = subtotal * 0.15
			case "Gold":
				tierDiscount = subtotal * 0.10
			case "Silver":
				tierDiscount = subtotal * 0.05
			}
			
			taxRate := 0.08
			taxAmount := (subtotal - tierDiscount) * taxRate
			total := subtotal - tierDiscount + taxAmount
			
			// Create receipt number
			timestamp := time.Now().Unix()
			receiptNumber := fmt.Sprintf("RCP-%d-TEST", timestamp)
			
			// Begin transaction
			tx, err := db.Begin()
			if err != nil {
				log.Fatalf("Error starting transaction: %v", err)
			}
			
			// Insert sale
			result, err := tx.Exec(`
				INSERT INTO sales (
					product_id, quantity, price_per_unit, subtotal, total,
					tax_rate, tax_amount, payment_method, receipt_number, sale_date,
					customer_id, customer_name, loyalty_discount, loyalty_tier
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, 
			productID, quantity, productPrice, subtotal, total,
			taxRate, taxAmount, "cash", receiptNumber, time.Now(),
			firstCustomerID, "Test Customer", tierDiscount, customerTier)
			
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error creating test sale: %v", err)
			}
			
			// Get the new sale ID
			saleID, err := result.LastInsertId()
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error getting sale ID: %v", err)
			}
			
			// Update product stock
			_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", quantity, productID)
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error updating product stock: %v", err)
			}
			
			// Calculate points earned
			var multiplier float64 = 1.0
			switch customerTier {
			case "Platinum":
				multiplier = 2.0
			case "Gold":
				multiplier = 1.5
			case "Silver":
				multiplier = 1.2
			}
			
			pointsEarned := int((subtotal - tierDiscount) * multiplier)
			
			// Insert customer_sales record
			_, err = tx.Exec(`
				INSERT INTO customer_sales (
					sale_id, customer_id, points_earned, points_used, created_at
				) VALUES (?, ?, ?, 0, ?)
			`, saleID, firstCustomerID, pointsEarned, time.Now())
			
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error linking sale to customer: %v", err)
			}
			
			// Update customer points and total purchases
			_, err = tx.Exec(`
				UPDATE customers SET 
					loyalty_points = loyalty_points + ?, 
					last_purchase_date = ?,
					total_purchases = total_purchases + ?
				WHERE id = ?
			`, pointsEarned, time.Now(), total, firstCustomerID)
			
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error updating customer points: %v", err)
			}
			
			// Commit transaction
			if err := tx.Commit(); err != nil {
				log.Fatalf("Error committing transaction: %v", err)
			}
			
			fmt.Printf("\nCreated test sale for customer ID %d:\n", firstCustomerID)
			fmt.Printf("Product: %s (ID: %d)\n", productName, productID)
			fmt.Printf("Quantity: %d\n", quantity)
			fmt.Printf("Subtotal: $%.2f\n", subtotal)
			if tierDiscount > 0 {
				fmt.Printf("Loyalty Discount: $%.2f\n", tierDiscount)
			}
			fmt.Printf("Tax (%.0f%%): $%.2f\n", taxRate*100, taxAmount)
			fmt.Printf("Total: $%.2f\n", total)
			fmt.Printf("Points Earned: %d\n", pointsEarned)
			fmt.Printf("Receipt Number: %s\n", receiptNumber)
		}
	}
}

func addTestCustomer(db *sql.DB, name, email, phone string, points int, tier string) {
	_, err := db.Exec(`
		INSERT INTO customers (
			name, email, phone, join_date, loyalty_points, loyalty_tier, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, name, email, phone, time.Now(), points, tier, time.Now(), time.Now())
	
	if err != nil {
		log.Printf("Error adding test customer %s: %v", name, err)
	}
}