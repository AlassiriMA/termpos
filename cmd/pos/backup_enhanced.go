package main

import (
        "fmt"
        "os"
        "path/filepath"
        "strings"
        "time"

        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/security"
)

var (
        // Enhanced backup command flags
        encryptBackup    bool
        backupPassword   string
        backupRotate     bool
        backupCompress   bool
        backupVerify     bool
        backupName       string
        backupRetention  int
        restoreFromPath  string
        scheduleBackup   bool
        scheduleInterval int
)

// enhancedBackupCmd represents the enhanced backup command
var enhancedBackupCmd = &cobra.Command{
        Use:   "backup-enhanced [path]",
        Short: "Create an enhanced backup with encryption",
        Long: `Create a backup of the POS database with encryption and enhanced options.
This command allows for database encryption, compression, verification, 
and rotation of backups.`,
        Args: cobra.MaximumNArgs(1),
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
                } else {
                        // Get from settings if not specified
                        settings, err := db.GetSettings()
                        if err == nil && settings.Backup.BackupPath != "" {
                                backupPath = settings.Backup.BackupPath
                        } else {
                                backupPath = "./backups" // Default fallback
                        }
                }

                // Ensure backup directory exists
                if err := os.MkdirAll(backupPath, 0755); err != nil {
                        fmt.Printf("Error creating backup directory: %v\n", err)
                        return
                }

                // Generate backup filename with timestamp
                timestamp := time.Now().Format("20060102_150405")
                if backupName == "" {
                        backupName = fmt.Sprintf("pos_backup_%s.db", timestamp)
                } else if !filepath.IsAbs(backupName) && !strings.HasSuffix(backupName, ".db") {
                        backupName = backupName + ".db"
                }

                backupFilepath := filepath.Join(backupPath, backupName)

                // Initialize encryption if needed
                if encryptBackup {
                        // If no specific password is provided, initialize with environment variable
                        if backupPassword == "" {
                                if err := security.InitEncryption(); err != nil {
                                        fmt.Printf("Warning: Failed to initialize encryption: %v\n", err)
                                        fmt.Println("Proceeding with unencrypted backup")
                                        encryptBackup = false
                                }
                        }
                }

                // Check if path already exists as a directory
                if stat, err := os.Stat(backupFilepath); err == nil && stat.IsDir() {
                        // Remove directory if it exists (this can happen due to a bug)
                        if err := os.RemoveAll(backupFilepath); err != nil {
                                fmt.Printf("Error removing existing directory: %v\n", err)
                                return
                        }
                }

                // Perform backup
                fmt.Println("Creating backup...")
                err := db.BackupDatabaseViaVacuum(backupFilepath)
                if err != nil {
                        fmt.Printf("Error creating backup: %v\n", err)
                        return
                }

                // Encrypt backup if requested
                if encryptBackup {
                        fmt.Println("Encrypting backup...")
                        err = encryptBackupFile(backupFilepath, backupPassword)
                        if err != nil {
                                fmt.Printf("Error encrypting backup: %v\n", err)
                                // Continue with unencrypted backup
                        } else {
                                // Rename to indicate encryption
                                encryptedPath := backupFilepath + ".enc"
                                if err := os.Rename(backupFilepath+".tmp", encryptedPath); err != nil {
                                        fmt.Printf("Error finalizing encrypted backup: %v\n", err)
                                } else {
                                        backupFilepath = encryptedPath
                                }
                        }
                }

                // Compress backup if requested
                if backupCompress {
                        fmt.Println("Compressing backup...")
                        // Compression logic would go here
                        // For now, we'll just simulate it with a message
                        fmt.Println("Compression not yet implemented")
                }

                // Verify backup if requested
                if backupVerify {
                        fmt.Println("Verifying backup...")
                        if err := verifyBackup(backupFilepath, encryptBackup); err != nil {
                                fmt.Printf("Backup verification failed: %v\n", err)
                        } else {
                                fmt.Println("Backup verified successfully")
                        }
                }

                // Rotate backups if requested
                if backupRotate {
                        if backupRetention <= 0 {
                                // Get retention from settings
                                settings, err := db.GetSettings()
                                if err == nil && settings.Backup.KeepBackupCount > 0 {
                                        backupRetention = settings.Backup.KeepBackupCount
                                } else {
                                        backupRetention = 7 // Default
                                }
                        }

                        fmt.Printf("Rotating backups (keeping %d)...\n", backupRetention)
                        if err := db.CleanupBackups(backupPath, backupRetention); err != nil {
                                fmt.Printf("Error rotating backups: %v\n", err)
                        } else {
                                fmt.Printf("Successfully rotated backups, keeping %d most recent\n", backupRetention)
                        }
                }

                // Schedule automatic backup if requested
                if scheduleBackup {
                        fmt.Println("Configuring scheduled backups...")
                        if scheduleInterval <= 0 {
                                scheduleInterval = 24 // Default to daily
                        }

                        // Update settings to enable automatic backups
                        settings, err := db.GetSettings()
                        if err != nil {
                                fmt.Printf("Error getting settings: %v\n", err)
                        } else {
                                // Update backup settings
                                settings.Backup.AutoBackupEnabled = true
                                settings.Backup.BackupInterval = scheduleInterval
                                settings.Backup.BackupPath = backupPath
                                settings.Backup.KeepBackupCount = backupRetention

                                // Save settings
                                username := "system"
                                if session != nil {
                                        username = session.Username
                                }
                                
                                if err := db.SaveSettings(settings, username); err != nil {
                                        fmt.Printf("Error scheduling backups: %v\n", err)
                                } else {
                                        fmt.Printf("Automatic backups scheduled every %d hours\n", scheduleInterval)
                                }
                        }
                }

                // Update last backup time in settings
                settings, err := db.GetSettings()
                if err == nil {
                        settings.Backup.LastBackupTime = time.Now().Format(time.RFC3339)
                        username := "system"
                        if session != nil {
                                username = session.Username
                        }
                        
                        if err := db.SaveSettings(settings, username); err != nil {
                                fmt.Printf("Warning: Could not update last backup time: %v\n", err)
                        }
                }

                // Log the backup creation in audit log
                if session != nil {
                        description := fmt.Sprintf("Created backup at %s", backupFilepath)
                        additionalInfo := fmt.Sprintf("encrypted=%t,compressed=%t,verified=%t",
                                encryptBackup, backupCompress, backupVerify)
                        
                        db.AddAuditLog(
                                session.Username,
                                db.ActionBackup,
                                "database",
                                "backup",
                                description,
                                "",
                                "",
                                "",
                                additionalInfo,
                        )
                }

                fmt.Printf("Backup created successfully at %s\n", backupFilepath)
        },
}

// encryptBackupFile encrypts a backup file
func encryptBackupFile(filepath string, password string) error {
        // Read the backup file
        data, err := os.ReadFile(filepath)
        if err != nil {
                return fmt.Errorf("failed to read backup file: %w", err)
        }

        // Set custom encryption key if provided
        if password != "" {
                // For simplicity, we're not implementing custom password handling
                // In a real implementation, you would use the password to derive a key
                fmt.Println("Custom password encryption not implemented, using default key")
        }

        // Encrypt the data
        encrypted, err := security.Encrypt(string(data))
        if err != nil {
                return fmt.Errorf("failed to encrypt backup: %w", err)
        }

        // Write to temporary file
        tempFile := filepath + ".tmp"
        err = os.WriteFile(tempFile, []byte(encrypted), 0644)
        if err != nil {
                return fmt.Errorf("failed to write encrypted backup: %w", err)
        }

        return nil
}

// verifyBackup verifies a backup file
func verifyBackup(filepath string, isEncrypted bool) error {
        // For encrypted backups, try to decrypt
        if isEncrypted {
                // Read the encrypted backup
                data, err := os.ReadFile(filepath)
                if err != nil {
                        return fmt.Errorf("failed to read encrypted backup: %w", err)
                }

                // Try to decrypt (this verifies the encryption is valid)
                _, err = security.Decrypt(string(data))
                if err != nil {
                        return fmt.Errorf("backup verification failed (invalid encryption): %w", err)
                }
        }

        // For all backups, check file integrity
        // In a real implementation, you might check SQLite database integrity as well
        fileInfo, err := os.Stat(filepath)
        if err != nil {
                return fmt.Errorf("failed to access backup file: %w", err)
        }

        if fileInfo.Size() == 0 {
                return fmt.Errorf("backup file is empty")
        }

        return nil
}

// restoreBackupCmd restores a database from a backup
var restoreBackupCmd = &cobra.Command{
        Use:   "restore [path]",
        Short: "Restore database from a backup",
        Long:  `Restore the POS database from a backup file, with support for encrypted backups.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:restore") != nil {
                        fmt.Println("Error: You don't have permission to restore backups")
                        return
                }

                backupPath := args[0]

                // Check if file exists
                if _, err := os.Stat(backupPath); os.IsNotExist(err) {
                        fmt.Printf("Error: Backup file %s does not exist\n", backupPath)
                        return
                }

                // Check if it's an encrypted backup
                isEncrypted := false
                if filepath.Ext(backupPath) == ".enc" {
                        isEncrypted = true
                }

                // Confirm restore operation
                fmt.Println("WARNING: Restoring will replace the current database with the backup.")
                fmt.Println("All current data will be lost if not backed up.")
                fmt.Print("Are you sure you want to continue? (y/N): ")
                var confirm string
                fmt.Scanln(&confirm)
                if confirm != "y" && confirm != "Y" {
                        fmt.Println("Restore operation cancelled")
                        return
                }

                // Handle encrypted backup
                if isEncrypted {
                        fmt.Println("Detected encrypted backup file")
                        
                        // Decrypt the backup
                        decryptedPath := backupPath + ".decrypted"
                        if err := decryptBackup(backupPath, decryptedPath, backupPassword); err != nil {
                                fmt.Printf("Error decrypting backup: %v\n", err)
                                return
                        }
                        
                        // Use decrypted file for restore
                        backupPath = decryptedPath
                }

                // Stop database connections before restore
                if err := db.CloseDB(); err != nil {
                        fmt.Printf("Warning: Failed to close database connections: %v\n", err)
                }

                // Perform restore
                fmt.Println("Restoring database from backup...")
                if err := db.RestoreDatabase(backupPath); err != nil {
                        fmt.Printf("Error restoring database: %v\n", err)
                        return
                }

                // Log the restore operation in audit log
                if session != nil {
                        description := fmt.Sprintf("Restored database from backup at %s", backupPath)
                        db.AddAuditLog(
                                session.Username,
                                db.ActionBackup,
                                "database",
                                "restore",
                                description,
                                "",
                                "",
                                "",
                                "",
                        )
                }

                // Clean up temporary decrypted file
                if isEncrypted {
                        if err := os.Remove(backupPath); err != nil {
                                fmt.Printf("Warning: Failed to clean up temporary decrypted file: %v\n", err)
                        }
                }

                fmt.Println("Database restored successfully")
        },
}

// decryptBackup decrypts a backup file
func decryptBackup(encryptedPath, decryptedPath, password string) error {
        // Read the encrypted backup
        data, err := os.ReadFile(encryptedPath)
        if err != nil {
                return fmt.Errorf("failed to read encrypted backup: %w", err)
        }

        // Set custom decryption key if provided
        if password != "" {
                // For simplicity, we're not implementing custom password handling
                // In a real implementation, you would use the password to derive a key
                fmt.Println("Custom password decryption not implemented, using default key")
        }

        // Decrypt the data
        decrypted, err := security.Decrypt(string(data))
        if err != nil {
                return fmt.Errorf("failed to decrypt backup: %w", err)
        }

        // Write to output file
        err = os.WriteFile(decryptedPath, []byte(decrypted), 0644)
        if err != nil {
                return fmt.Errorf("failed to write decrypted backup: %w", err)
        }

        return nil
}

// listBackupsCmd lists available backups
var listBackupsCmd = &cobra.Command{
        Use:   "list-backups [path]",
        Short: "List available backups",
        Long:  `List all available backups with details about creation time, size, and type.`,
        Args:  cobra.MaximumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check if user has permission
                session := auth.GetCurrentUser()
                if session == nil || auth.RequirePermission("setting:read") != nil {
                        fmt.Println("Error: You don't have permission to view backups")
                        return
                }

                // Determine backup path
                backupPath := ""
                if len(args) > 0 {
                        backupPath = args[0]
                } else {
                        // Get from settings if not specified
                        settings, err := db.GetSettings()
                        if err == nil && settings.Backup.BackupPath != "" {
                                backupPath = settings.Backup.BackupPath
                        } else {
                                backupPath = "./backups" // Default fallback
                        }
                }

                // Check if directory exists
                if _, err := os.Stat(backupPath); os.IsNotExist(err) {
                        fmt.Printf("Backup directory %s does not exist\n", backupPath)
                        return
                }

                // Get list of backup files
                files, err := os.ReadDir(backupPath)
                if err != nil {
                        fmt.Printf("Error reading backup directory: %v\n", err)
                        return
                }

                if len(files) == 0 {
                        fmt.Println("No backups found")
                        return
                }

                // Display backup list
                fmt.Printf("Backups in %s:\n\n", backupPath)
                fmt.Printf("%-30s %-20s %-10s %s\n", "FILENAME", "CREATED", "SIZE", "TYPE")
                fmt.Println(strings.Repeat("-", 80))

                for _, file := range files {
                        if file.IsDir() {
                                continue
                        }

                        filename := file.Name()
                        if !strings.HasPrefix(filename, "pos_backup_") && 
                           !strings.HasSuffix(filename, ".db") && 
                           !strings.HasSuffix(filename, ".enc") {
                                continue
                        }

                        // Get file info
                        fileInfo, err := os.Stat(filepath.Join(backupPath, filename))
                        if err != nil {
                                fmt.Printf("Error getting file info: %v\n", err)
                                continue
                        }

                        // Determine backup type
                        backupType := "Standard"
                        if strings.HasSuffix(filename, ".enc") {
                                backupType = "Encrypted"
                        }

                        // Format size
                        var sizeStr string
                        if fileInfo.Size() < 1024 {
                                sizeStr = fmt.Sprintf("%d B", fileInfo.Size())
                        } else if fileInfo.Size() < 1024*1024 {
                                sizeStr = fmt.Sprintf("%.1f KB", float64(fileInfo.Size())/1024)
                        } else {
                                sizeStr = fmt.Sprintf("%.1f MB", float64(fileInfo.Size())/(1024*1024))
                        }

                        // Parse creation time from filename if possible
                        created := fileInfo.ModTime().Format("2006-01-02 15:04:05")
                        if parts := strings.Split(filename, "_"); len(parts) >= 3 {
                                if timeStr := strings.Split(parts[2], ".")[0]; len(timeStr) == 8 {
                                        if t, err := time.Parse("20060102", timeStr); err == nil {
                                                created = t.Format("2006-01-02")
                                        }
                                }
                        }

                        fmt.Printf("%-30s %-20s %-10s %s\n", filename, created, sizeStr, backupType)
                }
        },
}

func init() {
        // Enhanced backup command
        rootCmd.AddCommand(enhancedBackupCmd)
        enhancedBackupCmd.Flags().BoolVar(&encryptBackup, "encrypt", false, "Encrypt the backup")
        enhancedBackupCmd.Flags().StringVar(&backupPassword, "password", "", "Password for encryption (if not provided, uses system key)")
        enhancedBackupCmd.Flags().BoolVar(&backupRotate, "rotate", true, "Rotate backups (delete old ones)")
        enhancedBackupCmd.Flags().BoolVar(&backupCompress, "compress", false, "Compress the backup")
        enhancedBackupCmd.Flags().BoolVar(&backupVerify, "verify", true, "Verify the backup after creation")
        enhancedBackupCmd.Flags().StringVar(&backupName, "name", "", "Custom backup filename")
        enhancedBackupCmd.Flags().IntVar(&backupRetention, "keep", 0, "Number of backups to keep (0 = use settings)")
        enhancedBackupCmd.Flags().BoolVar(&scheduleBackup, "schedule", false, "Set up scheduled backups")
        enhancedBackupCmd.Flags().IntVar(&scheduleInterval, "interval", 24, "Hours between backups")

        // Restore command
        rootCmd.AddCommand(restoreBackupCmd)
        restoreBackupCmd.Flags().StringVar(&backupPassword, "password", "", "Password for decryption (if not provided, uses system key)")

        // List backups command
        rootCmd.AddCommand(listBackupsCmd)
}