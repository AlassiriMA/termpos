package db

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "strings"
        "time"

        "termpos/internal/models"
)

// GetSettings retrieves all settings from the database
func GetSettings() (models.Settings, error) {
        var settingsJSON string
        var lastUpdated string
        var lastUpdatedBy sql.NullString
        var id int

        query := `SELECT id, settings_json, last_updated, last_updated_by FROM settings ORDER BY id DESC LIMIT 1`
        err := DB.QueryRow(query).Scan(&id, &settingsJSON, &lastUpdated, &lastUpdatedBy)
        
        if err != nil {
                if err == sql.ErrNoRows {
                        // No settings found, create default
                        return initDefaultSettings()
                }
                return models.Settings{}, fmt.Errorf("failed to get settings: %w", err)
        }

        var settings models.Settings
        err = json.Unmarshal([]byte(settingsJSON), &settings)
        if err != nil {
                return models.Settings{}, fmt.Errorf("failed to parse settings JSON: %w", err)
        }

        settings.ID = id
        settings.LastUpdated = lastUpdated
        if lastUpdatedBy.Valid {
                settings.LastUpdatedBy = lastUpdatedBy.String
        }

        return settings, nil
}

// SaveSettings saves settings to the database
func SaveSettings(settings models.Settings, username string) error {
        // First validate the settings
        if err := settings.Validate(); err != nil {
                return err
        }

        // Convert settings to JSON
        settingsJSON, err := json.Marshal(settings)
        if err != nil {
                return fmt.Errorf("failed to marshal settings to JSON: %w", err)
        }

        // Update last updated time
        now := time.Now().Format(time.RFC3339)

        // Use insert or update logic
        query := `
                INSERT INTO settings (settings_json, last_updated, last_updated_by) 
                VALUES (?, ?, ?) 
                RETURNING id
        `
        
        var id int
        err = DB.QueryRow(query, string(settingsJSON), now, username).Scan(&id)
        if err != nil {
                return fmt.Errorf("failed to save settings: %w", err)
        }

        return nil
}

// initDefaultSettings initializes the default settings in the database
func initDefaultSettings() (models.Settings, error) {
        settings := models.NewDefaultSettings()
        
        // Convert settings to JSON
        settingsJSON, err := json.Marshal(settings)
        if err != nil {
                return models.Settings{}, fmt.Errorf("failed to marshal default settings to JSON: %w", err)
        }

        // Insert into database
        query := `
                INSERT INTO settings (settings_json, last_updated, last_updated_by) 
                VALUES (?, ?, ?) 
                RETURNING id
        `
        
        var id int
        err = DB.QueryRow(query, string(settingsJSON), settings.LastUpdated, "system").Scan(&id)
        if err != nil {
                return models.Settings{}, fmt.Errorf("failed to initialize default settings: %w", err)
        }

        settings.ID = id
        return settings, nil
}

// GetSettingValue retrieves a specific setting by key path
func GetSettingValue(keyPath string) (string, error) {
        settings, err := GetSettings()
        if err != nil {
                return "", err
        }

        // Convert settings to JSON
        settingsJSON, err := json.Marshal(settings)
        if err != nil {
                return "", fmt.Errorf("failed to marshal settings to JSON: %w", err)
        }

        // Parse as generic JSON to extract values by path
        var settingsMap map[string]interface{}
        err = json.Unmarshal(settingsJSON, &settingsMap)
        if err != nil {
                return "", fmt.Errorf("failed to parse settings JSON: %w", err)
        }

        // Extract value by key path (simple implementation)
        parts := []string{keyPath}
        current := settingsMap
        for i, part := range parts {
                if i == len(parts)-1 {
                        if value, ok := current[part]; ok {
                                // Convert value to string
                                if str, ok := value.(string); ok {
                                        return str, nil
                                }
                                // Try JSON marshal for non-string values
                                valueJSON, err := json.Marshal(value)
                                if err != nil {
                                        return "", fmt.Errorf("failed to convert setting value to string: %w", err)
                                }
                                return string(valueJSON), nil
                        }
                        return "", fmt.Errorf("setting key not found: %s", keyPath)
                }
                
                nextMap, ok := current[part].(map[string]interface{})
                if !ok {
                        return "", fmt.Errorf("invalid key path: %s", keyPath)
                }
                current = nextMap
        }

        return "", fmt.Errorf("setting key not found: %s", keyPath)
}

// BackupDatabaseViaVacuum creates a backup of the database using SQLite's VACUUM INTO command
// This is an alternative method to the file-based approach in db.go
func BackupDatabaseViaVacuum(backupPath string) error {
        // If a full path with filename is provided, use it directly
        isFullPath := strings.HasSuffix(backupPath, ".db")
        
        // If it's not a full path, treat it as a directory
        if !isFullPath {
                if backupPath == "" {
                        settings, err := GetSettings()
                        if err != nil {
                                return err
                        }
                        backupPath = settings.Backup.BackupPath
                }

                // Ensure backup directory exists
                err := ensureBackupDir(backupPath)
                if err != nil {
                        return err
                }

                // Create backup timestamp and filename
                timestamp := time.Now().Format("20060102_150405")
                backupFileName := filepath.Join(backupPath, fmt.Sprintf("pos_backup_%s.db", timestamp))
                backupPath = backupFileName
        } else {
                // Ensure the directory for the full path exists
                dir := filepath.Dir(backupPath)
                if err := ensureBackupDir(dir); err != nil {
                        return err
                }
        }

        // Check if path exists as a directory
        if fileInfo, err := os.Stat(backupPath); err == nil && fileInfo.IsDir() {
                return fmt.Errorf("backup path %s is a directory; expected a file path", backupPath)
        }

        // Use SQLite's backup mechanism
        backup := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
        _, err := DB.Exec(backup)
        if err != nil {
                return fmt.Errorf("failed to create database backup: %w", err)
        }

        // Update last backup time in settings
        settings, err := GetSettings()
        if err != nil {
                return err
        }
        settings.Backup.LastBackupTime = time.Now().Format(time.RFC3339)
        
        return SaveSettings(settings, "system")
}

// ensureBackupDir ensures that the backup directory exists
func ensureBackupDir(path string) error {
        // Create the directory if it doesn't exist
        if _, err := os.Stat(path); os.IsNotExist(err) {
                // Create directory with all parents if needed
                if err := os.MkdirAll(path, 0755); err != nil {
                        return fmt.Errorf("failed to create backup directory '%s': %w", path, err)
                }
        }
        return nil
}

// ExportSettings exports all settings to a JSON file
func ExportSettings(filePath string) error {
        settings, err := GetSettings()
        if err != nil {
                return err
        }

        jsonData, err := settings.ExportToJSON()
        if err != nil {
                return err
        }

        // Ensure directory exists
        dir := filepath.Dir(filePath)
        if err := ensureBackupDir(dir); err != nil {
                return err
        }

        // Write the settings to a file
        if err := os.WriteFile(filePath, []byte(jsonData), 0644); err != nil {
                return fmt.Errorf("failed to write settings to file: %w", err)
        }

        fmt.Printf("Settings exported to: %s\n", filePath)
        return nil
}

// ImportSettings imports settings from a JSON file
func ImportSettings(jsonData string, username string) error {
        settings, err := models.ImportFromJSON(jsonData)
        if err != nil {
                return fmt.Errorf("failed to parse settings JSON: %w", err)
        }

        // Validate settings
        if err := settings.Validate(); err != nil {
                return err
        }

        // Save to database
        return SaveSettings(settings, username)
}