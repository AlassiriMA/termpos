package main

import (
        "bufio"
        "encoding/json"
        "fmt"
        "io/ioutil"
        "os"
        "path/filepath"
        "strings"

        "github.com/olekukonko/tablewriter"
        "github.com/spf13/cobra"
)

// Configuration file structure
type Config struct {
        System struct {
                DefaultMode       string `json:"default_mode"`
                ServerPort        int    `json:"server_port"`
                LogLevel          string `json:"log_level"`
                DBPath            string `json:"db_path"`
                BackupPath        string `json:"backup_path"`
                BackupFrequency   string `json:"backup_frequency"`
                BackupRetentionDays int   `json:"backup_retention_days"`
                BackupEncrypt     bool   `json:"backup_encrypt"`
        } `json:"system"`
        Business struct {
                Name     string  `json:"name"`
                Address  string  `json:"address"`
                City     string  `json:"city"`
                State    string  `json:"state"`
                Zip      string  `json:"zip"`
                Country  string  `json:"country"`
                Phone    string  `json:"phone"`
                Email    string  `json:"email"`
                Website  string  `json:"website"`
                Currency string  `json:"currency"`
                TaxRate  float64 `json:"tax_rate"`
                Timezone string  `json:"timezone"`
        } `json:"business"`
        Features struct {
                EnableLoyaltyProgram bool `json:"enable_loyalty_program"`
                EnableStaffManagement bool `json:"enable_staff_management"`
                EnableAuditLogging   bool `json:"enable_audit_logging"`
                EnableEmailReceipts  bool `json:"enable_email_receipts"`
                EnableLowStockAlerts bool `json:"enable_low_stock_alerts"`
        } `json:"features"`
        Security struct {
                MinPasswordLength      int  `json:"min_password_length"`
                PasswordRequireMixedCase bool `json:"password_require_mixed_case"`
                PasswordRequireNumber  bool `json:"password_require_number"`
                PasswordRequireSpecial bool `json:"password_require_special"`
                SessionTimeoutMinutes  int  `json:"session_timeout_minutes"`
                MaxLoginAttempts      int  `json:"max_login_attempts"`
        } `json:"security"`
        Integrations struct {
                Email struct {
                        Enabled   bool   `json:"enabled"`
                        Provider  string `json:"provider"`
                        Host      string `json:"host"`
                        Port      int    `json:"port"`
                        Username  string `json:"username"`
                        Password  string `json:"password"`
                        FromEmail string `json:"from_email"`
                        FromName  string `json:"from_name"`
                } `json:"email"`
                PaymentGateway struct {
                        Enabled       bool   `json:"enabled"`
                        Provider      string `json:"provider"`
                        ApiKey        string `json:"api_key"`
                        WebhookSecret string `json:"webhook_secret"`
                } `json:"payment_gateway"`
                Accounting struct {
                        Enabled   bool   `json:"enabled"`
                        Provider  string `json:"provider"`
                        ApiKey    string `json:"api_key"`
                        CompanyId string `json:"company_id"`
                } `json:"accounting"`
        } `json:"integrations"`
}

// Default configuration values
func getDefaultConfig() Config {
        config := Config{}
        
        // System defaults
        config.System.DefaultMode = "classic"
        config.System.ServerPort = 8000
        config.System.LogLevel = "info"
        config.System.DBPath = "./pos.db"
        config.System.BackupPath = "./backups"
        config.System.BackupFrequency = "daily"
        config.System.BackupRetentionDays = 7
        config.System.BackupEncrypt = true
        
        // Business defaults (empty strings)
        config.Business.Currency = "USD"
        config.Business.TaxRate = 8.0
        config.Business.Timezone = "UTC"
        
        // Features defaults
        config.Features.EnableLoyaltyProgram = true
        config.Features.EnableStaffManagement = true
        config.Features.EnableAuditLogging = true
        config.Features.EnableEmailReceipts = false
        config.Features.EnableLowStockAlerts = true
        
        // Security defaults
        config.Security.MinPasswordLength = 8
        config.Security.PasswordRequireMixedCase = true
        config.Security.PasswordRequireNumber = true
        config.Security.PasswordRequireSpecial = true
        config.Security.SessionTimeoutMinutes = 30
        config.Security.MaxLoginAttempts = 5
        
        // Integrations defaults
        config.Integrations.Email.Enabled = false
        config.Integrations.Email.Provider = "smtp"
        config.Integrations.Email.Port = 587
        
        config.Integrations.PaymentGateway.Enabled = false
        config.Integrations.Accounting.Enabled = false
        
        return config
}

// readInput reads user input with a prompt and optional default value
func readInput(reader *bufio.Reader, prompt string, defaultValue string) string {
        displayPrompt := prompt
        if defaultValue != "" {
                displayPrompt = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
        }
        
        fmt.Print(displayPrompt)
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)
        
        if input == "" {
                return defaultValue
        }
        return input
}

// readBoolInput reads a boolean input (y/n) with a prompt and default value
func readBoolInput(reader *bufio.Reader, prompt string, defaultValue bool) bool {
        defaultStr := "n"
        if defaultValue {
                defaultStr = "y"
        }
        
        displayPrompt := fmt.Sprintf("%s (y/n) [%s]: ", prompt, defaultStr)
        fmt.Print(displayPrompt)
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(strings.ToLower(input))
        
        if input == "" {
                return defaultValue
        }
        
        return input == "y" || input == "yes"
}

// runInitCommand is the initialization setup wizard
func runInitCommand(cmd *cobra.Command, args []string) error {
        fmt.Println(blue("╔══════════════════════════════════════════════════════════╗"))
        fmt.Println(blue("║                 TermPOS Setup Wizard                     ║"))
        fmt.Println(blue("╚══════════════════════════════════════════════════════════╝"))
        fmt.Println()
        fmt.Println(cyan("This wizard will help you set up your TermPOS system."))
        fmt.Println(cyan("Press Enter to accept the default values (shown in [brackets])"))
        fmt.Println()
        
        configDir := "./config"
        configFile := filepath.Join(configDir, "config.json")
        
        // Check if config file already exists
        if _, err := os.Stat(configFile); err == nil {
                fmt.Println(yellow("Configuration file already exists at: " + configFile))
                reader := bufio.NewReader(os.Stdin)
                overwrite := readBoolInput(reader, "Do you want to overwrite it?", false)
                if !overwrite {
                        fmt.Println(green("Setup canceled. Using existing configuration."))
                        return nil
                }
        }
        
        // Ensure config directory exists
        if _, err := os.Stat(configDir); os.IsNotExist(err) {
                os.MkdirAll(configDir, 0755)
                fmt.Println(green("Created configuration directory: " + configDir))
        }
        
        // Ensure data directory exists
        dataDir := "./data"
        if _, err := os.Stat(dataDir); os.IsNotExist(err) {
                os.MkdirAll(dataDir, 0755)
                fmt.Println(green("Created data directory: " + dataDir))
        }
        
        // Ensure backups directory exists
        backupsDir := "./backups"
        if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
                os.MkdirAll(backupsDir, 0755)
                fmt.Println(green("Created backups directory: " + backupsDir))
        }
        
        // Get default configuration
        config := getDefaultConfig()
        reader := bufio.NewReader(os.Stdin)
        
        // Business Information
        fmt.Println(blue("\n--- Business Information ---"))
        config.Business.Name = readInput(reader, "Business name", config.Business.Name)
        config.Business.Address = readInput(reader, "Street address", config.Business.Address)
        config.Business.City = readInput(reader, "City", config.Business.City)
        config.Business.State = readInput(reader, "State/Province", config.Business.State)
        config.Business.Zip = readInput(reader, "ZIP/Postal code", config.Business.Zip)
        config.Business.Country = readInput(reader, "Country", config.Business.Country)
        config.Business.Phone = readInput(reader, "Phone number", config.Business.Phone)
        config.Business.Email = readInput(reader, "Email address", config.Business.Email)
        config.Business.Website = readInput(reader, "Website", config.Business.Website)
        
        // System Configuration
        fmt.Println(blue("\n--- System Configuration ---"))
        defaultModeInput := readInput(reader, "Default mode (classic, agent, assistant)", config.System.DefaultMode)
        if defaultModeInput == "classic" || defaultModeInput == "agent" || defaultModeInput == "assistant" {
                config.System.DefaultMode = defaultModeInput
        }
        config.Features.EnableLoyaltyProgram = readBoolInput(reader, "Enable loyalty program", config.Features.EnableLoyaltyProgram)
        config.Features.EnableStaffManagement = readBoolInput(reader, "Enable staff management", config.Features.EnableStaffManagement)
        config.Features.EnableAuditLogging = readBoolInput(reader, "Enable audit logging", config.Features.EnableAuditLogging)
        config.Features.EnableLowStockAlerts = readBoolInput(reader, "Enable low stock alerts", config.Features.EnableLowStockAlerts)
        
        // Save configuration to file
        configJson, err := json.MarshalIndent(config, "", "  ")
        if err != nil {
                return fmt.Errorf("error serializing configuration: %v", err)
        }
        
        err = ioutil.WriteFile(configFile, configJson, 0644)
        if err != nil {
                return fmt.Errorf("error saving configuration: %v", err)
        }
        
        fmt.Println(green("\nConfiguration saved to: " + configFile))
        
        // Display summary
        fmt.Println(blue("\n--- Configuration Summary ---"))
        
        table := tablewriter.NewWriter(os.Stdout)
        table.SetHeader([]string{"Setting", "Value"})
        table.SetBorder(false)
        table.SetColumnColor(tablewriter.Colors{tablewriter.FgCyanColor}, tablewriter.Colors{tablewriter.FgWhiteColor})
        
        table.Append([]string{"Business Name", config.Business.Name})
        table.Append([]string{"Default Mode", config.System.DefaultMode})
        table.Append([]string{"Database Path", config.System.DBPath})
        table.Append([]string{"Backup Path", config.System.BackupPath})
        table.Append([]string{"Loyalty Program", fmt.Sprintf("%t", config.Features.EnableLoyaltyProgram)})
        table.Append([]string{"Staff Management", fmt.Sprintf("%t", config.Features.EnableStaffManagement)})
        table.Append([]string{"Audit Logging", fmt.Sprintf("%t", config.Features.EnableAuditLogging)})
        table.Append([]string{"Low Stock Alerts", fmt.Sprintf("%t", config.Features.EnableLowStockAlerts)})
        
        table.Render()
        
        fmt.Println(green("\nSetup completed successfully!"))
        fmt.Println(cyan("\nTo start the application:"))
        fmt.Printf("  %s\n", yellow("./termpos"))
        fmt.Printf("  %s\n", yellow("./termpos --mode agent --port 8000"))
        
        return nil
}

// initCmd represents the init command for first-time setup
var initCmd = &cobra.Command{
        Use:   "init",
        Short: "Initialize the POS system with a setup wizard",
        Long: `Run the interactive setup wizard to configure the POS system for first use.
This will help you set up business information, system preferences, and integration options.
If a configuration file already exists, you'll be asked if you want to overwrite it.`,
        RunE: runInitCommand,
}

func init() {
        rootCmd.AddCommand(initCmd)
}