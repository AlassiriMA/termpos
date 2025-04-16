package main

import (
        "fmt"
        "os"
        "path/filepath"

        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/models"
)

// workflowCmd represents the workflow command for managing workflows and jobs
var workflowCmd = &cobra.Command{
        Use:   "workflow",
        Short: "Manage automated POS workflows",
        Long:  `Configure and run automated workflows and scheduled jobs like backups and reports.`,
        Run: func(cmd *cobra.Command, args []string) {
                cmd.Help()
        },
}

// backupWorkflowCmd configures and runs backup workflows
var backupWorkflowCmd = &cobra.Command{
        Use:   "backup [configure|run]",
        Short: "Configure or run database backup workflows",
        Long:  `Configure automatic database backups or run an immediate backup.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                action := args[0]

                // Check permissions - only admin and manager can configure workflows
                if action == "configure" {
                        // Get current user
                        session := auth.GetCurrentUser()
                        if session == nil {
                                fmt.Println("Error: You must be logged in to configure workflows")
                                return
                        }

                        // Create user object for permission check
                        user := &models.User{
                                ID:       session.UserID,
                                Username: session.Username,
                                Role:     session.Role,
                        }

                        if !auth.HasPermission(user, auth.PermissionConfigureWorkflows) {
                                fmt.Println("Error: You don't have permission to configure workflows")
                                return
                        }
                }

                // Running a backup requires at least manager permissions
                if action == "run" {
                        // Get current user
                        session := auth.GetCurrentUser()
                        if session == nil {
                                fmt.Println("Error: You must be logged in to run backup operations")
                                return
                        }

                        // Create user object for permission check
                        user := &models.User{
                                ID:       session.UserID,
                                Username: session.Username,
                                Role:     session.Role,
                        }

                        if !auth.HasPermission(user, auth.PermissionRunBackups) {
                                fmt.Println("Error: You don't have permission to run backup operations")
                                return
                        }
                }

                switch action {
                case "configure":
                        // Configure backup settings
                        fmt.Println("Configuring backup workflow...")
                        intervalFlag, _ := cmd.Flags().GetInt("interval")
                        pathFlag, _ := cmd.Flags().GetString("path")
                        enabledFlag, _ := cmd.Flags().GetBool("enabled")
                        keepCountFlag, _ := cmd.Flags().GetInt("keep")

                        // Call configure backup function
                        if err := configureBackupWorkflow(intervalFlag, pathFlag, enabledFlag, keepCountFlag); err != nil {
                                fmt.Printf("Error configuring backup workflow: %v\n", err)
                                return
                        }
                        fmt.Println("Backup workflow configured successfully")

                case "run":
                        // Run an immediate backup
                        fmt.Println("Running backup workflow...")
                        if err := runBackupWorkflow(); err != nil {
                                fmt.Printf("Error running backup: %v\n", err)
                                return
                        }
                        fmt.Println("Backup completed successfully")

                default:
                        fmt.Printf("Unknown action: %s\n", action)
                        cmd.Help()
                }
        },
}

// reportsWorkflowCmd configures and runs report workflows
var reportsWorkflowCmd = &cobra.Command{
        Use:   "reports [configure|run]",
        Short: "Configure or run automated report workflows",
        Long:  `Configure automatic report generation and export or run an immediate report.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                action := args[0]

                // Check permissions - only admin and manager can configure workflows
                if action == "configure" {
                        // Get current user
                        session := auth.GetCurrentUser()
                        if session == nil {
                                fmt.Println("Error: You must be logged in to configure workflows")
                                return
                        }

                        // Create user object for permission check
                        user := &models.User{
                                ID:       session.UserID,
                                Username: session.Username,
                                Role:     session.Role,
                        }

                        if !auth.HasPermission(user, auth.PermissionConfigureWorkflows) {
                                fmt.Println("Error: You don't have permission to configure workflows")
                                return
                        }
                }

                // Running reports requires at least manager permissions
                if action == "run" {
                        // Get current user
                        session := auth.GetCurrentUser()
                        if session == nil {
                                fmt.Println("Error: You must be logged in to run report workflows")
                                return
                        }

                        // Create user object for permission check
                        user := &models.User{
                                ID:       session.UserID,
                                Username: session.Username,
                                Role:     session.Role,
                        }

                        if !auth.HasPermission(user, auth.PermissionGenerateReports) {
                                fmt.Println("Error: You don't have permission to run report workflows")
                                return
                        }
                }

                switch action {
                case "configure":
                        // Configure reports workflow
                        fmt.Println("Configuring reports workflow...")
                        // TODO: Implement report configuration

                case "run":
                        // Run an immediate report
                        fmt.Println("Running reports workflow...")
                        // TODO: Implement report generation workflow

                default:
                        fmt.Printf("Unknown action: %s\n", action)
                        cmd.Help()
                }
        },
}

// initialize sets up the workflow commands
func init() {
        rootCmd.AddCommand(workflowCmd)
        workflowCmd.AddCommand(backupWorkflowCmd)
        workflowCmd.AddCommand(reportsWorkflowCmd)

        // Add flags for backup configuration
        backupWorkflowCmd.Flags().Int("interval", 24, "Backup interval in hours")
        backupWorkflowCmd.Flags().String("path", "./backups", "Backup directory path")
        backupWorkflowCmd.Flags().Bool("enabled", true, "Enable automatic backups")
        backupWorkflowCmd.Flags().Int("keep", 7, "Number of backups to keep")
}

// configureBackupWorkflow sets up the backup workflow
func configureBackupWorkflow(interval int, path string, enabled bool, keepCount int) error {
        // Ensure the backup directory exists
        if enabled {
                if err := os.MkdirAll(path, 0755); err != nil {
                        return fmt.Errorf("failed to create backup directory: %w", err)
                }
        }

        // Get current settings
        settings, err := db.GetSettings()
        if err != nil {
                return fmt.Errorf("failed to get settings: %w", err)
        }

        // Update backup settings
        settings.Backup.AutoBackupEnabled = enabled
        settings.Backup.BackupInterval = interval
        settings.Backup.BackupPath = path
        settings.Backup.KeepBackupCount = keepCount

        // Get current user (for audit trails)
        session := auth.GetCurrentUser()
        username := "system"
        if session != nil {
                username = session.Username
        }

        // Save updated settings
        if err := db.SaveSettings(settings, username); err != nil {
                return fmt.Errorf("failed to save backup settings: %w", err)
        }

        fmt.Printf("Backup workflow configured with:\n")
        fmt.Printf("  Interval: %d hours\n", interval)
        fmt.Printf("  Path: %s\n", path)
        fmt.Printf("  Enabled: %t\n", enabled)
        fmt.Printf("  Keep count: %d\n", keepCount)

        return nil
}

// runBackupWorkflow executes a backup job
func runBackupWorkflow() error {
        // Get settings to get backup configuration
        settings, err := db.GetSettings()
        if err != nil {
                // Use default values if settings can't be retrieved
                fmt.Printf("Warning: Could not retrieve settings, using defaults: %v\n", err)
                err = db.BackupDatabase("./backups")
                if err != nil {
                        return fmt.Errorf("failed to backup database: %w", err)
                }
                
                // Clean up old backups (keep default 7)
                if err := db.CleanupBackups("./backups", 7); err != nil {
                        return fmt.Errorf("failed to cleanup old backups: %w", err)
                }
                
                return nil
        }
        
        // Use configured values from settings
        backupPath := settings.Backup.BackupPath
        if backupPath == "" {
                backupPath = "./backups" // Fallback to default
        }

        // Perform the backup
        if err := db.BackupDatabase(backupPath); err != nil {
                return fmt.Errorf("failed to backup database: %w", err)
        }
        
        // Clean up old backups
        keepCount := settings.Backup.KeepBackupCount
        if keepCount <= 0 {
                keepCount = 7 // Default to keeping 7 backups
        }
        
        if err := db.CleanupBackups(backupPath, keepCount); err != nil {
                return fmt.Errorf("failed to cleanup old backups: %w", err)
        }

        return nil
}

// cleanupOldBackups removes old backup files keeping only the most recent ones
func cleanupOldBackups(path string, keepCount int) error {
        // List all backup files
        files, err := filepath.Glob(filepath.Join(path, "pos_backup_*.db"))
        if err != nil {
                return err
        }

        // If we have fewer files than we want to keep, return
        if len(files) <= keepCount {
                return nil
        }

        // Sort files by modification time (newest first)
        // This is a simplification - in a real implementation, we'd sort by timestamp
        fmt.Printf("Found %d backup files, keeping %d\n", len(files), keepCount)
        
        // Remove oldest files (those that would be deleted)
        filesToDelete := files[keepCount:]
        for _, file := range filesToDelete {
                fmt.Printf("Removing old backup: %s\n", filepath.Base(file))
                if err := os.Remove(file); err != nil {
                        return fmt.Errorf("failed to remove old backup %s: %w", file, err)
                }
        }

        return nil
}