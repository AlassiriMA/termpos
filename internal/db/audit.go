package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// AuditAction represents the type of action performed in an audit log
type AuditAction string

// Define common audit actions
const (
	ActionCreate     AuditAction = "create"
	ActionUpdate     AuditAction = "update"
	ActionDelete     AuditAction = "delete"
	ActionLogin      AuditAction = "login"
	ActionLogout     AuditAction = "logout"
	ActionExport     AuditAction = "export"
	ActionImport     AuditAction = "import"
	ActionBackup     AuditAction = "backup"
	ActionRestore    AuditAction = "restore"
	ActionSettingsMod AuditAction = "settings_change"
	ActionPermissionMod AuditAction = "permission_change"
	ActionUserMod    AuditAction = "user_change"
	ActionAccess     AuditAction = "access"
	ActionExecute    AuditAction = "execute"
	ActionSale       AuditAction = "sale"
	ActionRefund     AuditAction = "refund"
	ActionInventory  AuditAction = "inventory"
)

// AuditLog represents an entry in the audit log
type AuditLog struct {
	ID            int64       `json:"id"`
	Timestamp     time.Time   `json:"timestamp"`
	Username      string      `json:"username"`
	Action        AuditAction `json:"action"`
	ResourceType  string      `json:"resource_type"`
	ResourceID    string      `json:"resource_id"`
	Description   string      `json:"description"`
	PreviousValue string      `json:"previous_value,omitempty"`
	NewValue      string      `json:"new_value,omitempty"`
	IPAddress     string      `json:"ip_address,omitempty"`
	AdditionalInfo string     `json:"additional_info,omitempty"`
}

// AddAuditLog adds a new audit log entry
func AddAuditLog(username string, action AuditAction, resourceType, resourceID, description, previousValue, newValue, ipAddress, additionalInfo string) error {
	// Get database connection
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Insert audit log
	query := `
		INSERT INTO audit_logs (
			username, action, resource_type, resource_id, description, 
			previous_value, new_value, ip_address, additional_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(
		query,
		username,
		string(action),
		resourceType,
		resourceID,
		description,
		previousValue,
		newValue,
		ipAddress,
		additionalInfo,
	)

	if err != nil {
		return fmt.Errorf("failed to add audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func GetAuditLogs(username string, action AuditAction, resourceType, startDate, endDate string, limit, offset int) ([]AuditLog, error) {
	// Get database connection
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Build query with filters
	query := "SELECT id, timestamp, username, action, resource_type, resource_id, description, previous_value, new_value, ip_address, additional_info FROM audit_logs WHERE 1=1"
	params := []interface{}{}

	if username != "" {
		query += " AND username = ?"
		params = append(params, username)
	}

	if action != "" {
		query += " AND action = ?"
		params = append(params, string(action))
	}

	if resourceType != "" {
		query += " AND resource_type = ?"
		params = append(params, resourceType)
	}

	if startDate != "" {
		query += " AND timestamp >= ?"
		params = append(params, startDate)
	}

	if endDate != "" {
		query += " AND timestamp <= ?"
		// Adjust end date to include the entire day if it's just a date
		if len(endDate) == 10 { // YYYY-MM-DD format
			endDate += " 23:59:59"
		}
		params = append(params, endDate)
	}

	// Add order by and pagination
	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += " LIMIT ?"
		params = append(params, limit)

		if offset > 0 {
			query += " OFFSET ?"
			params = append(params, offset)
		}
	}

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	// Process results
	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var timestamp string

		err := rows.Scan(
			&log.ID,
			&timestamp,
			&log.Username,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.Description,
			&log.PreviousValue,
			&log.NewValue,
			&log.IPAddress,
			&log.AdditionalInfo,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Parse timestamp
		log.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			// Try alternative format
			log.Timestamp, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return logs, nil
}

// PurgeOldAuditLogs removes audit logs older than the specified number of days
func PurgeOldAuditLogs(days int) (int64, error) {
	// Get database connection
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -days)
	cutoffStr := cutoffDate.Format("2006-01-02")

	// Delete old logs
	result, err := db.Exec("DELETE FROM audit_logs WHERE timestamp < ?", cutoffStr)
	if err != nil {
		return 0, fmt.Errorf("failed to purge old audit logs: %w", err)
	}

	// Get number of deleted rows
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// AuditLogStatistics returns statistics about audit logs
func AuditLogStatistics() (map[string]interface{}, error) {
	// Get database connection
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	stats := make(map[string]interface{})

	// Get total count
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_count"] = totalCount

	if totalCount == 0 {
		// No logs, return early
		return stats, nil
	}

	// Get oldest and newest logs
	var oldestTimestamp, newestTimestamp string
	err = db.QueryRow("SELECT timestamp FROM audit_logs ORDER BY timestamp ASC LIMIT 1").Scan(&oldestTimestamp)
	if err == nil {
		stats["oldest_log"] = oldestTimestamp
	}

	err = db.QueryRow("SELECT timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT 1").Scan(&newestTimestamp)
	if err == nil {
		stats["newest_log"] = newestTimestamp
	}

	// Get action counts
	rows, err := db.Query("SELECT action, COUNT(*) FROM audit_logs GROUP BY action")
	if err != nil {
		return nil, fmt.Errorf("failed to get action counts: %w", err)
	}
	defer rows.Close()

	actionCounts := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, fmt.Errorf("failed to scan action count: %w", err)
		}
		actionCounts[action] = count
	}
	stats["action_counts"] = actionCounts

	// Get resource type counts
	rows, err = db.Query("SELECT resource_type, COUNT(*) FROM audit_logs GROUP BY resource_type")
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type counts: %w", err)
	}
	defer rows.Close()

	resourceCounts := make(map[string]int)
	for rows.Next() {
		var resourceType string
		var count int
		if err := rows.Scan(&resourceType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan resource type count: %w", err)
		}
		resourceCounts[resourceType] = count
	}
	stats["resource_counts"] = resourceCounts

	return stats, nil
}

// ExportAuditLogs exports audit logs to a JSON file
func ExportAuditLogs(filepath string, startDate string, endDate string) error {
	// Get logs for the specified date range
	logs, err := GetAuditLogs("", "", "", startDate, endDate, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	// Export to JSON file
	return writeJSONToFile(logs, filepath)
}

// AuditDiff creates a structured diff for audit logs
func AuditDiff(oldValue, newValue interface{}) (string, string, error) {
	// Convert old value to JSON string
	var oldJSON, newJSON string
	var err error

	if oldValue != nil {
		oldBytes, err := json.Marshal(oldValue)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal old value: %w", err)
		}
		oldJSON = string(oldBytes)
	}

	// Convert new value to JSON string
	if newValue != nil {
		newBytes, err := json.Marshal(newValue)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal new value: %w", err)
		}
		newJSON = string(newBytes)
	}

	return oldJSON, newJSON, err
}

// LogUserAction is a helper function to log user actions
func LogUserAction(username string, action AuditAction, resourceType string, resourceID string, description string) error {
	return AddAuditLog(username, action, resourceType, resourceID, description, "", "", "", "")
}

// LogDataChange logs a change to data with previous and new values
func LogDataChange(username string, action AuditAction, resourceType string, resourceID string, description string, oldValue interface{}, newValue interface{}) error {
	prevJSON, newJSON, err := AuditDiff(oldValue, newValue)
	if err != nil {
		return err
	}

	return AddAuditLog(username, action, resourceType, resourceID, description, prevJSON, newJSON, "", "")
}