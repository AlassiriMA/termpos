package handlers

import (
        "database/sql"
        "fmt"
        "strings"
        "time"

        "termpos/internal/db"
        "termpos/internal/models"
)

// RecordSale records a new sale with optional discount, tax, and payment information
func RecordSale(sale models.Sale) (int, error) {
        // Validate the sale
        if err := sale.Validate(); err != nil {
                return 0, err
        }

        var id int64
        err := db.Transaction(func(tx *sql.Tx) error {
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

                // Set default payment method if not specified
                if sale.PaymentMethod == "" {
                        sale.PaymentMethod = "cash"
                }
                
                // Calculate subtotal
                subtotal := product.Price * float64(sale.Quantity)
                
                // Apply discount if specified
                if sale.DiscountAmount <= 0 && sale.DiscountCode != "" {
                        // In a real system, we would look up the discount code
                        // For now, apply a 10% discount if code provided but no amount
                        sale.DiscountAmount = subtotal * 0.1
                }
                
                // Ensure discount doesn't exceed subtotal
                if sale.DiscountAmount > subtotal {
                        sale.DiscountAmount = subtotal
                }
                
                // Apply tax if specified
                if sale.TaxRate <= 0 {
                        // Default tax rate (can be made configurable)
                        sale.TaxRate = 0.08 // 8% tax
                }
                
                // Calculate tax amount (applies to post-discount amount)
                discountedAmount := subtotal - sale.DiscountAmount
                sale.TaxAmount = discountedAmount * sale.TaxRate
                
                // Calculate total with tax
                total := discountedAmount + sale.TaxAmount
                
                // Generate receipt number
                receiptNum := fmt.Sprintf("RCP-%d-%s", time.Now().Unix(), randomString(4))
                
                // Insert the sale
                result, err := tx.Exec(
                        `INSERT INTO sales (
                                product_id, quantity, price_per_unit, 
                                discount_amount, discount_code, 
                                tax_rate, tax_amount, 
                                subtotal, total, 
                                payment_method, payment_reference,
                                receipt_number, customer_email, customer_phone,
                                notes, sale_date
                        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
                        sale.ProductID, sale.Quantity, product.Price,
                        sale.DiscountAmount, sale.DiscountCode,
                        sale.TaxRate, sale.TaxAmount,
                        subtotal, total,
                        sale.PaymentMethod, sale.PaymentReference,
                        receiptNum, sale.CustomerEmail, sale.CustomerPhone,
                        sale.Notes, time.Now(),
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
                return DecrementProductStock(tx, sale.ProductID, sale.Quantity)
        })

        if err != nil {
                return 0, fmt.Errorf("failed to record sale: %w", err)
        }

        return int(id), nil
}

// Generate a random string for receipt numbers
func randomString(length int) string {
        const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        b := make([]byte, length)
        for i := range b {
                b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
                // Add small sleep to ensure different values
                time.Sleep(1 * time.Nanosecond)
        }
        return string(b)
}

// GenerateReceipt generates a formatted receipt for a sale
func GenerateReceipt(saleID int) (string, error) {
        // Get the sale details
        query := `
                SELECT 
                        s.id, 
                        s.product_id, 
                        p.name, 
                        s.quantity, 
                        s.price_per_unit, 
                        s.discount_amount,
                        s.discount_code,
                        s.tax_rate,
                        s.tax_amount,
                        s.subtotal,
                        s.total, 
                        s.payment_method,
                        s.payment_reference,
                        s.receipt_number,
                        s.customer_email,
                        s.customer_phone,
                        s.sale_date
                FROM sales s
                JOIN products p ON s.product_id = p.id
                WHERE s.id = ?
        `
        
        var sale models.Sale
        var discountCode, paymentRef, receiptNum, custEmail, custPhone sql.NullString
        
        err := db.DB.QueryRow(query, saleID).Scan(
                &sale.ID,
                &sale.ProductID,
                &sale.ProductName,
                &sale.Quantity,
                &sale.PricePerUnit,
                &sale.DiscountAmount,
                &discountCode,
                &sale.TaxRate,
                &sale.TaxAmount,
                &sale.Subtotal,
                &sale.Total,
                &sale.PaymentMethod,
                &paymentRef,
                &receiptNum,
                &custEmail,
                &custPhone,
                &sale.SaleDate,
        )
        
        if err != nil {
                if err == sql.ErrNoRows {
                        return "", fmt.Errorf("sale not found")
                }
                return "", fmt.Errorf("failed to retrieve sale data: %w", err)
        }
        
        // Transfer NULL string values to the struct
        if discountCode.Valid {
                sale.DiscountCode = discountCode.String
        }
        if paymentRef.Valid {
                sale.PaymentReference = paymentRef.String
        }
        if receiptNum.Valid {
                sale.ReceiptNumber = receiptNum.String
        }
        if custEmail.Valid {
                sale.CustomerEmail = custEmail.String
        }
        if custPhone.Valid {
                sale.CustomerPhone = custPhone.String
        }
        
        // Format the receipt
        var sb strings.Builder
        
        sb.WriteString("===========================================\n")
        sb.WriteString("               SALES RECEIPT               \n")
        sb.WriteString("===========================================\n")
        sb.WriteString(fmt.Sprintf("Receipt Number: %s\n", sale.ReceiptNumber))
        sb.WriteString(fmt.Sprintf("Date: %s\n", sale.SaleDate.Format("2006-01-02 15:04:05")))
        sb.WriteString("-------------------------------------------\n")
        
        // Item details
        sb.WriteString(fmt.Sprintf("Item: %s\n", sale.ProductName))
        sb.WriteString(fmt.Sprintf("Quantity: %d\n", sale.Quantity))
        sb.WriteString(fmt.Sprintf("Price per unit: $%.2f\n", sale.PricePerUnit))
        sb.WriteString(fmt.Sprintf("Subtotal: $%.2f\n", sale.Subtotal))
        
        // Discount (if applicable)
        if sale.DiscountAmount > 0 {
                sb.WriteString(fmt.Sprintf("Discount: $%.2f", sale.DiscountAmount))
                if sale.DiscountCode != "" {
                        sb.WriteString(fmt.Sprintf(" (Code: %s)", sale.DiscountCode))
                }
                sb.WriteString("\n")
        }
        
        // Tax
        sb.WriteString(fmt.Sprintf("Tax (%.1f%%): $%.2f\n", sale.TaxRate*100, sale.TaxAmount))
        
        // Total
        sb.WriteString("-------------------------------------------\n")
        sb.WriteString(fmt.Sprintf("TOTAL: $%.2f\n", sale.Total))
        
        // Payment info
        sb.WriteString("-------------------------------------------\n")
        sb.WriteString(fmt.Sprintf("Payment Method: %s\n", sale.PaymentMethod))
        if sale.PaymentReference != "" {
                sb.WriteString(fmt.Sprintf("Reference: %s\n", sale.PaymentReference))
        }
        
        // Customer info if available
        if sale.CustomerEmail != "" || sale.CustomerPhone != "" {
                sb.WriteString("-------------------------------------------\n")
                if sale.CustomerEmail != "" {
                        sb.WriteString(fmt.Sprintf("Customer Email: %s\n", sale.CustomerEmail))
                }
                if sale.CustomerPhone != "" {
                        sb.WriteString(fmt.Sprintf("Customer Phone: %s\n", sale.CustomerPhone))
                }
        }
        
        sb.WriteString("===========================================\n")
        sb.WriteString("          Thank you for your purchase!     \n")
        sb.WriteString("===========================================\n")
        
        return sb.String(), nil
}

// EmailReceipt sends a receipt via email
func EmailReceipt(saleID int, email string) error {
        // In a production system, this would connect to an email service
        // For the purposes of this demo, we'll just log that we would send an email
        receipt, err := GenerateReceipt(saleID)
        if err != nil {
                return err
        }
        
        // In a real system, this would send an email
        fmt.Printf("Would send email to %s with receipt:\n%s\n", email, receipt)
        
        return nil
}

// GetAllSales retrieves all sales with product information
func GetAllSales() ([]models.Sale, error) {
        var sales []models.Sale

        query := `
                SELECT 
                        s.id, 
                        s.product_id, 
                        p.name, 
                        s.quantity, 
                        s.price_per_unit, 
                        s.discount_amount,
                        s.discount_code,
                        s.tax_rate,
                        s.tax_amount,
                        s.subtotal,
                        s.total, 
                        s.payment_method,
                        s.payment_reference,
                        s.receipt_number,
                        s.customer_email,
                        s.customer_phone,
                        s.notes,
                        s.sale_date
                FROM sales s
                JOIN products p ON s.product_id = p.id
                ORDER BY s.sale_date DESC
        `

        rows, err := db.DB.Query(query)
        if err != nil {
                return nil, fmt.Errorf("failed to query sales: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
                var sale models.Sale
                
                // Initialize pointers for NULL fields
                var discountCode, paymentRef, receiptNum, custEmail, custPhone, notes sql.NullString
                
                err := rows.Scan(
                        &sale.ID,
                        &sale.ProductID,
                        &sale.ProductName,
                        &sale.Quantity,
                        &sale.PricePerUnit,
                        &sale.DiscountAmount,
                        &discountCode,
                        &sale.TaxRate,
                        &sale.TaxAmount,
                        &sale.Subtotal,
                        &sale.Total,
                        &sale.PaymentMethod,
                        &paymentRef,
                        &receiptNum,
                        &custEmail,
                        &custPhone,
                        &notes,
                        &sale.SaleDate,
                )
                if err != nil {
                        return nil, fmt.Errorf("failed to scan sale: %w", err)
                }
                
                // Transfer NULL string values to the struct
                if discountCode.Valid {
                        sale.DiscountCode = discountCode.String
                }
                if paymentRef.Valid {
                        sale.PaymentReference = paymentRef.String
                }
                if receiptNum.Valid {
                        sale.ReceiptNumber = receiptNum.String
                }
                if custEmail.Valid {
                        sale.CustomerEmail = custEmail.String
                }
                if custPhone.Valid {
                        sale.CustomerPhone = custPhone.String
                }
                if notes.Valid {
                        sale.Notes = notes.String
                }
                
                sales = append(sales, sale)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating sales: %w", err)
        }

        return sales, nil
}
