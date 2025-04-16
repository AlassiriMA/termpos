package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"termpos/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("username already exists")
)

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (models.User, error) {
	var user models.User
	query := `
	SELECT id, username, password_hash, role, created_at, last_login_at, active
	FROM users
	WHERE id = ?`

	var lastLoginTime sql.NullTime
	err := DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&lastLoginTime,
		&user.Active,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	if lastLoginTime.Valid {
		user.LastLoginAt = lastLoginTime.Time
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	query := `
	SELECT id, username, password_hash, role, created_at, last_login_at, active
	FROM users
	WHERE username = ?`

	var lastLoginTime sql.NullTime
	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&lastLoginTime,
		&user.Active,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	if lastLoginTime.Valid {
		user.LastLoginAt = lastLoginTime.Time
	}

	return user, nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]models.User, error) {
	query := `
	SELECT id, username, password_hash, role, created_at, last_login_at, active
	FROM users
	ORDER BY username`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastLoginTime sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&lastLoginTime,
			&user.Active,
		)

		if err != nil {
			return nil, err
		}

		if lastLoginTime.Valid {
			user.LastLoginAt = lastLoginTime.Time
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser adds a new user to the database
func CreateUser(user models.User) (int, error) {
	// Check if username already exists
	exists, err := userExists(user.Username)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrUserExists
	}

	query := `
	INSERT INTO users (username, password_hash, role, active)
	VALUES (?, ?, ?, ?)`

	result, err := DB.Exec(query, user.Username, user.PasswordHash, user.Role, user.Active)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateUser updates user information
func UpdateUser(user models.User) error {
	query := `
	UPDATE users
	SET role = ?, active = ?
	WHERE id = ?`

	result, err := DB.Exec(query, user.Role, user.Active, user.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(userID int, passwordHash string) error {
	query := `
	UPDATE users
	SET password_hash = ?
	WHERE id = ?`

	result, err := DB.Exec(query, passwordHash, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates a user's last login time
func UpdateLastLogin(userID int) error {
	query := `
	UPDATE users
	SET last_login_at = ?
	WHERE id = ?`

	result, err := DB.Exec(query, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser removes a user from the system
func DeleteUser(userID int) error {
	query := `
	DELETE FROM users
	WHERE id = ?`

	result, err := DB.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// userExists checks if a username is already taken
func userExists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ?"
	
	err := DB.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	
	return count > 0, nil
}