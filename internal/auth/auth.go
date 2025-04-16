package auth

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"termpos/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInsufficientPerms  = errors.New("insufficient permissions for this operation")
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
	return session, nil
}

// Logout clears the current session
func Logout() {
	CurrentSession = nil
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

	if !user.HasPermission(permission) {
		return ErrInsufficientPerms
	}

	return nil
}