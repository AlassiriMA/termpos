package db

import (
        "database/sql"
        "fmt"
        "regexp"
        "strings"
        "time"

        "termpos/internal/security"
)

// SensitiveDataEntry represents a sensitive data item in the database
type SensitiveDataEntry struct {
        ID           int64
        ResourceType string
        ResourceID   int64
        FieldName    string
        Value        string
        CreatedAt    string
        UpdatedAt    string
}

// StoreSensitiveData encrypts and stores sensitive data for a resource
func StoreSensitiveData(resourceType string, resourceID int64, fieldName string, value string) error {
        db, err := GetDB()
        if err != nil {
                return fmt.Errorf("failed to get database connection: %w", err)
        }

        // Check if field already exists
        var count int
        err = db.QueryRow(
                "SELECT COUNT(*) FROM sensitive_data WHERE resource_type = ? AND resource_id = ? AND field_name = ?",
                resourceType, resourceID, fieldName,
        ).Scan(&count)
        if err != nil {
                return fmt.Errorf("failed to check for existing sensitive data: %w", err)
        }

        // Encrypt the value
        encrypted, err := security.Encrypt(value)
        if err != nil {
                return fmt.Errorf("failed to encrypt sensitive data: %w", err)
        }

        // Update or insert
        if count > 0 {
                _, err = db.Exec(
                        "UPDATE sensitive_data SET value = ?, updated_at = ? WHERE resource_type = ? AND resource_id = ? AND field_name = ?",
                        encrypted, time.Now().Format(time.RFC3339), resourceType, resourceID, fieldName,
                )
                if err != nil {
                        return fmt.Errorf("failed to update sensitive data: %w", err)
                }
        } else {
                _, err = db.Exec(
                        "INSERT INTO sensitive_data (resource_type, resource_id, field_name, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
                        resourceType, resourceID, fieldName, encrypted, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339),
                )
                if err != nil {
                        return fmt.Errorf("failed to store sensitive data: %w", err)
                }
        }

        return nil
}

// GetSensitiveData retrieves and decrypts sensitive data for a resource
func GetSensitiveData(resourceType string, resourceID int64, fieldName string) (string, error) {
        db, err := GetDB()
        if err != nil {
                return "", fmt.Errorf("failed to get database connection: %w", err)
        }

        var encrypted string
        err = db.QueryRow(
                "SELECT value FROM sensitive_data WHERE resource_type = ? AND resource_id = ? AND field_name = ?",
                resourceType, resourceID, fieldName,
        ).Scan(&encrypted)
        if err != nil {
                if err == sql.ErrNoRows {
                        return "", fmt.Errorf("sensitive data not found")
                }
                return "", fmt.Errorf("failed to retrieve sensitive data: %w", err)
        }

        // Decrypt the value
        decrypted, err := security.Decrypt(encrypted)
        if err != nil {
                return "", fmt.Errorf("failed to decrypt sensitive data: %w", err)
        }

        return decrypted, nil
}

// DeleteSensitiveData deletes sensitive data for a resource
func DeleteSensitiveData(resourceType string, resourceID int64, fieldName string) error {
        db, err := GetDB()
        if err != nil {
                return fmt.Errorf("failed to get database connection: %w", err)
        }

        result, err := db.Exec(
                "DELETE FROM sensitive_data WHERE resource_type = ? AND resource_id = ? AND field_name = ?",
                resourceType, resourceID, fieldName,
        )
        if err != nil {
                return fmt.Errorf("failed to delete sensitive data: %w", err)
        }

        rows, err := result.RowsAffected()
        if err != nil {
                return fmt.Errorf("failed to get rows affected: %w", err)
        }

        if rows == 0 {
                return fmt.Errorf("sensitive data not found")
        }

        return nil
}

// DeleteAllSensitiveDataForResource deletes all sensitive data for a resource
func DeleteAllSensitiveDataForResource(resourceType string, resourceID int64) error {
        db, err := GetDB()
        if err != nil {
                return fmt.Errorf("failed to get database connection: %w", err)
        }

        _, err = db.Exec(
                "DELETE FROM sensitive_data WHERE resource_type = ? AND resource_id = ?",
                resourceType, resourceID,
        )
        if err != nil {
                return fmt.Errorf("failed to delete sensitive data: %w", err)
        }

        return nil
}

// GetSensitiveDataFields gets all field names for sensitive data for a resource
func GetSensitiveDataFields(resourceType string, resourceID int64) ([]string, error) {
        db, err := GetDB()
        if err != nil {
                return nil, fmt.Errorf("failed to get database connection: %w", err)
        }

        rows, err := db.Query(
                "SELECT field_name FROM sensitive_data WHERE resource_type = ? AND resource_id = ?",
                resourceType, resourceID,
        )
        if err != nil {
                return nil, fmt.Errorf("failed to retrieve sensitive data fields: %w", err)
        }
        defer rows.Close()

        var fields []string
        for rows.Next() {
                var field string
                if err := rows.Scan(&field); err != nil {
                        return nil, fmt.Errorf("failed to scan field name: %w", err)
                }
                fields = append(fields, field)
        }

        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating sensitive data fields: %w", err)
        }

        return fields, nil
}

// IsSensitiveField checks if a field name is sensitive based on patterns
func IsSensitiveField(fieldName string) bool {
        // Convert to lowercase for case-insensitive comparison
        name := strings.ToLower(fieldName)

        // Common sensitive field name patterns
        sensitivePatterns := []string{
                "password", "passwd", "pass", "pwd",
                "secret", "key", "token", "auth",
                "credential", "api_key", "apikey",
                "pin", "code", "cvv", "credit",
                "card", "ssn", "social", "tax",
                "license", "private", "confidential",
                "personal", "sensitive", "secure",
        }

        // Check for exact matches or substrings
        for _, pattern := range sensitivePatterns {
                if strings.Contains(name, pattern) {
                        return true
                }
        }

        // Check for email pattern
        emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
        if emailPattern.MatchString(fieldName) {
                return true
        }

        // Check for credit card pattern
        ccPattern := regexp.MustCompile(`^(?:\d{4}[- ]?){3}\d{4}$`)
        if ccPattern.MatchString(fieldName) {
                return true
        }

        // Check for phone number pattern
        phonePattern := regexp.MustCompile(`^\+?(?:\d{1,3})?[-. (]?\d{3}[-. )]?\d{3}[-. ]?\d{4}$`)
        if phonePattern.MatchString(fieldName) {
                return true
        }

        return false
}

// RedactSensitiveData checks if data is sensitive and redacts it if necessary
func RedactSensitiveData(data string) string {
        if IsSensitiveField(data) {
                return "<redacted>"
        }
        return data
}