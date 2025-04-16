package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"termpos/internal/auth"
	"termpos/internal/db"
)

var (
	// Audit log command flags
	auditStartDate string
	auditEndDate   string
	auditUsername  string
	auditAction    string
	auditResource  string
	auditLimit     int
	auditOffset    int
	auditExport    string
	auditPurge     int
	auditStats     bool
)

// auditCmd represents the audit command for managing and viewing audit logs
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View and manage audit logs",
	Long:  `View, export, and manage audit logs for compliance and tracking changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// auditListCmd shows audit logs with filtering options
var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit logs",
	Long:  `Display audit logs with various filtering options.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user has permission
		session := auth.GetCurrentUser()
		if session == nil || auth.RequirePermission("audit:view") != nil {
			fmt.Println("Error: You don't have permission to view audit logs")
			return
		}

		// Get audit logs with filters
		logs, err := db.GetAuditLogs(
			auditUsername,
			db.AuditAction(auditAction),
			auditResource,
			auditStartDate,
			auditEndDate,
			auditLimit,
			auditOffset,
		)
		if err != nil {
			fmt.Printf("Error getting audit logs: %v\n", err)
			return
		}

		if len(logs) == 0 {
			fmt.Println("No audit logs found matching the criteria")
			return
		}

		// Display audit logs in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Timestamp", "User", "Action", "Resource", "Resource ID", "Description"})
		table.SetBorder(false)
		table.SetColumnSeparator(" | ")

		for _, log := range logs {
			// Format timestamp
			timestamp := log.Timestamp.Format("2006-01-02 15:04:05")

			// Truncate description if too long
			description := log.Description
			if len(description) > 50 {
				description = description[:47] + "..."
			}

			table.Append([]string{
				fmt.Sprintf("%d", log.ID),
				timestamp,
				log.Username,
				string(log.Action),
				log.ResourceType,
				log.ResourceID,
				description,
			})
		}

		table.Render()

		// Show pagination info
		if auditLimit > 0 {
			fmt.Printf("Showing %d records (offset: %d, limit: %d)\n", len(logs), auditOffset, auditLimit)
		} else {
			fmt.Printf("Showing %d records\n", len(logs))
		}
	},
}

// auditViewCmd shows details of a specific audit log
var auditViewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "View details of a specific audit log",
	Long:  `Show all details of a specific audit log entry including previous and new values.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user has permission
		session := auth.GetCurrentUser()
		if session == nil || auth.RequirePermission("audit:view") != nil {
			fmt.Println("Error: You don't have permission to view audit logs")
			return
		}

		// Parse log ID
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Printf("Error: Invalid audit log ID: %v\n", err)
			return
		}

		// Get all logs and find the one with matching ID
		logs, err := db.GetAuditLogs("", "", "", "", "", 0, 0)
		if err != nil {
			fmt.Printf("Error getting audit logs: %v\n", err)
			return
		}

		var targetLog *db.AuditLog
		for i, log := range logs {
			if log.ID == id {
				targetLog = &logs[i]
				break
			}
		}

		if targetLog == nil {
			fmt.Printf("Error: Audit log with ID %d not found\n", id)
			return
		}

		// Display detailed audit log information
		fmt.Println("=== Audit Log Details ===")
		fmt.Printf("ID: %d\n", targetLog.ID)
		fmt.Printf("Timestamp: %s\n", targetLog.Timestamp.Format(time.RFC3339))
		fmt.Printf("User: %s\n", targetLog.Username)
		fmt.Printf("Action: %s\n", targetLog.Action)
		fmt.Printf("Resource Type: %s\n", targetLog.ResourceType)
		fmt.Printf("Resource ID: %s\n", targetLog.ResourceID)
		fmt.Printf("Description: %s\n", targetLog.Description)
		
		if targetLog.IPAddress != "" {
			fmt.Printf("IP Address: %s\n", targetLog.IPAddress)
		}
		
		if targetLog.AdditionalInfo != "" {
			fmt.Printf("Additional Info: %s\n", targetLog.AdditionalInfo)
		}
		
		fmt.Println()
		
		if targetLog.PreviousValue != "" {
			fmt.Println("Previous Value:")
			fmt.Println(targetLog.PreviousValue)
			fmt.Println()
		}
		
		if targetLog.NewValue != "" {
			fmt.Println("New Value:")
			fmt.Println(targetLog.NewValue)
			fmt.Println()
		}
	},
}

// auditExportCmd exports audit logs
var auditExportCmd = &cobra.Command{
	Use:   "export [filepath]",
	Short: "Export audit logs to a JSON file",
	Long:  `Export audit logs to a JSON file with optional filtering by date range.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user has permission
		session := auth.GetCurrentUser()
		if session == nil || auth.RequirePermission("audit:export") != nil {
			fmt.Println("Error: You don't have permission to export audit logs")
			return
		}

		filepath := args[0]
		err := db.ExportAuditLogs(filepath, auditStartDate, auditEndDate)
		if err != nil {
			fmt.Printf("Error exporting audit logs: %v\n", err)
			return
		}

		fmt.Printf("Audit logs exported to %s\n", filepath)
	},
}

// auditPurgeCmd removes old audit logs
var auditPurgeCmd = &cobra.Command{
	Use:   "purge [days]",
	Short: "Remove audit logs older than specified days",
	Long:  `Delete audit logs that are older than the specified number of days.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user has permission
		session := auth.GetCurrentUser()
		if session == nil || auth.RequirePermission("audit:purge") != nil {
			fmt.Println("Error: You don't have permission to purge audit logs")
			return
		}

		// Parse days
		days, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error: Invalid number of days: %v\n", err)
			return
		}

		if days <= 0 {
			fmt.Println("Error: Number of days must be positive")
			return
		}

		// Confirm action with user
		fmt.Printf("WARNING: This will permanently delete audit logs older than %d days.\n", days)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Operation cancelled")
			return
		}

		// Purge old logs
		count, err := db.PurgeOldAuditLogs(days)
		if err != nil {
			fmt.Printf("Error purging audit logs: %v\n", err)
			return
		}

		fmt.Printf("Successfully purged %d audit logs older than %d days\n", count, days)
	},
}

// auditStatsCmd shows statistics about audit logs
var auditStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show audit log statistics",
	Long:  `Display statistics about the audit logs, such as total count, date ranges, and action types.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user has permission
		session := auth.GetCurrentUser()
		if session == nil || auth.RequirePermission("audit:view") != nil {
			fmt.Println("Error: You don't have permission to view audit logs")
			return
		}

		// Get statistics
		stats, err := db.AuditLogStatistics()
		if err != nil {
			fmt.Printf("Error getting audit log statistics: %v\n", err)
			return
		}

		// Display statistics
		fmt.Println("=== Audit Log Statistics ===")
		fmt.Printf("Total Logs: %d\n", stats["total_count"])
		
		if stats["oldest_log"] != nil {
			fmt.Printf("Oldest Log: %s\n", stats["oldest_log"])
		} else {
			fmt.Println("Oldest Log: No logs found")
		}
		
		if stats["newest_log"] != nil {
			fmt.Printf("Newest Log: %s\n", stats["newest_log"])
		} else {
			fmt.Println("Newest Log: No logs found")
		}
		
		fmt.Println()
		
		fmt.Println("Actions:")
		if actionCounts, ok := stats["action_counts"].(map[string]int); ok {
			for action, count := range actionCounts {
				fmt.Printf("  %s: %d\n", action, count)
			}
		}
		
		fmt.Println()
		
		fmt.Println("Resource Types:")
		if resourceCounts, ok := stats["resource_counts"].(map[string]int); ok {
			for resourceType, count := range resourceCounts {
				fmt.Printf("  %s: %d\n", resourceType, count)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditViewCmd)
	auditCmd.AddCommand(auditExportCmd)
	auditCmd.AddCommand(auditPurgeCmd)
	auditCmd.AddCommand(auditStatsCmd)

	// Add flags to audit list command
	auditListCmd.Flags().StringVar(&auditStartDate, "start", "", "Start date (YYYY-MM-DD)")
	auditListCmd.Flags().StringVar(&auditEndDate, "end", "", "End date (YYYY-MM-DD)")
	auditListCmd.Flags().StringVar(&auditUsername, "user", "", "Filter by username")
	auditListCmd.Flags().StringVar(&auditAction, "action", "", "Filter by action type")
	auditListCmd.Flags().StringVar(&auditResource, "resource", "", "Filter by resource type")
	auditListCmd.Flags().IntVar(&auditLimit, "limit", 50, "Limit number of results")
	auditListCmd.Flags().IntVar(&auditOffset, "offset", 0, "Offset for pagination")

	// Add flags to audit export command
	auditExportCmd.Flags().StringVar(&auditStartDate, "start", "", "Start date (YYYY-MM-DD)")
	auditExportCmd.Flags().StringVar(&auditEndDate, "end", "", "End date (YYYY-MM-DD)")
}