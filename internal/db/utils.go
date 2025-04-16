package db

import (
        "encoding/json"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "time"
)

// writeJSONToFile writes a JSON-serializable object to a file
func writeJSONToFile(data interface{}, filePath string) error {
        // Create the output directory if it doesn't exist
        dir := filepath.Dir(filePath)
        if dir != "." && dir != "" {
                if err := os.MkdirAll(dir, 0755); err != nil {
                        return fmt.Errorf("failed to create directory: %w", err)
                }
        }

        // Marshal the data with indentation for readability
        jsonData, err := json.MarshalIndent(data, "", "  ")
        if err != nil {
                return fmt.Errorf("failed to marshal data to JSON: %w", err)
        }

        // Write to the file
        err = os.WriteFile(filePath, jsonData, 0644)
        if err != nil {
                return fmt.Errorf("failed to write to file: %w", err)
        }

        return nil
}

// RestoreDatabase restores the database from a backup file
func RestoreDatabase(backupPath string) error {
        // Check if the backup file exists
        if _, err := os.Stat(backupPath); os.IsNotExist(err) {
                return fmt.Errorf("backup file does not exist: %s", backupPath)
        }

        // Get the current database path
        dbPath := GetDatabasePath()

        // Close the current database connection
        if err := CloseDB(); err != nil {
                return fmt.Errorf("failed to close database connection: %w", err)
        }

        // Rename the current database file to a backup
        tempBackup := dbPath + ".backup_before_restore." + time.Now().Format("20060102_150405")
        if err := os.Rename(dbPath, tempBackup); err != nil {
                // Try to reopen the database
                if initErr := Initialize(dbPath); initErr != nil {
                        fmt.Printf("Warning: Failed to reopen database after restore attempt: %v\n", initErr)
                }
                return fmt.Errorf("failed to create backup of current database: %w", err)
        }

        // Copy the backup to the current database location
        srcFile, err := os.Open(backupPath)
        if err != nil {
                // Restore the original database
                os.Rename(tempBackup, dbPath)
                // Try to reopen the database
                if initErr := Initialize(dbPath); initErr != nil {
                        fmt.Printf("Warning: Failed to reopen database after restore attempt: %v\n", initErr)
                }
                return fmt.Errorf("failed to open backup file: %w", err)
        }
        defer srcFile.Close()

        // Create the destination file
        dstFile, err := os.Create(dbPath)
        if err != nil {
                // Restore the original database
                os.Rename(tempBackup, dbPath)
                // Try to reopen the database
                if initErr := Initialize(dbPath); initErr != nil {
                        fmt.Printf("Warning: Failed to reopen database after restore attempt: %v\n", initErr)
                }
                return fmt.Errorf("failed to create new database file: %w", err)
        }
        defer dstFile.Close()

        // Copy the backup to the database file
        if _, err := io.Copy(dstFile, srcFile); err != nil {
                // Restore the original database
                dstFile.Close() // Close before rename
                os.Rename(tempBackup, dbPath)
                // Try to reopen the database
                if initErr := Initialize(dbPath); initErr != nil {
                        fmt.Printf("Warning: Failed to reopen database after restore attempt: %v\n", initErr)
                }
                return fmt.Errorf("failed to copy backup to database: %w", err)
        }

        // Reopen the database
        if err := Initialize(dbPath); err != nil {
                // Try to restore the original database
                os.Rename(tempBackup, dbPath)
                if initErr := Initialize(dbPath); initErr != nil {
                        return fmt.Errorf("failed to restore original database after restore failure: %v (original error: %w)", initErr, err)
                }
                return fmt.Errorf("failed to initialize restored database: %w", err)
        }

        // Log the restore operation
        fmt.Printf("Database restored successfully from %s\n", backupPath)
        fmt.Printf("Original database backed up at %s\n", tempBackup)

        return nil
}