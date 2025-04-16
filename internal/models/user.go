package models

import (
	"time"
)

// Role represents a user's access level
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleCashier Role = "cashier"
)

// User represents a system user
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose the password hash in JSON
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	LastLoginAt  time.Time `json:"last_login_at,omitempty"`
	Active       bool      `json:"active"`
}

// HasPermission checks if the user has permission to perform an action
func (u *User) HasPermission(permission string) bool {
	switch permission {
	case "user:manage":
		return u.Role == RoleAdmin

	case "report:generate":
		return u.Role == RoleAdmin || u.Role == RoleManager

	case "product:manage":
		return u.Role == RoleAdmin || u.Role == RoleManager

	case "sales:view":
		return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleCashier

	case "inventory:view":
		return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleCashier

	case "sales:create":
		return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleCashier

	default:
		return false
	}
}

// IsAdmin checks if the user is an administrator
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsManager checks if the user is a manager
func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

// IsCashier checks if the user is a cashier
func (u *User) IsCashier() bool {
	return u.Role == RoleCashier
}