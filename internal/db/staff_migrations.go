package db

import "time"

// alterUsersTableForStaff adds staff-related columns to the users table
func alterUsersTableForStaff() error {
	// SQLite doesn't support adding multiple columns in a single ALTER TABLE statement
	// so we need to execute multiple statements
	queries := []string{
		"ALTER TABLE users ADD COLUMN full_name TEXT;",
		"ALTER TABLE users ADD COLUMN email TEXT;",
		"ALTER TABLE users ADD COLUMN phone TEXT;",
		"ALTER TABLE users ADD COLUMN address TEXT;",
		"ALTER TABLE users ADD COLUMN hire_date TIMESTAMP;",
		"ALTER TABLE users ADD COLUMN position TEXT;",
		"ALTER TABLE users ADD COLUMN department TEXT;",
		"ALTER TABLE users ADD COLUMN notes TEXT;",
		"ALTER TABLE users ADD COLUMN emergency_contact TEXT;",
	}
	
	for _, query := range queries {
		// Execute the query and ignore "duplicate column" errors
		_, err := DB.Exec(query)
		if err != nil {
			// Check if the error is just that the column already exists
			if err.Error() == "duplicate column name: full_name" || 
			   err.Error() == "duplicate column name: email" ||
			   err.Error() == "duplicate column name: phone" ||
			   err.Error() == "duplicate column name: address" ||
			   err.Error() == "duplicate column name: hire_date" ||
			   err.Error() == "duplicate column name: position" ||
			   err.Error() == "duplicate column name: department" ||
			   err.Error() == "duplicate column name: notes" ||
			   err.Error() == "duplicate column name: emergency_contact" {
				continue
			}
			return err
		}
	}
	
	// Update existing admin user with some defaults if full_name is null
	_, err := DB.Exec(`
		UPDATE users 
		SET full_name = 'System Administrator',
		    position = 'Administrator',
		    department = 'Management',
		    hire_date = ?
		WHERE username = 'admin' AND full_name IS NULL
	`, time.Now())
	if err != nil {
		return err
	}
	
	return nil
}