package main

import (
        "bufio"
        "fmt"
        "os"
        "strings"

        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/models"
)

var (
        loginCmd = &cobra.Command{
                Use:   "login",
                Short: "Log in to the system",
                Long:  "Authenticate with username and password to access the system",
                RunE:  runLogin,
        }

        logoutCmd = &cobra.Command{
                Use:   "logout",
                Short: "Log out of the system",
                Long:  "End your current session and log out of the system",
                Run: func(cmd *cobra.Command, args []string) {
                        auth.Logout()
                        fmt.Println("You have been logged out")
                },
        }

        userCmd = &cobra.Command{
                Use:   "user",
                Short: "User management commands",
                Long:  "Manage system users with commands to add, update, list and delete users",
        }

        userListCmd = &cobra.Command{
                Use:   "list",
                Short: "List all users",
                Long:  "Display a list of all users in the system",
                RunE:  runUserList,
        }

        userAddCmd = &cobra.Command{
                Use:   "add [username] [role]",
                Short: "Add a new user",
                Long:  "Create a new user with the specified username and role (admin, manager, cashier)",
                Args:  cobra.ExactArgs(2),
                RunE:  runUserAdd,
        }

        userUpdateCmd = &cobra.Command{
                Use:   "update [username] [role]",
                Short: "Update a user's role",
                Long:  "Change the role of an existing user",
                Args:  cobra.ExactArgs(2),
                RunE:  runUserUpdate,
        }

        userActivateCmd = &cobra.Command{
                Use:   "activate [username]",
                Short: "Activate a user account",
                Long:  "Enable a previously deactivated user account",
                Args:  cobra.ExactArgs(1),
                RunE:  runUserActivate,
        }

        userDeactivateCmd = &cobra.Command{
                Use:   "deactivate [username]",
                Short: "Deactivate a user account",
                Long:  "Disable a user account without deleting it",
                Args:  cobra.ExactArgs(1),
                RunE:  runUserDeactivate,
        }

        userResetPasswordCmd = &cobra.Command{
                Use:   "reset-password [username]",
                Short: "Reset a user's password",
                Long:  "Reset the password for the specified user",
                Args:  cobra.ExactArgs(1),
                RunE:  runUserResetPassword,
        }
)

// initUserCommands adds the user-related commands to the root command
func initUserCommands() {
        // Add login and logout commands
        rootCmd.AddCommand(loginCmd)
        rootCmd.AddCommand(logoutCmd)

        // Add user management commands
        rootCmd.AddCommand(userCmd)
        userCmd.AddCommand(userListCmd)
        userCmd.AddCommand(userAddCmd)
        userCmd.AddCommand(userUpdateCmd)
        userCmd.AddCommand(userActivateCmd)
        userCmd.AddCommand(userDeactivateCmd)
        userCmd.AddCommand(userResetPasswordCmd)
}

// runLogin handles the login command
func runLogin(cmd *cobra.Command, args []string) error {
        // Check if already logged in
        if auth.IsAuthenticated() {
                session := auth.GetCurrentUser()
                fmt.Printf("Already logged in as %s (Role: %s)\n", session.Username, session.Role)
                return nil
        }

        reader := bufio.NewReader(os.Stdin)

        // Get username
        fmt.Print("Username: ")
        username, err := reader.ReadString('\n')
        if err != nil {
                return fmt.Errorf("failed to read username: %w", err)
        }
        username = strings.TrimSpace(username)

        // Get password
        fmt.Print("Password: ")
        password, err := reader.ReadString('\n')
        if err != nil {
                return fmt.Errorf("failed to read password: %w", err)
        }
        password = strings.TrimSpace(password)

        // Authenticate
        session, err := auth.Login(username, password, db.GetUserByUsername, db.UpdateLastLogin)
        if err != nil {
                if err == auth.ErrInvalidCredentials {
                        return fmt.Errorf("invalid username or password")
                }
                if err == auth.ErrUserInactive {
                        return fmt.Errorf("this account has been deactivated")
                }
                return fmt.Errorf("login failed: %w", err)
        }

        fmt.Printf("Successfully logged in as %s (Role: %s)\n", session.Username, session.Role)
        return nil
}

// runUserList handles the user list command
func runUserList(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        // Get all users
        users, err := db.GetAllUsers()
        if err != nil {
                return fmt.Errorf("failed to retrieve users: %w", err)
        }

        // Print users
        fmt.Println("Users:")
        fmt.Println("------------------------------")
        fmt.Printf("%-5s | %-15s | %-10s | %-8s\n", "ID", "USERNAME", "ROLE", "STATUS")
        fmt.Println("------------------------------")

        for _, user := range users {
                status := "Active"
                if !user.Active {
                        status = "Inactive"
                }

                fmt.Printf("%-5d | %-15s | %-10s | %-8s\n", user.ID, user.Username, user.Role, status)
        }

        return nil
}

// runUserAdd handles the user add command
func runUserAdd(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        username := args[0]
        roleStr := args[1]

        // Validate role
        role := models.Role(roleStr)
        if role != models.RoleAdmin && role != models.RoleManager && role != models.RoleCashier {
                return fmt.Errorf("invalid role: %s (must be admin, manager, or cashier)", roleStr)
        }

        // Get and confirm password
        password, err := getPassword("Enter password: ", "Confirm password: ")
        if err != nil {
                return err
        }

        // Hash password
        hash, err := auth.HashPassword(password)
        if err != nil {
                return fmt.Errorf("failed to hash password: %w", err)
        }

        // Create user
        user := models.User{
                Username:     username,
                PasswordHash: hash,
                Role:         role,
                Active:       true,
        }

        id, err := db.CreateUser(user)
        if err != nil {
                if err == db.ErrUserExists {
                        return fmt.Errorf("user '%s' already exists", username)
                }
                return fmt.Errorf("failed to create user: %w", err)
        }

        fmt.Printf("User '%s' created successfully with ID %d\n", username, id)
        return nil
}

// runUserUpdate handles the user update command
func runUserUpdate(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        username := args[0]
        roleStr := args[1]

        // Validate role
        role := models.Role(roleStr)
        if role != models.RoleAdmin && role != models.RoleManager && role != models.RoleCashier {
                return fmt.Errorf("invalid role: %s (must be admin, manager, or cashier)", roleStr)
        }

        // Get user
        user, err := db.GetUserByUsername(username)
        if err != nil {
                if err == db.ErrUserNotFound {
                        return fmt.Errorf("user '%s' not found", username)
                }
                return fmt.Errorf("failed to retrieve user: %w", err)
        }

        // Update role
        user.Role = role
        if err := db.UpdateUser(user); err != nil {
                return fmt.Errorf("failed to update user: %w", err)
        }

        fmt.Printf("User '%s' role updated to '%s'\n", username, role)
        return nil
}

// runUserActivate handles the user activate command
func runUserActivate(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        username := args[0]

        // Get user
        user, err := db.GetUserByUsername(username)
        if err != nil {
                if err == db.ErrUserNotFound {
                        return fmt.Errorf("user '%s' not found", username)
                }
                return fmt.Errorf("failed to retrieve user: %w", err)
        }

        // Update active status
        user.Active = true
        if err := db.UpdateUser(user); err != nil {
                return fmt.Errorf("failed to activate user: %w", err)
        }

        fmt.Printf("User '%s' has been activated\n", username)
        return nil
}

// runUserDeactivate handles the user deactivate command
func runUserDeactivate(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        username := args[0]

        // Get user
        user, err := db.GetUserByUsername(username)
        if err != nil {
                if err == db.ErrUserNotFound {
                        return fmt.Errorf("user '%s' not found", username)
                }
                return fmt.Errorf("failed to retrieve user: %w", err)
        }

        // Check if user is trying to deactivate themselves
        session := auth.GetCurrentUser()
        if session != nil && session.Username == user.Username {
                return fmt.Errorf("you cannot deactivate your own account")
        }

        // Update active status
        user.Active = false
        if err := db.UpdateUser(user); err != nil {
                return fmt.Errorf("failed to deactivate user: %w", err)
        }

        fmt.Printf("User '%s' has been deactivated\n", username)
        return nil
}

// runUserResetPassword handles the user reset-password command
func runUserResetPassword(cmd *cobra.Command, args []string) error {
        // Check permissions
        if err := auth.RequirePermission("user:manage"); err != nil {
                return err
        }

        username := args[0]

        // Get user
        user, err := db.GetUserByUsername(username)
        if err != nil {
                if err == db.ErrUserNotFound {
                        return fmt.Errorf("user '%s' not found", username)
                }
                return fmt.Errorf("failed to retrieve user: %w", err)
        }

        // Get and confirm new password
        password, err := getPassword("Enter new password: ", "Confirm new password: ")
        if err != nil {
                return err
        }

        // Hash password
        hash, err := auth.HashPassword(password)
        if err != nil {
                return fmt.Errorf("failed to hash password: %w", err)
        }

        // Update password
        if err := db.UpdateUserPassword(user.ID, hash); err != nil {
                return fmt.Errorf("failed to update password: %w", err)
        }

        fmt.Printf("Password for user '%s' has been reset\n", username)
        return nil
}

// getPassword prompts for a password and confirms it
func getPassword(prompt, confirmPrompt string) (string, error) {
        reader := bufio.NewReader(os.Stdin)

        // Prompt for password
        fmt.Print(prompt)
        password, err := reader.ReadString('\n')
        if err != nil {
                return "", fmt.Errorf("failed to read password: %w", err)
        }
        password = strings.TrimSpace(password)

        // Prompt for confirmation
        fmt.Print(confirmPrompt)
        confirm, err := reader.ReadString('\n')
        if err != nil {
                return "", fmt.Errorf("failed to read password confirmation: %w", err)
        }
        confirm = strings.TrimSpace(confirm)

        // Validate passwords match
        if password != confirm {
                return "", fmt.Errorf("passwords do not match")
        }

        // Validate password length
        if len(password) < 6 {
                return "", fmt.Errorf("password must be at least 6 characters")
        }

        return password, nil
}