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
        SELECT id, username, password_hash, role, created_at, last_login_at, active,
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE id = ?`

        var lastLoginTime, hireDate sql.NullTime
        var fullName, email, phone, address, position, department, notes, emergencyContact sql.NullString
        
        err := DB.QueryRow(query, id).Scan(
                &user.ID,
                &user.Username,
                &user.PasswordHash,
                &user.Role,
                &user.CreatedAt,
                &lastLoginTime,
                &user.Active,
                &fullName,
                &email,
                &phone,
                &address,
                &hireDate,
                &position,
                &department,
                &notes,
                &emergencyContact,
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
        
        // Set staff fields from nullable values
        if fullName.Valid {
                user.FullName = fullName.String
        }
        if email.Valid {
                user.Email = email.String
        }
        if phone.Valid {
                user.Phone = phone.String
        }
        if address.Valid {
                user.Address = address.String
        }
        if hireDate.Valid {
                user.HireDate = hireDate.Time
        }
        if position.Valid {
                user.Position = position.String
        }
        if department.Valid {
                user.Department = department.String
        }
        if notes.Valid {
                user.Notes = notes.String
        }
        if emergencyContact.Valid {
                user.EmergencyContact = emergencyContact.String
        }

        return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (models.User, error) {
        var user models.User
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active,
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE username = ?`

        var lastLoginTime, hireDate sql.NullTime
        var fullName, email, phone, address, position, department, notes, emergencyContact sql.NullString
        
        err := DB.QueryRow(query, username).Scan(
                &user.ID,
                &user.Username,
                &user.PasswordHash,
                &user.Role,
                &user.CreatedAt,
                &lastLoginTime,
                &user.Active,
                &fullName,
                &email,
                &phone,
                &address,
                &hireDate,
                &position,
                &department,
                &notes,
                &emergencyContact,
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
        
        // Set staff fields from nullable values
        if fullName.Valid {
                user.FullName = fullName.String
        }
        if email.Valid {
                user.Email = email.String
        }
        if phone.Valid {
                user.Phone = phone.String
        }
        if address.Valid {
                user.Address = address.String
        }
        if hireDate.Valid {
                user.HireDate = hireDate.Time
        }
        if position.Valid {
                user.Position = position.String
        }
        if department.Valid {
                user.Department = department.String
        }
        if notes.Valid {
                user.Notes = notes.String
        }
        if emergencyContact.Valid {
                user.EmergencyContact = emergencyContact.String
        }

        return user, nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]models.User, error) {
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active, 
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
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
                var lastLoginTime, hireDate sql.NullTime
                var fullName, email, phone, address, position, department, notes, emergencyContact sql.NullString

                err := rows.Scan(
                        &user.ID,
                        &user.Username,
                        &user.PasswordHash,
                        &user.Role,
                        &user.CreatedAt,
                        &lastLoginTime,
                        &user.Active,
                        &fullName,
                        &email,
                        &phone,
                        &address,
                        &hireDate,
                        &position,
                        &department,
                        &notes,
                        &emergencyContact,
                )

                if err != nil {
                        return nil, err
                }

                if lastLoginTime.Valid {
                        user.LastLoginAt = lastLoginTime.Time
                }
                
                // Set staff fields from nullable values
                if fullName.Valid {
                        user.FullName = fullName.String
                }
                if email.Valid {
                        user.Email = email.String
                }
                if phone.Valid {
                        user.Phone = phone.String
                }
                if address.Valid {
                        user.Address = address.String
                }
                if hireDate.Valid {
                        user.HireDate = hireDate.Time
                }
                if position.Valid {
                        user.Position = position.String
                }
                if department.Valid {
                        user.Department = department.String
                }
                if notes.Valid {
                        user.Notes = notes.String
                }
                if emergencyContact.Valid {
                        user.EmergencyContact = emergencyContact.String
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
        INSERT INTO users (
                username, password_hash, role, active, 
                full_name, email, phone, address, hire_date, 
                position, department, notes, emergency_contact
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

        // Set default hire date to now if not specified
        if user.HireDate.IsZero() {
                user.HireDate = time.Now()
        }

        result, err := DB.Exec(
                query, 
                user.Username, 
                user.PasswordHash, 
                user.Role, 
                user.Active,
                user.FullName,
                user.Email,
                user.Phone,
                user.Address,
                user.HireDate,
                user.Position,
                user.Department,
                user.Notes,
                user.EmergencyContact,
        )
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
        SET role = ?, active = ?, 
            full_name = ?, email = ?, phone = ?, address = ?,
            hire_date = ?, position = ?, department = ?, 
            notes = ?, emergency_contact = ?
        WHERE id = ?`

        result, err := DB.Exec(
                query, 
                user.Role, 
                user.Active,
                user.FullName,
                user.Email,
                user.Phone,
                user.Address,
                user.HireDate,
                user.Position,
                user.Department,
                user.Notes,
                user.EmergencyContact,
                user.ID,
        )
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

// FindUsersByFullName retrieves users with matching full name
func FindUsersByFullName(name string) ([]models.User, error) {
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active, 
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE full_name LIKE ?
        ORDER BY full_name`

        // Add wildcard characters for LIKE query
        searchPattern := "%" + name + "%"
        rows, err := DB.Query(query, searchPattern)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var users []models.User
        for rows.Next() {
                var user models.User
                var lastLoginTime, hireDate sql.NullTime
                var fullName, email, phone, address, position, department, notes, emergencyContact sql.NullString

                err := rows.Scan(
                        &user.ID,
                        &user.Username,
                        &user.PasswordHash,
                        &user.Role,
                        &user.CreatedAt,
                        &lastLoginTime,
                        &user.Active,
                        &fullName,
                        &email,
                        &phone,
                        &address,
                        &hireDate,
                        &position,
                        &department,
                        &notes,
                        &emergencyContact,
                )

                if err != nil {
                        return nil, err
                }

                if lastLoginTime.Valid {
                        user.LastLoginAt = lastLoginTime.Time
                }
                
                // Set staff fields from nullable values
                if fullName.Valid {
                        user.FullName = fullName.String
                }
                if email.Valid {
                        user.Email = email.String
                }
                if phone.Valid {
                        user.Phone = phone.String
                }
                if address.Valid {
                        user.Address = address.String
                }
                if hireDate.Valid {
                        user.HireDate = hireDate.Time
                }
                if position.Valid {
                        user.Position = position.String
                }
                if department.Valid {
                        user.Department = department.String
                }
                if notes.Valid {
                        user.Notes = notes.String
                }
                if emergencyContact.Valid {
                        user.EmergencyContact = emergencyContact.String
                }

                users = append(users, user)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return users, nil
}

// GetUserByEmail retrieves a user by email address
func GetUserByEmail(email string) (models.User, error) {
        var user models.User
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active,
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE email = ?`

        var lastLoginTime, hireDate sql.NullTime
        var fullName, userEmail, phone, address, position, department, notes, emergencyContact sql.NullString
        
        err := DB.QueryRow(query, email).Scan(
                &user.ID,
                &user.Username,
                &user.PasswordHash,
                &user.Role,
                &user.CreatedAt,
                &lastLoginTime,
                &user.Active,
                &fullName,
                &userEmail,
                &phone,
                &address,
                &hireDate,
                &position,
                &department,
                &notes,
                &emergencyContact,
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
        
        // Set staff fields from nullable values
        if fullName.Valid {
                user.FullName = fullName.String
        }
        if userEmail.Valid {
                user.Email = userEmail.String
        }
        if phone.Valid {
                user.Phone = phone.String
        }
        if address.Valid {
                user.Address = address.String
        }
        if hireDate.Valid {
                user.HireDate = hireDate.Time
        }
        if position.Valid {
                user.Position = position.String
        }
        if department.Valid {
                user.Department = department.String
        }
        if notes.Valid {
                user.Notes = notes.String
        }
        if emergencyContact.Valid {
                user.EmergencyContact = emergencyContact.String
        }

        return user, nil
}

// GetUserByPhone retrieves a user by phone number
func GetUserByPhone(phone string) (models.User, error) {
        var user models.User
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active,
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE phone = ?`

        var lastLoginTime, hireDate sql.NullTime
        var fullName, email, userPhone, address, position, department, notes, emergencyContact sql.NullString
        
        err := DB.QueryRow(query, phone).Scan(
                &user.ID,
                &user.Username,
                &user.PasswordHash,
                &user.Role,
                &user.CreatedAt,
                &lastLoginTime,
                &user.Active,
                &fullName,
                &email,
                &userPhone,
                &address,
                &hireDate,
                &position,
                &department,
                &notes,
                &emergencyContact,
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
        
        // Set staff fields from nullable values
        if fullName.Valid {
                user.FullName = fullName.String
        }
        if email.Valid {
                user.Email = email.String
        }
        if userPhone.Valid {
                user.Phone = userPhone.String
        }
        if address.Valid {
                user.Address = address.String
        }
        if hireDate.Valid {
                user.HireDate = hireDate.Time
        }
        if position.Valid {
                user.Position = position.String
        }
        if department.Valid {
                user.Department = department.String
        }
        if notes.Valid {
                user.Notes = notes.String
        }
        if emergencyContact.Valid {
                user.EmergencyContact = emergencyContact.String
        }

        return user, nil
}

// FindUsersByDepartment retrieves users from a specific department
func FindUsersByDepartment(department string) ([]models.User, error) {
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active, 
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE department = ?
        ORDER BY full_name`

        rows, err := DB.Query(query, department)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var users []models.User
        for rows.Next() {
                var user models.User
                var lastLoginTime, hireDate sql.NullTime
                var fullName, email, phone, address, position, userDepartment, notes, emergencyContact sql.NullString

                err := rows.Scan(
                        &user.ID,
                        &user.Username,
                        &user.PasswordHash,
                        &user.Role,
                        &user.CreatedAt,
                        &lastLoginTime,
                        &user.Active,
                        &fullName,
                        &email,
                        &phone,
                        &address,
                        &hireDate,
                        &position,
                        &userDepartment,
                        &notes,
                        &emergencyContact,
                )

                if err != nil {
                        return nil, err
                }

                if lastLoginTime.Valid {
                        user.LastLoginAt = lastLoginTime.Time
                }
                
                // Set staff fields from nullable values
                if fullName.Valid {
                        user.FullName = fullName.String
                }
                if email.Valid {
                        user.Email = email.String
                }
                if phone.Valid {
                        user.Phone = phone.String
                }
                if address.Valid {
                        user.Address = address.String
                }
                if hireDate.Valid {
                        user.HireDate = hireDate.Time
                }
                if position.Valid {
                        user.Position = position.String
                }
                if userDepartment.Valid {
                        user.Department = userDepartment.String
                }
                if notes.Valid {
                        user.Notes = notes.String
                }
                if emergencyContact.Valid {
                        user.EmergencyContact = emergencyContact.String
                }

                users = append(users, user)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return users, nil
}

// FindUsersByRole retrieves users with a specific role
func FindUsersByRole(role models.Role) ([]models.User, error) {
        query := `
        SELECT id, username, password_hash, role, created_at, last_login_at, active, 
               full_name, email, phone, address, hire_date, position, department, notes, emergency_contact
        FROM users
        WHERE role = ?
        ORDER BY full_name`

        rows, err := DB.Query(query, role)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var users []models.User
        for rows.Next() {
                var user models.User
                var lastLoginTime, hireDate sql.NullTime
                var fullName, email, phone, address, position, department, notes, emergencyContact sql.NullString

                err := rows.Scan(
                        &user.ID,
                        &user.Username,
                        &user.PasswordHash,
                        &user.Role,
                        &user.CreatedAt,
                        &lastLoginTime,
                        &user.Active,
                        &fullName,
                        &email,
                        &phone,
                        &address,
                        &hireDate,
                        &position,
                        &department,
                        &notes,
                        &emergencyContact,
                )

                if err != nil {
                        return nil, err
                }

                if lastLoginTime.Valid {
                        user.LastLoginAt = lastLoginTime.Time
                }
                
                // Set staff fields from nullable values
                if fullName.Valid {
                        user.FullName = fullName.String
                }
                if email.Valid {
                        user.Email = email.String
                }
                if phone.Valid {
                        user.Phone = phone.String
                }
                if address.Valid {
                        user.Address = address.String
                }
                if hireDate.Valid {
                        user.HireDate = hireDate.Time
                }
                if position.Valid {
                        user.Position = position.String
                }
                if department.Valid {
                        user.Department = department.String
                }
                if notes.Valid {
                        user.Notes = notes.String
                }
                if emergencyContact.Valid {
                        user.EmergencyContact = emergencyContact.String
                }

                users = append(users, user)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return users, nil
}