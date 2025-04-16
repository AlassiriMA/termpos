package main

import (
        "bufio"
        "fmt"
        "os"
        "strconv"
        "strings"
        "time"

        "github.com/olekukonko/tablewriter"
        "github.com/spf13/cobra"
        "termpos/internal/auth"
        "termpos/internal/db"
        "termpos/internal/models"
)

var (
        // Flag variables for staff commands
        staffFullName      string
        staffEmail         string
        staffPhone         string
        staffAddress       string
        staffPosition      string
        staffDepartment    string
        staffNotes         string
        staffEmergContact  string
)

// staffCmd represents the staff command
var staffCmd = &cobra.Command{
        Use:   "staff",
        Short: "Manage staff members",
        Long:  `Staff management commands for adding, updating, viewing, and deleting staff profiles.`,
}

// staffListCmd lists all staff members with details
var staffListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all staff members",
        Long:  `Display a detailed list of all staff members in the system.`,
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to view staff information")
                        return
                }
                
                // Get all users
                users, err := db.GetAllUsers()
                if err != nil {
                        fmt.Printf("Error retrieving staff: %v\n", err)
                        return
                }
                
                // Create table output
                table := tablewriter.NewWriter(os.Stdout)
                table.SetHeader([]string{"ID", "NAME", "ROLE", "POSITION", "DEPARTMENT", "EMAIL", "PHONE", "HIRE DATE"})
                table.SetBorder(false)
                
                for _, user := range users {
                        hireDate := ""
                        if !user.HireDate.IsZero() {
                                hireDate = user.HireDate.Format("2006-01-02")
                        }
                        
                        table.Append([]string{
                                fmt.Sprintf("%d", user.ID),
                                getDisplayName(user),
                                string(user.Role),
                                user.Position,
                                user.Department,
                                user.Email,
                                user.Phone,
                                hireDate,
                        })
                }
                
                table.Render()
                fmt.Printf("Total staff members: %d\n", len(users))
        },
}

// staffAddCmd adds a new staff member
var staffAddCmd = &cobra.Command{
        Use:   "add [username] [role]",
        Short: "Add a new staff member",
        Long:  `Create a new staff member with username, role, and staff profile information.`,
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to add staff members")
                        return
                }
                
                username := args[0]
                roleStr := args[1]
                
                // Validate role
                role := models.Role(roleStr)
                if role != models.RoleAdmin && role != models.RoleManager && role != models.RoleCashier {
                        fmt.Printf("Error: Invalid role: %s (must be admin, manager, or cashier)\n", roleStr)
                        return
                }
                
                var hash string
                
                // Check if a password hash was provided (for automated testing)
                usePasswordHash, _ := cmd.Flags().GetString("password-hash")
                usePassword, _ := cmd.Flags().GetString("password")
                
                if usePasswordHash != "" {
                        // Use provided hash directly (for testing/migration)
                        hash = usePasswordHash
                } else if usePassword != "" {
                        // Use provided password and hash it
                        var err error
                        hash, err = auth.HashPassword(usePassword)
                        if err != nil {
                                fmt.Printf("Error hashing password: %v\n", err)
                                return
                        }
                } else {
                        // Get and confirm password interactively
                        password, err := getPassword("Enter password: ", "Confirm password: ")
                        if err != nil {
                                fmt.Printf("Error: %v\n", err)
                                return
                        }
                        
                        // Hash password
                        hash, err = auth.HashPassword(password)
                        if err != nil {
                                fmt.Printf("Error hashing password: %v\n", err)
                                return
                        }
                }
                
                // Create user with staff details
                user := models.User{
                        Username:     username,
                        PasswordHash: hash,
                        Role:         role,
                        Active:       true,
                        FullName:     staffFullName,
                        Email:        staffEmail,
                        Phone:        staffPhone,
                        Address:      staffAddress,
                        Position:     staffPosition,
                        Department:   staffDepartment,
                        Notes:        staffNotes,
                        EmergencyContact: staffEmergContact,
                        HireDate:     time.Now(),
                }
                
                id, err := db.CreateUser(user)
                if err != nil {
                        if err == db.ErrUserExists {
                                fmt.Printf("Error: User '%s' already exists\n", username)
                        } else {
                                fmt.Printf("Error creating staff member: %v\n", err)
                        }
                        return
                }
                
                fmt.Printf("Staff member '%s' created successfully with ID %d\n", username, id)
                fmt.Printf("Full Name: %s\n", staffFullName)
                fmt.Printf("Role: %s\n", role)
                if staffPosition != "" {
                        fmt.Printf("Position: %s\n", staffPosition)
                }
                if staffDepartment != "" {
                        fmt.Printf("Department: %s\n", staffDepartment)
                }
        },
}

// staffGetCmd gets a staff member's details
var staffGetCmd = &cobra.Command{
        Use:   "get [username or id]",
        Short: "Get staff member details",
        Long:  `Display detailed information for a specific staff member.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to view staff information")
                        return
                }
                
                query := args[0]
                var user models.User
                var err error
                
                // Try to parse as ID first
                if id, err := strconv.Atoi(query); err == nil {
                        user, err = db.GetUserByID(id)
                } else {
                        // Try as username
                        user, err = db.GetUserByUsername(query)
                }
                
                if err != nil {
                        fmt.Printf("Error retrieving staff member: %v\n", err)
                        return
                }
                
                // Display detailed information
                fmt.Println("Staff Member Details:")
                fmt.Println("----------------------------------")
                fmt.Printf("ID: %d\n", user.ID)
                fmt.Printf("Username: %s\n", user.Username)
                fmt.Printf("Full Name: %s\n", getDisplayName(user))
                fmt.Printf("Role: %s\n", user.Role)
                fmt.Printf("Status: %s\n", getStatusText(user.Active))
                
                if user.Position != "" {
                        fmt.Printf("Position: %s\n", user.Position)
                }
                if user.Department != "" {
                        fmt.Printf("Department: %s\n", user.Department)
                }
                if user.Email != "" {
                        fmt.Printf("Email: %s\n", user.Email)
                }
                if user.Phone != "" {
                        fmt.Printf("Phone: %s\n", user.Phone)
                }
                if user.Address != "" {
                        fmt.Printf("Address: %s\n", user.Address)
                }
                if !user.HireDate.IsZero() {
                        fmt.Printf("Hire Date: %s\n", user.HireDate.Format("2006-01-02"))
                        fmt.Printf("Tenure: %s\n", getTenure(user.HireDate))
                }
                if user.EmergencyContact != "" {
                        fmt.Printf("Emergency Contact: %s\n", user.EmergencyContact)
                }
                if user.Notes != "" {
                        fmt.Printf("Notes: %s\n", user.Notes)
                }
                
                // Show system info
                fmt.Println("----------------------------------")
                fmt.Printf("Last Login: %s\n", formatTime(user.LastLoginAt))
                fmt.Printf("Account Created: %s\n", formatTime(user.CreatedAt))
        },
}

// staffUpdateCmd updates a staff member's details
var staffUpdateCmd = &cobra.Command{
        Use:   "update [username or id]",
        Short: "Update staff member details",
        Long:  `Update the profile information for a specific staff member.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to update staff information")
                        return
                }
                
                query := args[0]
                var user models.User
                var err error
                
                // Try to parse as ID first
                if id, err := strconv.Atoi(query); err == nil {
                        user, err = db.GetUserByID(id)
                } else {
                        // Try as username
                        user, err = db.GetUserByUsername(query)
                }
                
                if err != nil {
                        fmt.Printf("Error retrieving staff member: %v\n", err)
                        return
                }
                
                // Update fields only if they're specified in flags
                updated := false
                
                if cmd.Flags().Changed("full-name") {
                        user.FullName = staffFullName
                        updated = true
                }
                if cmd.Flags().Changed("email") {
                        user.Email = staffEmail
                        updated = true
                }
                if cmd.Flags().Changed("phone") {
                        user.Phone = staffPhone
                        updated = true
                }
                if cmd.Flags().Changed("address") {
                        user.Address = staffAddress
                        updated = true
                }
                if cmd.Flags().Changed("position") {
                        user.Position = staffPosition
                        updated = true
                }
                if cmd.Flags().Changed("department") {
                        user.Department = staffDepartment
                        updated = true
                }
                if cmd.Flags().Changed("notes") {
                        user.Notes = staffNotes
                        updated = true
                }
                if cmd.Flags().Changed("emergency-contact") {
                        user.EmergencyContact = staffEmergContact
                        updated = true
                }
                
                if !updated {
                        fmt.Println("No changes specified. Use flags like --full-name, --email, etc. to specify changes.")
                        return
                }
                
                // Save the updated user
                if err := db.UpdateUser(user); err != nil {
                        fmt.Printf("Error updating staff member: %v\n", err)
                        return
                }
                
                fmt.Printf("Staff member '%s' updated successfully\n", user.Username)
        },
}

// staffFindCmd finds staff members by name, department, position or role
var staffFindCmd = &cobra.Command{
        Use:   "find [search term]",
        Short: "Find staff members",
        Long:  `Search for staff members by name, department, position, or role.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to view staff information")
                        return
                }
                
                searchTerm := args[0]
                
                // Try to find by role first (if the search term exactly matches a role)
                if searchTerm == string(models.RoleAdmin) || 
                   searchTerm == string(models.RoleManager) || 
                   searchTerm == string(models.RoleCashier) {
                        
                        users, err := db.FindUsersByRole(models.Role(searchTerm))
                        if err != nil {
                                fmt.Printf("Error searching staff by role: %v\n", err)
                                return
                        }
                        
                        displayStaffList(users, fmt.Sprintf("Staff with role '%s'", searchTerm))
                        return
                }
                
                // Find by name (partial match)
                nameMatches, err := db.FindUsersByFullName(searchTerm)
                if err != nil {
                        fmt.Printf("Error searching staff by name: %v\n", err)
                        return
                }
                
                // Find by department (exact match)
                deptMatches, err := db.FindUsersByDepartment(searchTerm)
                if err != nil {
                        fmt.Printf("Error searching staff by department: %v\n", err)
                        return
                }
                
                // Combine and de-duplicate results
                results := make(map[int]models.User)
                
                for _, user := range nameMatches {
                        results[user.ID] = user
                }
                
                for _, user := range deptMatches {
                        results[user.ID] = user
                }
                
                // Convert map to slice
                var staffList []models.User
                for _, user := range results {
                        staffList = append(staffList, user)
                }
                
                displayStaffList(staffList, fmt.Sprintf("Search results for '%s'", searchTerm))
        },
}

// staffSetRoleCmd sets a staff member's role
var staffSetRoleCmd = &cobra.Command{
        Use:   "set-role [username or id] [role]",
        Short: "Set staff role",
        Long:  `Update a staff member's role (admin, manager, cashier).`,
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("user:manage"); err != nil {
                        fmt.Println("Error: You don't have permission to update staff roles")
                        return
                }
                
                query := args[0]
                roleStr := args[1]
                
                // Validate role
                role := models.Role(roleStr)
                if role != models.RoleAdmin && role != models.RoleManager && role != models.RoleCashier {
                        fmt.Printf("Error: Invalid role: %s (must be admin, manager, or cashier)\n", roleStr)
                        return
                }
                
                var user models.User
                var err error
                
                // Try to parse as ID first
                if id, err := strconv.Atoi(query); err == nil {
                        user, err = db.GetUserByID(id)
                } else {
                        // Try as username
                        user, err = db.GetUserByUsername(query)
                }
                
                if err != nil {
                        fmt.Printf("Error retrieving staff member: %v\n", err)
                        return
                }
                
                // Check if user is trying to change their own role
                session := auth.GetCurrentUser()
                if session != nil && session.Username == user.Username {
                        reader := bufio.NewReader(os.Stdin)
                        fmt.Print("Warning: You are changing your own role. This may affect your permissions. Continue? (y/N): ")
                        confirm, _ := reader.ReadString('\n')
                        confirm = strings.TrimSpace(strings.ToLower(confirm))
                        if confirm != "y" {
                                fmt.Println("Operation cancelled")
                                return
                        }
                }
                
                // Update role
                oldRole := user.Role
                user.Role = role
                if err := db.UpdateUser(user); err != nil {
                        fmt.Printf("Error updating staff role: %v\n", err)
                        return
                }
                
                fmt.Printf("Staff member '%s' role updated from '%s' to '%s'\n", getDisplayName(user), oldRole, role)
        },
}

// Initialize the staff commands
func initStaffCommands() {
        // Add the staff command to the root command
        rootCmd.AddCommand(staffCmd)
        
        // Add staff subcommands
        staffCmd.AddCommand(staffListCmd)
        staffCmd.AddCommand(staffAddCmd)
        staffCmd.AddCommand(staffGetCmd)
        staffCmd.AddCommand(staffUpdateCmd)
        staffCmd.AddCommand(staffFindCmd)
        staffCmd.AddCommand(staffSetRoleCmd)
        
        // Add flags to the staff add command
        staffAddCmd.Flags().StringVar(&staffFullName, "full-name", "", "Staff member's full name")
        staffAddCmd.Flags().StringVar(&staffEmail, "email", "", "Staff member's email address")
        staffAddCmd.Flags().StringVar(&staffPhone, "phone", "", "Staff member's phone number")
        staffAddCmd.Flags().StringVar(&staffAddress, "address", "", "Staff member's address")
        staffAddCmd.Flags().StringVar(&staffPosition, "position", "", "Staff member's position/job title")
        staffAddCmd.Flags().StringVar(&staffDepartment, "department", "", "Staff member's department")
        staffAddCmd.Flags().StringVar(&staffNotes, "notes", "", "Additional notes about the staff member")
        staffAddCmd.Flags().StringVar(&staffEmergContact, "emergency-contact", "", "Staff member's emergency contact information")
        
        // Add password flags for automated testing
        staffAddCmd.Flags().String("password", "", "Pre-defined password to use (instead of interactive prompt)")
        staffAddCmd.Flags().String("password-hash", "", "Pre-computed password hash to use directly (for testing/migration)")
        
        // Add flags to the staff update command
        staffUpdateCmd.Flags().StringVar(&staffFullName, "full-name", "", "Staff member's full name")
        staffUpdateCmd.Flags().StringVar(&staffEmail, "email", "", "Staff member's email address")
        staffUpdateCmd.Flags().StringVar(&staffPhone, "phone", "", "Staff member's phone number")
        staffUpdateCmd.Flags().StringVar(&staffAddress, "address", "", "Staff member's address")
        staffUpdateCmd.Flags().StringVar(&staffPosition, "position", "", "Staff member's position/job title")
        staffUpdateCmd.Flags().StringVar(&staffDepartment, "department", "", "Staff member's department")
        staffUpdateCmd.Flags().StringVar(&staffNotes, "notes", "", "Additional notes about the staff member")
        staffUpdateCmd.Flags().StringVar(&staffEmergContact, "emergency-contact", "", "Staff member's emergency contact information")
        
        // Make full-name required for staff add command
        staffAddCmd.MarkFlagRequired("full-name")
}

// Helper functions

// getDisplayName returns the full name if available, otherwise the username
func getDisplayName(user models.User) string {
        if user.FullName != "" {
                return user.FullName
        }
        return user.Username
}

// getStatusText returns the text representation of an active status
func getStatusText(active bool) string {
        if active {
                return "Active"
        }
        return "Inactive"
}

// formatTime formats a time or returns "Never" if the time is zero
func formatTime(t time.Time) string {
        if t.IsZero() {
                return "Never"
        }
        return t.Format("2006-01-02 15:04:05")
}

// getTenure calculates and formats the tenure of a staff member
func getTenure(hireDate time.Time) string {
        if hireDate.IsZero() {
                return "Unknown"
        }
        
        now := time.Now()
        years := now.Year() - hireDate.Year()
        
        // Adjust for not having reached the hire date anniversary yet
        if now.Month() < hireDate.Month() || (now.Month() == hireDate.Month() && now.Day() < hireDate.Day()) {
                years--
        }
        
        months := int(now.Month() - hireDate.Month())
        if months < 0 {
                months += 12
        }
        
        if now.Day() < hireDate.Day() {
                months--
                if months < 0 {
                        months += 12
                }
        }
        
        if years > 0 {
                if months > 0 {
                        return fmt.Sprintf("%d years, %d months", years, months)
                }
                return fmt.Sprintf("%d years", years)
        }
        
        if months > 0 {
                return fmt.Sprintf("%d months", months)
        }
        
        days := now.Sub(hireDate).Hours() / 24
        return fmt.Sprintf("%.0f days", days)
}

// displayStaffList shows a table of staff members
func displayStaffList(users []models.User, title string) {
        if len(users) == 0 {
                fmt.Println("No staff members found")
                return
        }
        
        fmt.Printf("\n%s (%d results):\n", title, len(users))
        
        table := tablewriter.NewWriter(os.Stdout)
        table.SetHeader([]string{"ID", "NAME", "ROLE", "POSITION", "DEPARTMENT", "EMAIL", "PHONE"})
        table.SetBorder(false)
        
        for _, user := range users {
                table.Append([]string{
                        fmt.Sprintf("%d", user.ID),
                        getDisplayName(user),
                        string(user.Role),
                        user.Position,
                        user.Department,
                        user.Email,
                        user.Phone,
                })
        }
        
        table.Render()
}