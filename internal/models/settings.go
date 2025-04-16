package models

import (
        "encoding/json"
        "fmt"
        "time"
)

// StoreInfo contains details about the store
type StoreInfo struct {
        Name            string `json:"name"`
        Address         string `json:"address,omitempty"`
        Phone           string `json:"phone,omitempty"`
        Email           string `json:"email,omitempty"`
        Website         string `json:"website,omitempty"`
        TaxID           string `json:"tax_id,omitempty"`
        RegistrationNum string `json:"registration_num,omitempty"`
        Logo            string `json:"logo_path,omitempty"` // Path to logo file
        ReceiptFooter   string `json:"receipt_footer,omitempty"`
        ReceiptHeader   string `json:"receipt_header,omitempty"`
}

// TaxSettings contains tax configuration
type TaxSettings struct {
        DefaultTaxRate    float64            `json:"default_tax_rate"`
        TaxRatesByProduct map[int]float64    `json:"tax_rates_by_product,omitempty"`  // Product ID -> tax rate
        TaxRatesByCategory map[int]float64   `json:"tax_rates_by_category,omitempty"` // Category ID -> tax rate
        TaxInclusive      bool               `json:"tax_inclusive"`                    // Whether prices include tax
}

// ProductSettings contains product configuration
type ProductSettings struct {
        DefaultCategory        int    `json:"default_category_id"`
        DefaultSupplier        int    `json:"default_supplier_id"`
        LowStockThreshold      int    `json:"low_stock_threshold"`
        EnableBatchTracking    bool   `json:"enable_batch_tracking"`
        EnableExpiryTracking   bool   `json:"enable_expiry_tracking"`
        EnableLocationTracking bool   `json:"enable_location_tracking"`
        SKUPrefix              string `json:"sku_prefix,omitempty"`
}

// PaymentSettings contains payment configuration
type PaymentSettings struct {
        EnabledPaymentMethods []string          `json:"enabled_payment_methods"`
        DefaultPaymentMethod  string            `json:"default_payment_method"`
        PaymentGateways       map[string]string `json:"payment_gateways,omitempty"` // Gateway name -> config
}

// ReceiptSettings contains receipt configuration
type ReceiptSettings struct {
        ReceiptNumberPrefix  string `json:"receipt_number_prefix"`
        PrintReceiptByDefault bool   `json:"print_receipt_by_default"`
        EmailReceiptByDefault bool   `json:"email_receipt_by_default"`
        ShowTaxDetails       bool   `json:"show_tax_details"`
        ShowDiscountDetails  bool   `json:"show_discount_details"`
        ShowPaymentDetails   bool   `json:"show_payment_details"`
}

// BackupSettings contains backup configuration
type BackupSettings struct {
        AutoBackupEnabled     bool   `json:"auto_backup_enabled"`
        BackupInterval        int    `json:"backup_interval_hours"`
        BackupPath            string `json:"backup_path"`
        KeepBackupCount       int    `json:"keep_backup_count"`
        LastBackupTime        string `json:"last_backup_time,omitempty"`
}

// SystemSettings contains system configuration
type SystemSettings struct {
        Language             string `json:"language"`
        Currency             string `json:"currency"`
        CurrencySymbol       string `json:"currency_symbol"`
        DateFormat           string `json:"date_format"`
        TimeFormat           string `json:"time_format"`
        DefaultOperatingMode string `json:"default_operating_mode"`
}

// Settings represents all POS settings
type Settings struct {
        ID              int             `json:"id"`
        Store           StoreInfo       `json:"store"`
        Tax             TaxSettings     `json:"tax"`
        Product         ProductSettings `json:"product"`
        Payment         PaymentSettings `json:"payment"`
        Receipt         ReceiptSettings `json:"receipt"`
        Backup          BackupSettings  `json:"backup"`
        System          SystemSettings  `json:"system"`
        LastUpdated     string          `json:"last_updated"`
        LastUpdatedBy   string          `json:"last_updated_by,omitempty"`
}

// Validate checks if the settings are valid
func (s *Settings) Validate() error {
        // Basic validation, more validation can be added as needed
        if s.Store.Name == "" {
                return fmt.Errorf("store name cannot be empty")
        }
        return nil
}

// NewDefaultSettings creates a new settings object with default values
func NewDefaultSettings() Settings {
        now := time.Now().Format(time.RFC3339)
        return Settings{
                Store: StoreInfo{
                        Name:          "My Store",
                        ReceiptFooter: "Thank you for your purchase!",
                },
                Tax: TaxSettings{
                        DefaultTaxRate:    8.0,  // 8%
                        TaxInclusive:      false,
                        TaxRatesByProduct: make(map[int]float64),
                        TaxRatesByCategory: make(map[int]float64),
                },
                Product: ProductSettings{
                        DefaultCategory:        1, // Default category ID
                        DefaultSupplier:        1, // Default supplier ID
                        LowStockThreshold:      5,
                        EnableBatchTracking:    false,
                        EnableExpiryTracking:   false,
                        EnableLocationTracking: false,
                },
                Payment: PaymentSettings{
                        EnabledPaymentMethods: []string{"cash", "card", "mobile"},
                        DefaultPaymentMethod:  "cash",
                        PaymentGateways:       make(map[string]string),
                },
                Receipt: ReceiptSettings{
                        ReceiptNumberPrefix:   "RCP-",
                        PrintReceiptByDefault: false,
                        EmailReceiptByDefault: false,
                        ShowTaxDetails:        true,
                        ShowDiscountDetails:   true,
                        ShowPaymentDetails:    true,
                },
                Backup: BackupSettings{
                        AutoBackupEnabled: false,
                        BackupInterval:    24, // Once per day
                        BackupPath:        "./backups",
                        KeepBackupCount:   7,  // Keep last 7 backups
                },
                System: SystemSettings{
                        Language:             "en",
                        Currency:             "USD",
                        CurrencySymbol:       "$",
                        DateFormat:           "2006-01-02",
                        TimeFormat:           "15:04:05",
                        DefaultOperatingMode: "classic",
                },
                LastUpdated: now,
        }
}

// ExportToJSON exports settings to a JSON string
func (s *Settings) ExportToJSON() (string, error) {
        data, err := json.MarshalIndent(s, "", "  ")
        if err != nil {
                return "", err
        }
        return string(data), nil
}

// ImportFromJSON imports settings from a JSON string
func ImportFromJSON(jsonStr string) (Settings, error) {
        var settings Settings
        err := json.Unmarshal([]byte(jsonStr), &settings)
        if err != nil {
                return Settings{}, err
        }
        return settings, nil
}