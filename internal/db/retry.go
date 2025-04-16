package db

import (
        "fmt"
        "time"
)

// RetryConfig defines the configuration for database retry operations
type RetryConfig struct {
        MaxRetries     int
        InitialDelay   time.Duration
        MaxDelay       time.Duration
        BackoffFactor  float64
        RetryableError func(error) bool
}

// DefaultRetryConfig provides standard retry settings
var DefaultRetryConfig = RetryConfig{
        MaxRetries:    5,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      2 * time.Second,
        BackoffFactor: 2.0,
        RetryableError: func(err error) bool {
                return isDBLockedError(err)
        },
}

// WithRetry executes a database operation with retry logic
func WithRetry(operation func() error) error {
        return WithCustomRetry(DefaultRetryConfig, operation)
}

// WithCustomRetry executes a database operation with custom retry configuration
func WithCustomRetry(config RetryConfig, operation func() error) error {
        var err error
        delay := config.InitialDelay

        for attempt := 1; attempt <= config.MaxRetries; attempt++ {
                // Try the operation
                err = operation()

                // If successful or error is not retryable, return
                if err == nil || !config.RetryableError(err) {
                        return err
                }

                // If this was the last attempt, return the error
                if attempt == config.MaxRetries {
                        return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries, err)
                }

                // Wait before retrying with exponential backoff
                if delay > config.MaxDelay {
                        delay = config.MaxDelay
                }
                time.Sleep(delay)
                delay = time.Duration(float64(delay) * config.BackoffFactor)
        }

        return err // Should never reach here
}

// isDBLockedError checks if an error is a database locked error
func isDBLockedError(err error) bool {
        if err == nil {
                return false
        }
        errorMsg := err.Error()
        return errorMsg == "database is locked" || 
               errorMsg == "database table is locked" ||
               errorMsg == "database busy" ||
               errorMsg == "locked" ||
               errorMsg == "busy" ||
               errorMsg == "resource temporarily unavailable" ||
               // Common SQLite error message fragments
               (len(errorMsg) > 5 && 
                (errorMsg[:5] == "lock " ||
                 errorMsg[:5] == "busy:" ||
                 (len(errorMsg) > 4 && 
                  (errorMsg[:4] == "SQLITE_BUSY" ||
                   errorMsg[:4] == "SQLITE_LOCKED"))))
}