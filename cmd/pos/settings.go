package main

import (
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "strings"
        "time"

        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/models"

        "github.com/olekukonko/tablewriter"
        "github.com/spf13/cobra"
)

// settingsCmd represents the settings command
var settingsCmd = &cobra.Command{
        Use:   "settings",
        Short: "Manage POS settings",
        Long:  `Configure store details, product defaults, and system settings.`,
        Run: func(cmd *cobra.Command, args []string) {
                cmd.Help()
        },
}

// settingsViewCmd shows current settings
var settingsViewCmd = &cobra.Command{
        Use:   "view",
        Short: "View current settings",
        Long:  `View the current configuration of the POS system.`,
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:read") != nil {
                        fmt.Println("Error: You don't have permission to view settings")
                        return
                }

                // Get settings from DB
                settings, err := db.GetSettings()
                if err != nil {
                        fmt.Printf("Error getting settings: %v\n", err)
                        return
                }

                // Format and print settings
                displaySettings(settings)
        },
}

// settingsUpdateCmd updates settings
var settingsUpdateCmd = &cobra.Command{
        Use:   "update [key] [value]",
        Short: "Update a setting",
        Long:  `Update a specific setting value. Keys use dot notation, e.g. 'store.name' or 'tax.default_tax_rate'.`,
        Args:  cobra.MinimumNArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:update") != nil {
                        fmt.Println("Error: You don't have permission to update settings")
                        return
                }

                key := args[0]
                value := args[1]

                // Get current settings
                settings, err := db.GetSettings()
                if err != nil {
                        fmt.Printf("Error getting settings: %v\n", err)
                        return
                }

                // Update the settings
                updated, err := updateSettingsByKey(settings, key, value)
                if err != nil {
                        fmt.Printf("Error updating settings: %v\n", err)
                        return
                }

                // Save updated settings to DB
                err = db.SaveSettings(updated, session.Username)
                if err != nil {
                        fmt.Printf("Error saving settings: %v\n", err)
                        return
                }

                fmt.Printf("Setting '%s' updated successfully to '%s'\n", key, value)
        },
}

// settingsExportCmd exports settings to a file
var settingsExportCmd = &cobra.Command{
        Use:   "export [filepath]",
        Short: "Export settings to a file",
        Long:  `Export the current configuration to a JSON file.`,
        Args:  cobra.MaximumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:export") != nil {
                        fmt.Println("Error: You don't have permission to export settings")
                        return
                }

                // Get settings from DB
                settings, err := db.GetSettings()
                if err != nil {
                        fmt.Printf("Error getting settings: %v\n", err)
                        return
                }

                // Export to JSON
                jsonData, err := settings.ExportToJSON()
                if err != nil {
                        fmt.Printf("Error exporting settings: %v\n", err)
                        return
                }

                // Determine output file
                outputFile := "./settings.json"
                if len(args) > 0 {
                        outputFile = args[0]
                }

                // Create directory if it doesn't exist
                dir := filepath.Dir(outputFile)
                if dir != "." {
                        if err := os.MkdirAll(dir, 0755); err != nil {
                                fmt.Printf("Error creating directory: %v\n", err)
                                return
                        }
                }

                // Write to file
                err = os.WriteFile(outputFile, []byte(jsonData), 0644)
                if err != nil {
                        fmt.Printf("Error writing to file: %v\n", err)
                        return
                }

                fmt.Printf("Settings exported to %s\n", outputFile)
        },
}

// settingsImportCmd imports settings from a file
var settingsImportCmd = &cobra.Command{
        Use:   "import [filepath]",
        Short: "Import settings from a file",
        Long:  `Import configuration from a JSON file.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:import") != nil {
                        fmt.Println("Error: You don't have permission to import settings")
                        return
                }

                // Read file
                jsonData, err := os.ReadFile(args[0])
                if err != nil {
                        fmt.Printf("Error reading file: %v\n", err)
                        return
                }

                // Import from JSON
                settings, err := models.ImportFromJSON(string(jsonData))
                if err != nil {
                        fmt.Printf("Error importing settings: %v\n", err)
                        return
                }

                // Save to DB
                err = db.SaveSettings(settings, session.Username)
                if err != nil {
                        fmt.Printf("Error saving settings: %v\n", err)
                        return
                }

                fmt.Println("Settings imported successfully")
        },
}

// backupCmd creates a database backup
var backupCmd = &cobra.Command{
        Use:   "backup [directory]",
        Short: "Create a database backup",
        Long:  `Create a backup of the POS database.`,
        Args:  cobra.MaximumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:backup") != nil {
                        fmt.Println("Error: You don't have permission to create backups")
                        return
                }

                // Determine backup path
                backupPath := ""
                if len(args) > 0 {
                        backupPath = args[0]
                }

                // Create backup
                err := db.BackupDatabase(backupPath)
                if err != nil {
                        fmt.Printf("Error creating backup: %v\n", err)
                        return
                }

                fmt.Println("Database backup created successfully")
        },
}

// settingsListCmd shows all settings
var settingsListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all settings",
        Long:  `Display a comprehensive list of all system settings.`,
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:read") != nil {
                        fmt.Println("Error: You don't have permission to view settings")
                        return
                }

                // Get settings from DB
                settings, err := db.GetSettings()
                if err != nil {
                        fmt.Printf("Error getting settings: %v\n", err)
                        return
                }

                // Format and print settings
                displaySettings(settings)
        },
}

// initialize sets up the settings commands
func init() {
        rootCmd.AddCommand(settingsCmd)
        settingsCmd.AddCommand(settingsViewCmd)
        settingsCmd.AddCommand(settingsListCmd)  // Add the list command
        settingsCmd.AddCommand(settingsUpdateCmd)
        settingsCmd.AddCommand(settingsExportCmd)
        settingsCmd.AddCommand(settingsImportCmd)
        settingsCmd.AddCommand(backupCmd)
}

// displaySettings formats and prints the settings to the console
func displaySettings(settings models.Settings) {
        // Print store info section
        fmt.Println("=== Store Information ===")
        storeTable := tablewriter.NewWriter(os.Stdout)
        storeTable.SetHeader([]string{"Setting", "Value"})
        storeTable.SetBorder(false)
        storeTable.SetColumnSeparator(" | ")
        storeTable.Append([]string{"Name", settings.Store.Name})
        if settings.Store.Address != "" {
                storeTable.Append([]string{"Address", settings.Store.Address})
        }
        if settings.Store.Phone != "" {
                storeTable.Append([]string{"Phone", settings.Store.Phone})
        }
        if settings.Store.Email != "" {
                storeTable.Append([]string{"Email", settings.Store.Email})
        }
        if settings.Store.Website != "" {
                storeTable.Append([]string{"Website", settings.Store.Website})
        }
        if settings.Store.TaxID != "" {
                storeTable.Append([]string{"Tax ID", settings.Store.TaxID})
        }
        if settings.Store.RegistrationNum != "" {
                storeTable.Append([]string{"Registration #", settings.Store.RegistrationNum})
        }
        if settings.Store.ReceiptFooter != "" {
                storeTable.Append([]string{"Receipt Footer", settings.Store.ReceiptFooter})
        }
        if settings.Store.ReceiptHeader != "" {
                storeTable.Append([]string{"Receipt Header", settings.Store.ReceiptHeader})
        }
        storeTable.Render()
        fmt.Println()

        // Print tax settings
        fmt.Println("=== Tax Settings ===")
        taxTable := tablewriter.NewWriter(os.Stdout)
        taxTable.SetHeader([]string{"Setting", "Value"})
        taxTable.SetBorder(false)
        taxTable.SetColumnSeparator(" | ")
        taxTable.Append([]string{"Default Tax Rate", fmt.Sprintf("%.2f%%", settings.Tax.DefaultTaxRate)})
        taxTable.Append([]string{"Tax-Inclusive Pricing", fmt.Sprintf("%t", settings.Tax.TaxInclusive)})
        taxTable.Render()
        fmt.Println()

        // Print product settings
        fmt.Println("=== Product Settings ===")
        productTable := tablewriter.NewWriter(os.Stdout)
        productTable.SetHeader([]string{"Setting", "Value"})
        productTable.SetBorder(false)
        productTable.SetColumnSeparator(" | ")
        productTable.Append([]string{"Default Category ID", fmt.Sprintf("%d", settings.Product.DefaultCategory)})
        productTable.Append([]string{"Default Supplier ID", fmt.Sprintf("%d", settings.Product.DefaultSupplier)})
        productTable.Append([]string{"Low Stock Threshold", fmt.Sprintf("%d", settings.Product.LowStockThreshold)})
        productTable.Append([]string{"Batch Tracking", fmt.Sprintf("%t", settings.Product.EnableBatchTracking)})
        productTable.Append([]string{"Expiry Tracking", fmt.Sprintf("%t", settings.Product.EnableExpiryTracking)})
        productTable.Append([]string{"Location Tracking", fmt.Sprintf("%t", settings.Product.EnableLocationTracking)})
        if settings.Product.SKUPrefix != "" {
                productTable.Append([]string{"SKU Prefix", settings.Product.SKUPrefix})
        }
        productTable.Render()
        fmt.Println()

        // Print payment settings
        fmt.Println("=== Payment Settings ===")
        paymentTable := tablewriter.NewWriter(os.Stdout)
        paymentTable.SetHeader([]string{"Setting", "Value"})
        paymentTable.SetBorder(false)
        paymentTable.SetColumnSeparator(" | ")
        paymentTable.Append([]string{"Enabled Payment Methods", strings.Join(settings.Payment.EnabledPaymentMethods, ", ")})
        paymentTable.Append([]string{"Default Payment Method", settings.Payment.DefaultPaymentMethod})
        paymentTable.Render()
        fmt.Println()

        // Print receipt settings
        fmt.Println("=== Receipt Settings ===")
        receiptTable := tablewriter.NewWriter(os.Stdout)
        receiptTable.SetHeader([]string{"Setting", "Value"})
        receiptTable.SetBorder(false)
        receiptTable.SetColumnSeparator(" | ")
        receiptTable.Append([]string{"Receipt Number Prefix", settings.Receipt.ReceiptNumberPrefix})
        receiptTable.Append([]string{"Print Receipt by Default", fmt.Sprintf("%t", settings.Receipt.PrintReceiptByDefault)})
        receiptTable.Append([]string{"Email Receipt by Default", fmt.Sprintf("%t", settings.Receipt.EmailReceiptByDefault)})
        receiptTable.Append([]string{"Show Tax Details", fmt.Sprintf("%t", settings.Receipt.ShowTaxDetails)})
        receiptTable.Append([]string{"Show Discount Details", fmt.Sprintf("%t", settings.Receipt.ShowDiscountDetails)})
        receiptTable.Append([]string{"Show Payment Details", fmt.Sprintf("%t", settings.Receipt.ShowPaymentDetails)})
        receiptTable.Render()
        fmt.Println()

        // Print backup settings
        fmt.Println("=== Backup Settings ===")
        backupTable := tablewriter.NewWriter(os.Stdout)
        backupTable.SetHeader([]string{"Setting", "Value"})
        backupTable.SetBorder(false)
        backupTable.SetColumnSeparator(" | ")
        backupTable.Append([]string{"Auto Backup", fmt.Sprintf("%t", settings.Backup.AutoBackupEnabled)})
        backupTable.Append([]string{"Backup Interval (hours)", fmt.Sprintf("%d", settings.Backup.BackupInterval)})
        backupTable.Append([]string{"Backup Path", settings.Backup.BackupPath})
        backupTable.Append([]string{"Keep Backup Count", fmt.Sprintf("%d", settings.Backup.KeepBackupCount)})
        if settings.Backup.LastBackupTime != "" {
                backupTable.Append([]string{"Last Backup", settings.Backup.LastBackupTime})
        }
        backupTable.Render()
        fmt.Println()

        // Print system settings
        fmt.Println("=== System Settings ===")
        systemTable := tablewriter.NewWriter(os.Stdout)
        systemTable.SetHeader([]string{"Setting", "Value"})
        systemTable.SetBorder(false)
        systemTable.SetColumnSeparator(" | ")
        systemTable.Append([]string{"Language", settings.System.Language})
        systemTable.Append([]string{"Currency", settings.System.Currency})
        systemTable.Append([]string{"Currency Symbol", settings.System.CurrencySymbol})
        systemTable.Append([]string{"Date Format", settings.System.DateFormat})
        systemTable.Append([]string{"Time Format", settings.System.TimeFormat})
        systemTable.Append([]string{"Default Operating Mode", settings.System.DefaultOperatingMode})
        systemTable.Render()
        fmt.Println()

        // Print last updated info
        fmt.Printf("Last Updated: %s", settings.LastUpdated)
        if settings.LastUpdatedBy != "" {
                fmt.Printf(" by %s", settings.LastUpdatedBy)
        }
        fmt.Println()
}

// updateSettingsByKey updates a specific setting by key path using dot notation
func updateSettingsByKey(settings models.Settings, key string, value string) (models.Settings, error) {
        // Convert settings to a map for easier manipulation
        settingsJSON, err := settings.ExportToJSON()
        if err != nil {
                return settings, err
        }

        var settingsMap map[string]interface{}
        err = json.Unmarshal([]byte(settingsJSON), &settingsMap)
        if err != nil {
                return settings, err
        }

        // Split the key path
        parts := strings.Split(key, ".")
        if len(parts) < 2 {
                return settings, fmt.Errorf("invalid key format, use 'section.key' format (e.g., 'store.name')")
        }

        // Navigate to the correct section
        section, found := settingsMap[parts[0]]
        if !found {
                return settings, fmt.Errorf("section '%s' not found", parts[0])
        }

        sectionMap, ok := section.(map[string]interface{})
        if !ok {
                return settings, fmt.Errorf("section '%s' is not a valid settings section", parts[0])
        }

        // Update the value
        fieldKey := parts[1]
        
        // Convert value to appropriate type based on current type
        currentVal, exists := sectionMap[fieldKey]
        if !exists {
                return settings, fmt.Errorf("key '%s' not found in section '%s'", fieldKey, parts[0])
        }

        var newValue interface{} = value
        
        switch currentVal.(type) {
        case float64:
                floatVal, err := parseFloat(value)
                if err != nil {
                        return settings, fmt.Errorf("invalid number format: %v", err)
                }
                newValue = floatVal
        case int:
                intVal, err := parseInt(value)
                if err != nil {
                        return settings, fmt.Errorf("invalid integer format: %v", err)
                }
                newValue = intVal
        case bool:
                boolVal, err := parseBool(value)
                if err != nil {
                        return settings, fmt.Errorf("invalid boolean format (use 'true' or 'false'): %v", err)
                }
                newValue = boolVal
        case []interface{}:
                // Assume comma-separated list for arrays
                items := strings.Split(value, ",")
                trimmedItems := make([]interface{}, len(items))
                for i, item := range items {
                        trimmedItems[i] = strings.TrimSpace(item)
                }
                newValue = trimmedItems
        }

        // Update the map
        sectionMap[fieldKey] = newValue
        settingsMap[parts[0]] = sectionMap

        // Convert back to settings object
        updatedJSON, err := json.Marshal(settingsMap)
        if err != nil {
                return settings, err
        }

        updatedSettings, err := models.ImportFromJSON(string(updatedJSON))
        if err != nil {
                return settings, err
        }

        // Preserve ID and other metadata
        updatedSettings.ID = settings.ID
        updatedSettings.LastUpdated = time.Now().Format(time.RFC3339)

        return updatedSettings, nil
}

// Helper functions for type conversion
func parseFloat(s string) (float64, error) {
        // Remove % sign if present
        s = strings.TrimSuffix(s, "%")
        var f float64
        _, err := fmt.Sscanf(s, "%f", &f)
        return f, err
}

func parseInt(s string) (int, error) {
        var i int
        _, err := fmt.Sscanf(s, "%d", &i)
        return i, err
}

func parseBool(s string) (bool, error) {
        lower := strings.ToLower(s)
        if lower == "true" || lower == "yes" || lower == "1" || lower == "on" {
                return true, nil
        } else if lower == "false" || lower == "no" || lower == "0" || lower == "off" {
                return false, nil
        }
        return false, fmt.Errorf("invalid boolean value")
}