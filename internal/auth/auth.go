package auth

import (
        "encoding/json"
        "errors"
        "fmt"
        "os"
        "path/filepath"
        "time"

        "golang.org/x/crypto/bcrypt"
        "termpos/internal/models"
)

var (
        ErrInvalidCredentials = errors.New("invalid username or password")
        ErrUserInactive       = errors.New("user account is inactive")
        ErrInsufficientPerms  = errors.New("insufficient permissions for this operation")
)

// Permission constants for workflow-related operations
const (
        PermissionConfigureWorkflows = "setting:workflow:configure"
        PermissionRunBackups         = "setting:backup:run"
        PermissionGenerateReports    = "report:generate"
)

// Session represents an authenticated user session
type Session struct {
        UserID       int
        Username     string
        Role         models.Role
        CreatedAt    time.Time
        LastActivity time.Time
}

// CurrentSession stores the active user session
var CurrentSession *Session

// sessionFilePath returns the path to the session file
func sessionFilePath() string {
        homeDir, _ := os.UserHomeDir()
        return filepath.Join(homeDir, ".termpos_session.json")
}

// SaveSession saves the current session to disk
func SaveSession() error {
        if CurrentSession == nil {
                // If no session, remove session file if it exists
                if _, err := os.Stat(sessionFilePath()); err == nil {
                        return os.Remove(sessionFilePath())
                }
                return nil
        }

        // Marshal session to JSON
        data, err := json.Marshal(CurrentSession)
        if err != nil {
                return err
        }

        // Write JSON to file
        return os.WriteFile(sessionFilePath(), data, 0600)
}

// LoadSession loads a saved session from disk
func LoadSession() error {
        // Check if session file exists
        if _, err := os.Stat(sessionFilePath()); os.IsNotExist(err) {
                CurrentSession = nil
                return nil
        }

        // Read file
        data, err := os.ReadFile(sessionFilePath())
        if err != nil {
                return err
        }

        // Unmarshal JSON
        session := &Session{}
        if err := json.Unmarshal(data, session); err != nil {
                return err
        }

        // Set current session
        CurrentSession = session
        return nil
}

// HashPassword creates a password hash using bcrypt
func HashPassword(password string) (string, error) {
        bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        return string(bytes), err
}

// CheckPasswordHash compares a password against a hash
func CheckPasswordHash(password, hash string) bool {
        err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
        return err == nil
}

// Login authenticates a user and creates a session
func Login(username, password string, getUser func(string) (models.User, error), updateLastLogin func(int) error) (*Session, error) {
        // Get user by username
        user, err := getUser(username)
        if err != nil {
                return nil, ErrInvalidCredentials
        }

        // Check if user is active
        if !user.Active {
                return nil, ErrUserInactive
        }

        // Verify password
        if !CheckPasswordHash(password, user.PasswordHash) {
                return nil, ErrInvalidCredentials
        }

        // Update last login time
        if err := updateLastLogin(user.ID); err != nil {
                return nil, err
        }

        // Create and store session
        session := &Session{
                UserID:       user.ID,
                Username:     user.Username,
                Role:         user.Role,
                CreatedAt:    time.Now(),
                LastActivity: time.Now(),
        }

        CurrentSession = session
        
        // Save session to disk
        if err := SaveSession(); err != nil {
                // Just log the error, don't fail the login
                // In a production app, use a proper logger
                fmt.Printf("Warning: Failed to save session: %v\n", err)
        }
        
        return session, nil
}

// Logout clears the current session
func Logout() {
        CurrentSession = nil
        
        // Remove session file
        if err := SaveSession(); err != nil {
                // Just log the error
                // In a production app, use a proper logger
                fmt.Printf("Warning: Failed to clear session: %v\n", err)
        }
}

// IsAuthenticated checks if there is an active session
func IsAuthenticated() bool {
        return CurrentSession != nil
}

// GetCurrentUser returns the current session
func GetCurrentUser() *Session {
        return CurrentSession
}

// RequirePermission checks if the current session has the required permission
func RequirePermission(permission string) error {
        if !IsAuthenticated() {
                return errors.New("authentication required")
        }

        user := models.User{
                ID:       CurrentSession.UserID,
                Username: CurrentSession.Username,
                Role:     CurrentSession.Role,
        }

        if !HasPermission(&user, permission) {
                return ErrInsufficientPerms
        }

        return nil
}

// HasPermission checks if a user has a specific permission
func HasPermission(user *models.User, permission string) bool {
        if user == nil {
                return false
        }

        // For simplicity, role-based checking
        // Admin has all permissions
        if user.Role == "admin" {
                return true
        }

        // Manager permissions
        if user.Role == "manager" {
                switch permission {
                case "setting:read", "setting:backup", "setting:export",
                        "product:read", "product:create", "product:update",
                        "sale:create", "user:read", "role:read",
                        // New workflow permissions
                        PermissionConfigureWorkflows, PermissionRunBackups, 
                        PermissionGenerateReports:
                        return true
                default:
                        return false
                }
        }

        // Cashier permissions
        if user.Role == "cashier" {
                switch permission {
                case "product:read", "sale:create":
                        return true
                default:
                        return false
                }
        }

        return false
}