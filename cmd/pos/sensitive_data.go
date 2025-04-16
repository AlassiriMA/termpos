package main

import (
        "fmt"
        "strconv"

        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
)

var (
        // Sensitive data command flags
        resourceType string
        fieldName    string
        listFields   bool
)

// sensitiveDataCmd represents the sensitive-data command
var sensitiveDataCmd = &cobra.Command{
        Use:   "sensitive",
        Short: "Manage sensitive data",
        Long:  `Store, retrieve, and manage encrypted sensitive data.`,
        Run: func(cmd *cobra.Command, args []string) {
                cmd.Help()
        },
}

// sensitiveStoreCmd stores sensitive data
var sensitiveStoreCmd = &cobra.Command{
        Use:   "store [resource-type] [resource-id] [field-name] [value]",
        Short: "Store a piece of sensitive data",
        Long:  `Encrypt and store a piece of sensitive data for a specific resource.`,
        Args:  cobra.ExactArgs(4),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("sensitive:write") != nil {
                        fmt.Println("Error: You don't have permission to store sensitive data")
                        return
                }

                // Parse arguments
                resourceType := args[0]
                resourceIDStr := args[1]
                fieldName := args[2]
                value := args[3]

                // Convert resource ID to int64
                resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
                if err != nil {
                        fmt.Printf("Error: Invalid resource ID: %v\n", err)
                        return
                }

                // Store the sensitive data
                err = db.StoreSensitiveData(resourceType, resourceID, fieldName, value)
                if err != nil {
                        fmt.Printf("Error storing sensitive data: %v\n", err)
                        return
                }

                // Log the action
                if session != nil {
                        description := fmt.Sprintf("Stored sensitive data for %s %d, field: %s", resourceType, resourceID, fieldName)
                        db.AddAuditLog(
                                session.Username,
                                db.ActionCreate,
                                "sensitive_data",
                                fmt.Sprintf("%s/%d/%s", resourceType, resourceID, fieldName),
                                description,
                                "",
                                "<encrypted>",
                                "",
                                "",
                        )
                }

                fmt.Printf("Sensitive data stored successfully for %s %d, field: %s\n", resourceType, resourceID, fieldName)
        },
}

// sensitiveGetCmd retrieves sensitive data
var sensitiveGetCmd = &cobra.Command{
        Use:   "get [resource-type] [resource-id] [field-name]",
        Short: "Get a piece of sensitive data",
        Long:  `Retrieve and decrypt a piece of sensitive data for a specific resource.`,
        Args:  cobra.ExactArgs(3),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("sensitive:read") != nil {
                        fmt.Println("Error: You don't have permission to retrieve sensitive data")
                        return
                }

                // Parse arguments
                resourceType := args[0]
                resourceIDStr := args[1]
                fieldName := args[2]

                // Convert resource ID to int64
                resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
                if err != nil {
                        fmt.Printf("Error: Invalid resource ID: %v\n", err)
                        return
                }

                // Retrieve the sensitive data
                value, err := db.GetSensitiveData(resourceType, resourceID, fieldName)
                if err != nil {
                        fmt.Printf("Error retrieving sensitive data: %v\n", err)
                        return
                }

                // Log the access
                if session != nil {
                        description := fmt.Sprintf("Retrieved sensitive data for %s %d, field: %s", resourceType, resourceID, fieldName)
                        db.AddAuditLog(
                                session.Username,
                                db.ActionAccess,
                                "sensitive_data",
                                fmt.Sprintf("%s/%d/%s", resourceType, resourceID, fieldName),
                                description,
                                "",
                                "",
                                "",
                                "",
                        )
                }

                fmt.Printf("Value: %s\n", value)
        },
}

// sensitiveDeleteCmd deletes sensitive data
var sensitiveDeleteCmd = &cobra.Command{
        Use:   "delete [resource-type] [resource-id] [field-name]",
        Short: "Delete a piece of sensitive data",
        Long:  `Delete a piece of sensitive data for a specific resource.`,
        Args:  cobra.MinimumNArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("sensitive:delete") != nil {
                        fmt.Println("Error: You don't have permission to delete sensitive data")
                        return
                }

                // Parse arguments
                resourceType := args[0]
                resourceIDStr := args[1]

                // Convert resource ID to int64
                resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
                if err != nil {
                        fmt.Printf("Error: Invalid resource ID: %v\n", err)
                        return
                }

                // If field name is provided, delete specific field
                if len(args) > 2 {
                        fieldName := args[2]
                        err = db.DeleteSensitiveData(resourceType, resourceID, fieldName)
                        if err != nil {
                                fmt.Printf("Error deleting sensitive data: %v\n", err)
                                return
                        }

                        // Log the deletion
                        if session != nil {
                                description := fmt.Sprintf("Deleted sensitive data for %s %d, field: %s", resourceType, resourceID, fieldName)
                                db.AddAuditLog(
                                        session.Username,
                                        db.ActionDelete,
                                        "sensitive_data",
                                        fmt.Sprintf("%s/%d/%s", resourceType, resourceID, fieldName),
                                        description,
                                        "",
                                        "",
                                        "",
                                        "",
                                )
                        }

                        fmt.Printf("Sensitive data deleted for %s %d, field: %s\n", resourceType, resourceID, fieldName)
                        return
                }

                // Delete all sensitive data for this resource
                err = db.DeleteAllSensitiveDataForResource(resourceType, resourceID)
                if err != nil {
                        fmt.Printf("Error deleting all sensitive data: %v\n", err)
                        return
                }

                // Log the deletion
                if session != nil {
                        description := fmt.Sprintf("Deleted all sensitive data for %s %d", resourceType, resourceID)
                        db.AddAuditLog(
                                session.Username,
                                db.ActionDelete,
                                "sensitive_data",
                                fmt.Sprintf("%s/%d", resourceType, resourceID),
                                description,
                                "",
                                "",
                                "",
                                "",
                        )
                }

                fmt.Printf("All sensitive data deleted for %s %d\n", resourceType, resourceID)
        },
}

// sensitiveListCmd lists available sensitive data fields
var sensitiveListCmd = &cobra.Command{
        Use:   "list [resource-type] [resource-id]",
        Short: "List available sensitive data fields",
        Long:  `List available sensitive data fields for a specific resource.`,
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("sensitive:read") != nil {
                        fmt.Println("Error: You don't have permission to list sensitive data")
                        return
                }

                // Parse arguments
                resourceType := args[0]
                resourceIDStr := args[1]

                // Convert resource ID to int64
                resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
                if err != nil {
                        fmt.Printf("Error: Invalid resource ID: %v\n", err)
                        return
                }

                // Get field names
                fields, err := db.GetSensitiveDataFields(resourceType, resourceID)
                if err != nil {
                        fmt.Printf("Error getting sensitive data fields: %v\n", err)
                        return
                }

                if len(fields) == 0 {
                        fmt.Printf("No sensitive data found for %s %d\n", resourceType, resourceID)
                        return
                }

                // Log the access
                if session != nil {
                        description := fmt.Sprintf("Listed sensitive data fields for %s %d", resourceType, resourceID)
                        db.AddAuditLog(
                                session.Username,
                                db.ActionAccess,
                                "sensitive_data",
                                fmt.Sprintf("%s/%d", resourceType, resourceID),
                                description,
                                "",
                                "",
                                "",
                                "",
                        )
                }

                fmt.Printf("Sensitive data fields for %s %d:\n", resourceType, resourceID)
                for i, field := range fields {
                        fmt.Printf("%d. %s\n", i+1, field)
                }
        },
}

// sensitiveCheckCmd checks if a specific field exists
var sensitiveCheckCmd = &cobra.Command{
        Use:   "check [resource-type] [resource-id] [field-name]",
        Short: "Check if a specific sensitive field exists",
        Long:  `Check if a specific sensitive field exists for a resource without revealing its value.`,
        Args:  cobra.ExactArgs(3),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("sensitive:read") != nil {
                        fmt.Println("Error: You don't have permission to check sensitive data")
                        return
                }

                // Parse arguments
                resourceType := args[0]
                resourceIDStr := args[1]
                fieldName := args[2]

                // Convert resource ID to int64
                resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
                if err != nil {
                        fmt.Printf("Error: Invalid resource ID: %v\n", err)
                        return
                }

                // Get field names
                fields, err := db.GetSensitiveDataFields(resourceType, resourceID)
                if err != nil {
                        fmt.Printf("Error checking sensitive data: %v\n", err)
                        return
                }

                // Check if field exists
                exists := false
                for _, field := range fields {
                        if field == fieldName {
                                exists = true
                                break
                        }
                }

                if exists {
                        fmt.Printf("Sensitive field '%s' exists for %s %d\n", fieldName, resourceType, resourceID)
                } else {
                        fmt.Printf("Sensitive field '%s' does not exist for %s %d\n", fieldName, resourceType, resourceID)
                }
        },
}

// sensitiveIsCmd checks if a value is sensitive
var sensitiveIsCmd = &cobra.Command{
        Use:   "is-sensitive [field-name]",
        Short: "Check if a field name should be treated as sensitive",
        Long:  `Check if a field name matches patterns that indicate it should be treated as sensitive data.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                fieldName := args[0]
                isSensitive := db.IsSensitiveField(fieldName)
                if isSensitive {
                        fmt.Printf("Field '%s' should be treated as sensitive data\n", fieldName)
                } else {
                        fmt.Printf("Field '%s' is not normally considered sensitive data\n", fieldName)
                }
        },
}

func init() {
        rootCmd.AddCommand(sensitiveDataCmd)
        sensitiveDataCmd.AddCommand(sensitiveStoreCmd)
        sensitiveDataCmd.AddCommand(sensitiveGetCmd)
        sensitiveDataCmd.AddCommand(sensitiveDeleteCmd)
        sensitiveDataCmd.AddCommand(sensitiveListCmd)
        sensitiveDataCmd.AddCommand(sensitiveCheckCmd)
        sensitiveDataCmd.AddCommand(sensitiveIsCmd)
}