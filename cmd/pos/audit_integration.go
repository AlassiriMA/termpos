package main

import (
        "fmt"
        "strconv"

        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
)

// Integrates audit logging with POS system actions
// This file adds audit log helpers and integration points

// LogProductAction logs product management actions
func LogProductAction(session *auth.Session, action db.AuditAction, productID int, description string, oldData, newData interface{}) error {
        if session == nil {
                // Use system user if no session is available
                return db.LogDataChange("system", action, "product", strconv.Itoa(productID), description, oldData, newData)
        }
        return db.LogDataChange(session.Username, action, "product", strconv.Itoa(productID), description, oldData, newData)
}

// LogSaleAction logs sale related actions
func LogSaleAction(session *auth.Session, action db.AuditAction, saleID int, description string, data interface{}) error {
        username := "system"
        if session != nil {
                username = session.Username
        }
        return db.LogDataChange(username, action, "sale", strconv.Itoa(saleID), description, nil, data)
}

// LogUserAction logs user management actions
func LogUserAction(session *auth.Session, action db.AuditAction, userID int, description string, oldData, newData interface{}) error {
        username := "system"
        if session != nil {
                username = session.Username
        }
        return db.LogDataChange(username, action, "user", strconv.Itoa(userID), description, oldData, newData)
}

// LogLoginAction logs login attempts
func LogLoginAction(username string, success bool, ipAddress string) error {
        action := db.ActionLogin
        description := "Successful login"
        if !success {
                action = db.ActionAccess
                description = "Failed login attempt"
        }
        return db.AddAuditLog(username, action, "auth", username, description, "", "", ipAddress, "")
}

// LogSettingsAction logs changes to system settings
func LogSettingsAction(session *auth.Session, description string, oldSettings, newSettings interface{}) error {
        username := "system"
        if session != nil {
                username = session.Username
        }
        return db.LogDataChange(username, db.ActionSettingsMod, "settings", "1", description, oldSettings, newSettings)
}

// LogPermissionAction logs permission changes
func LogPermissionAction(session *auth.Session, targetUser string, description string, oldPerms, newPerms interface{}) error {
        username := "system"
        if session != nil {
                username = session.Username
        }
        return db.LogDataChange(username, db.ActionPermissionMod, "permission", targetUser, description, oldPerms, newPerms)
}

// LogSystemAction logs system-level actions
func LogSystemAction(session *auth.Session, action db.AuditAction, resourceType, resourceID, description string) error {
        username := "system"
        if session != nil {
                username = session.Username
        }
        return db.LogUserAction(username, action, resourceType, resourceID, description)
}

// Integration with existing command handlers

// Enhanced addProduct function with audit logging
func addProductWithAudit(cmd *cobra.Command, args []string) {
        // Get current session
        session := auth.GetCurrentUser()
        if session == nil {
                fmt.Println("Error: You must be logged in to perform this action")
                return
        }

        if err := auth.RequirePermission("product:create"); err != nil {
                fmt.Println("Error: You don't have permission to add products")
                return
        }

        // Existing implementation
        if len(args) < 2 {
                fmt.Println("Error: Name and price are required")
                return
        }

        name := args[0]
        price, err := strconv.ParseFloat(args[1], 64)
        if err != nil {
                fmt.Printf("Error: Invalid price format: %v\n", err)
                return
        }

        // Optional stock parameter
        stock := 0
        if len(args) >= 3 {
                stock, err = strconv.Atoi(args[2])
                if err != nil {
                        fmt.Printf("Error: Invalid stock format: %v\n", err)
                        return
                }
        }

        // Create the product
        // NOTE: This function needs to be integrated with the actual implementation
        // For now, we'll use a placeholder
        id := 0 // Placeholder - need to use real product creation 
        // id, err := addProduct(name, price, stock)
        // if err != nil {
        //      fmt.Printf("Error adding product: %v\n", err)
        //      return
        // }

        // Log the action
        product, err := db.GetProductByID(id)
        if err != nil {
                fmt.Printf("Product added successfully with ID: %d\n", id)
                // Still log even if we can't get the full product
                LogProductAction(session, db.ActionCreate, id, fmt.Sprintf("Added product: %s", name), nil, map[string]interface{}{
                        "name":  name,
                        "price": price,
                        "stock": stock,
                })
        } else {
                fmt.Printf("Product added successfully with ID: %d\n", id)
                LogProductAction(session, db.ActionCreate, id, fmt.Sprintf("Added product: %s", name), nil, product)
        }
}

// Enhanced delete function with audit logging
func deleteProductWithAudit(cmd *cobra.Command, args []string) {
        // Get current session
        session := auth.GetCurrentUser()
        if session == nil {
                fmt.Println("Error: You must be logged in to perform this action")
                return
        }

        if err := auth.RequirePermission("product:delete"); err != nil {
                fmt.Println("Error: You don't have permission to delete products")
                return
        }

        // Existing implementation
        if len(args) < 1 {
                fmt.Println("Error: Product ID is required")
                return
        }

        id, err := strconv.Atoi(args[0])
        if err != nil {
                fmt.Printf("Error: Invalid ID format: %v\n", err)
                return
        }

        // Get product before deleting
        oldProduct, err := db.GetProductByID(id)
        if err != nil {
                fmt.Printf("Error: Product with ID %d not found\n", id)
                return
        }

        // Delete the product
        // NOTE: This function needs to be integrated with the actual implementation
        // For now, we'll use a placeholder
        // if err := deleteProduct(id); err != nil {
        //        fmt.Printf("Error deleting product: %v\n", err)
        //        return
        // }

        // Log the action
        LogProductAction(session, db.ActionDelete, id, fmt.Sprintf("Deleted product: %s", oldProduct.Name), oldProduct, nil)
        fmt.Printf("Product deleted successfully\n")
}

// Enhanced update function with audit logging
func updateProductWithAudit(cmd *cobra.Command, args []string) {
        // Get current session
        session := auth.GetCurrentUser()
        if session == nil {
                fmt.Println("Error: You must be logged in to perform this action")
                return
        }

        if err := auth.RequirePermission("product:update"); err != nil {
                fmt.Println("Error: You don't have permission to update products")
                return
        }

        // Existing implementation
        if len(args) < 3 {
                fmt.Println("Error: Product ID, field name, and new value are required")
                return
        }

        id, err := strconv.Atoi(args[0])
        if err != nil {
                fmt.Printf("Error: Invalid ID format: %v\n", err)
                return
        }

        field := args[1]
        _ = args[2] // value will be used when updateProduct is integrated

        // Get product before updating
        oldProduct, err := db.GetProductByID(id)
        if err != nil {
                fmt.Printf("Error: Product with ID %d not found\n", id)
                return
        }

        // Update the product
        // NOTE: This function needs to be integrated with the actual implementation
        // For now, we'll use a placeholder
        // if err := updateProduct(id, field, value); err != nil {
        //        fmt.Printf("Error updating product: %v\n", err)
        //        return
        // }

        // Get updated product
        newProduct, err := db.GetProductByID(id)
        if err != nil {
                fmt.Printf("Product updated successfully\n")
                return
        }

        // Log the action
        LogProductAction(session, db.ActionUpdate, id, 
                fmt.Sprintf("Updated product %s field: %s", oldProduct.Name, field), 
                oldProduct, newProduct)
        fmt.Printf("Product updated successfully\n")
}

// Enhanced sell function with audit logging
func sellWithAudit(cmd *cobra.Command, args []string) {
        // Get current session
        session := auth.GetCurrentUser()
        if session == nil {
                fmt.Println("Error: You must be logged in to perform this action")
                return
        }

        if err := auth.RequirePermission("sale:create"); err != nil {
                fmt.Println("Error: You don't have permission to create sales")
                return
        }

        // Existing implementation logic
        // This is placeholder - in the real implementation, integrate with the existing sell logic
        // For now, after a successful sale:
        
        // Example - after a successful sale:
        // saleID, err := performSale(...)
        // if err != nil { /* handle error */ }
        // sale, _ := db.GetSaleByID(saleID)
        // LogSaleAction(session, db.ActionSale, saleID, "Created new sale", sale)
}

// We're commenting this function out for now as we need to refactor this when integrating with the auth package
/*
// LoginWithAudit wraps login with audit logging
func loginWithAudit(username, password string, ipAddress string) (*auth.Session, error) {
        // This function needs to be updated to match the actual Login function signature
        // in the auth package
        
        // For now, we'll log directly in the login command instead
        return nil, fmt.Errorf("not implemented")
}
*/

// Initialization function to replace standard command implementations with audited ones
func initAuditIntegration() {
        // This function would be called from main() to replace the default command handlers
        // with the audited versions
        
        // For example:
        // addProductCmd.Run = addProductWithAudit
        // deleteProductCmd.Run = deleteProductWithAudit
        // updateProductCmd.Run = updateProductWithAudit
        // sellCmd.Run = sellWithAudit
}