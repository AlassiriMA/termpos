package main

import (
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
        // Customer command flags
        customerEmail         string
        customerPhone         string
        customerAddress       string
        customerNotes         string
        customerBirthday      string
        customerPreferredProd string
        loyaltyPoints         int
        loyaltyTier           string
)

// customerCmd represents the customer command
var customerCmd = &cobra.Command{
        Use:   "customer",
        Short: "Manage customers and loyalty program",
        Long:  `Manage customer records, view purchase history, and administer the loyalty program.`,
}

// customerAddCmd adds a new customer
var customerAddCmd = &cobra.Command{
        Use:   "add [name]",
        Short: "Add a new customer",
        Long:  `Add a new customer to the database with name and optional details like email, phone, address, etc.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:create"); err != nil {
                        fmt.Println("Error: You don't have permission to add customers")
                        return
                }

                name := args[0]
                
                // Basic validation
                if name == "" {
                        fmt.Println("Error: Customer name cannot be empty")
                        return
                }
                
                if customerEmail == "" && customerPhone == "" {
                        fmt.Println("Error: At least one contact method (email or phone) is required")
                        return
                }
                
                // Create customer object
                customer := models.Customer{
                        Name:              name,
                        Email:             customerEmail,
                        Phone:             customerPhone,
                        Address:           customerAddress,
                        Notes:             customerNotes,
                        Birthday:          customerBirthday,
                        PreferredProducts: customerPreferredProd,
                        JoinDate:          time.Now(),
                        LoyaltyPoints:     loyaltyPoints,
                        LoyaltyTier:       loyaltyTier,
                }
                
                // If no tier specified, set based on points
                if customer.LoyaltyTier == "" {
                        customer.LoyaltyTier = models.GetLoyaltyTierName(customer.LoyaltyPoints)
                }
                
                // Add customer
                id, err := db.AddCustomer(customer)
                if err != nil {
                        fmt.Printf("Error adding customer: %v\n", err)
                        return
                }
                
                fmt.Printf("Customer added successfully with ID: %d\n", id)
                
                // Show customer details
                customer.ID = id
                displayCustomerDetails(customer)
        },
}

// customerGetCmd gets a customer by ID, email, or phone
var customerGetCmd = &cobra.Command{
        Use:   "get [id|email|phone]",
        Short: "Get customer details",
        Long:  `Retrieve a customer's details by ID, email address, or phone number.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:read"); err != nil {
                        fmt.Println("Error: You don't have permission to view customers")
                        return
                }
                
                query := args[0]
                var customer models.Customer
                var err error
                
                // Try to interpret as ID first
                if id, err := strconv.Atoi(query); err == nil {
                        customer, err = db.GetCustomer(id)
                } else if strings.Contains(query, "@") {
                        // Looks like an email
                        customer, err = db.GetCustomerByEmail(query)
                } else {
                        // Try as phone number
                        customer, err = db.GetCustomerByPhone(query)
                }
                
                if err != nil {
                        fmt.Printf("Error retrieving customer: %v\n", err)
                        return
                }
                
                displayCustomerDetails(customer)
        },
}

// customerUpdateCmd updates customer information
var customerUpdateCmd = &cobra.Command{
        Use:   "update [id]",
        Short: "Update customer details",
        Long:  `Update a customer's information using the provided ID and flags for fields to update.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:update"); err != nil {
                        fmt.Println("Error: You don't have permission to update customers")
                        return
                }
                
                id, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: ID must be a number")
                        return
                }
                
                // Get current customer
                customer, err := db.GetCustomer(id)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                }
                
                // Update fields if provided
                if cmd.Flags().Changed("email") {
                        customer.Email = customerEmail
                }
                if cmd.Flags().Changed("phone") {
                        customer.Phone = customerPhone
                }
                if cmd.Flags().Changed("address") {
                        customer.Address = customerAddress
                }
                if cmd.Flags().Changed("notes") {
                        customer.Notes = customerNotes
                }
                if cmd.Flags().Changed("birthday") {
                        customer.Birthday = customerBirthday
                }
                if cmd.Flags().Changed("preferred-products") {
                        customer.PreferredProducts = customerPreferredProd
                }
                if cmd.Flags().Changed("loyalty-points") {
                        customer.LoyaltyPoints = loyaltyPoints
                        // Update tier based on points
                        customer.LoyaltyTier = models.GetLoyaltyTierName(loyaltyPoints)
                }
                if cmd.Flags().Changed("loyalty-tier") {
                        customer.LoyaltyTier = loyaltyTier
                }
                
                // Update customer
                err = db.UpdateCustomer(customer)
                if err != nil {
                        fmt.Printf("Error updating customer: %v\n", err)
                        return
                }
                
                fmt.Printf("Customer ID %d updated successfully\n", id)
                displayCustomerDetails(customer)
        },
}

// customerDeleteCmd deletes a customer
var customerDeleteCmd = &cobra.Command{
        Use:   "delete [id]",
        Short: "Delete a customer",
        Long:  `Delete a customer from the database by ID.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:delete"); err != nil {
                        fmt.Println("Error: You don't have permission to delete customers")
                        return
                }
                
                id, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: ID must be a number")
                        return
                }
                
                // Confirm deletion
                fmt.Printf("Are you sure you want to delete customer ID %d? This cannot be undone. (y/N): ", id)
                var confirm string
                fmt.Scanln(&confirm)
                if strings.ToLower(confirm) != "y" {
                        fmt.Println("Deletion cancelled")
                        return
                }
                
                // Delete customer
                err = db.DeleteCustomer(id)
                if err != nil {
                        fmt.Printf("Error deleting customer: %v\n", err)
                        return
                }
                
                fmt.Printf("Customer ID %d deleted successfully\n", id)
        },
}

// customerListCmd lists all customers
var customerListCmd = &cobra.Command{
        Use:   "list [filter]",
        Short: "List customers",
        Long:  `List all customers with optional filtering by name, email, or phone.`,
        Args:  cobra.MaximumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:read"); err != nil {
                        fmt.Println("Error: You don't have permission to view customers")
                        return
                }
                
                filter := ""
                if len(args) > 0 {
                        filter = args[0]
                }
                
                customers, err := db.ListCustomers(filter, 100, 0)
                if err != nil {
                        fmt.Printf("Error listing customers: %v\n", err)
                        return
                }
                
                if len(customers) == 0 {
                        fmt.Println("No customers found")
                        return
                }
                
                // Display as table
                table := tablewriter.NewWriter(os.Stdout)
                table.SetHeader([]string{"ID", "NAME", "PHONE", "EMAIL", "POINTS", "TIER"})
                table.SetBorder(false)
                
                for _, c := range customers {
                        table.Append([]string{
                                fmt.Sprintf("%d", c.ID),
                                c.Name,
                                c.Phone,
                                c.Email,
                                fmt.Sprintf("%d", c.LoyaltyPoints),
                                c.LoyaltyTier,
                        })
                }
                
                table.Render()
                fmt.Printf("Total customers: %d\n", len(customers))
        },
}

// customerHistoryCmd shows a customer's purchase history
var customerHistoryCmd = &cobra.Command{
        Use:   "history [id]",
        Short: "Show customer purchase history",
        Long:  `Show a customer's purchase history including products, points earned, and rewards used.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:read"); err != nil {
                        fmt.Println("Error: You don't have permission to view customer history")
                        return
                }
                
                id, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: ID must be a number")
                        return
                }
                
                // Get customer for display
                customer, err := db.GetCustomer(id)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                }
                
                // Get purchase history
                history, err := db.GetCustomerPurchaseHistory(id, 20)
                if err != nil {
                        fmt.Printf("Error retrieving purchase history: %v\n", err)
                        return
                }
                
                if len(history) == 0 {
                        fmt.Printf("No purchase history found for %s (ID: %d)\n", customer.Name, customer.ID)
                        return
                }
                
                // Display customer summary
                fmt.Printf("Purchase History for %s (ID: %d)\n", customer.Name, customer.ID)
                fmt.Printf("Loyalty Tier: %s, Points: %d\n", customer.LoyaltyTier, customer.LoyaltyPoints)
                fmt.Printf("Total Spent: $%.2f\n\n", customer.TotalPurchases)
                
                // Display purchase history
                table := tablewriter.NewWriter(os.Stdout)
                table.SetHeader([]string{"DATE", "PRODUCT", "QTY", "TOTAL", "POINTS", "USED", "REWARD"})
                table.SetBorder(false)
                
                for _, p := range history {
                        reward := "-"
                        if r, ok := p["reward_name"]; ok && r != nil {
                                reward = r.(string)
                        }
                        
                        table.Append([]string{
                                p["sale_date"].(string),
                                p["product_name"].(string),
                                fmt.Sprintf("%d", p["quantity"].(int)),
                                fmt.Sprintf("$%.2f", p["total"].(float64)),
                                fmt.Sprintf("%d", p["points_earned"].(int)),
                                fmt.Sprintf("%d", p["points_used"].(int)),
                                reward,
                        })
                }
                
                table.Render()
        },
}

// customerAddPointsCmd adds loyalty points to a customer
var customerAddPointsCmd = &cobra.Command{
        Use:   "add-points [id] [points]",
        Short: "Add loyalty points",
        Long:  `Add loyalty points to a customer's account and update their loyalty tier.`,
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:update"); err != nil {
                        fmt.Println("Error: You don't have permission to update customer points")
                        return
                }
                
                id, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: Customer ID must be a number")
                        return
                }
                
                points, err := strconv.Atoi(args[1])
                if err != nil || points <= 0 {
                        fmt.Println("Error: Points must be a positive number")
                        return
                }
                
                // Get current customer
                customer, err := db.GetCustomer(id)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                }
                
                // Update points
                customer.LoyaltyPoints += points
                customer.LoyaltyTier = models.GetLoyaltyTierName(customer.LoyaltyPoints)
                
                // Save customer
                err = db.UpdateCustomer(customer)
                if err != nil {
                        fmt.Printf("Error updating customer points: %v\n", err)
                        return
                }
                
                fmt.Printf("Added %d points to %s's account\n", points, customer.Name)
                fmt.Printf("New balance: %d points, Tier: %s\n", customer.LoyaltyPoints, customer.LoyaltyTier)
        },
}

// loyaltyRewardsCmd shows available loyalty rewards
var loyaltyRewardsCmd = &cobra.Command{
        Use:   "rewards",
        Short: "Show loyalty rewards",
        Long:  `Show available loyalty rewards that customers can redeem with their points.`,
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:read"); err != nil {
                        fmt.Println("Error: You don't have permission to view loyalty rewards")
                        return
                }
                
                rewards, err := db.GetLoyaltyRewards(true)
                if err != nil {
                        fmt.Printf("Error retrieving loyalty rewards: %v\n", err)
                        return
                }
                
                if len(rewards) == 0 {
                        fmt.Println("No loyalty rewards found")
                        return
                }
                
                // Display as table
                table := tablewriter.NewWriter(os.Stdout)
                table.SetHeader([]string{"ID", "NAME", "DESCRIPTION", "POINTS", "VALUE", "VALID"})
                table.SetBorder(false)
                
                for _, r := range rewards {
                        value := fmt.Sprintf("$%.2f", r.DiscountValue)
                        if r.IsPercentage {
                                value = fmt.Sprintf("%.0f%%", r.DiscountValue)
                        }
                        
                        table.Append([]string{
                                fmt.Sprintf("%d", r.ID),
                                r.Name,
                                r.Description,
                                fmt.Sprintf("%d", r.PointsCost),
                                value,
                                fmt.Sprintf("%d days", r.ValidDays),
                        })
                }
                
                table.Render()
        },
}

// customerRedeemRewardCmd redeems a reward for a customer
var customerRedeemRewardCmd = &cobra.Command{
        Use:   "redeem [customer_id] [reward_id]",
        Short: "Redeem a loyalty reward",
        Long:  `Redeem a loyalty reward for a customer using their loyalty points.`,
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:update"); err != nil {
                        fmt.Println("Error: You don't have permission to redeem rewards")
                        return
                }
                
                customerID, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: Customer ID must be a number")
                        return
                }
                
                rewardID, err := strconv.Atoi(args[1])
                if err != nil {
                        fmt.Println("Error: Reward ID must be a number")
                        return
                }
                
                // Get customer for display
                customer, err := db.GetCustomer(customerID)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                }
                
                // Confirm redemption
                fmt.Printf("Redeem reward for %s (ID: %d, Points: %d)? (y/N): ", 
                        customer.Name, customer.ID, customer.LoyaltyPoints)
                var confirm string
                fmt.Scanln(&confirm)
                if strings.ToLower(confirm) != "y" {
                        fmt.Println("Redemption cancelled")
                        return
                }
                
                // Redeem reward
                reward, err := db.RedeemLoyaltyReward(customerID, rewardID)
                if err != nil {
                        fmt.Printf("Error redeeming reward: %v\n", err)
                        return
                }
                
                // Get updated customer for display
                customer, _ = db.GetCustomer(customerID)
                
                fmt.Printf("Successfully redeemed '%s' for %s\n", reward.Name, customer.Name)
                fmt.Printf("Points used: %d, Remaining points: %d\n", reward.PointsCost, customer.LoyaltyPoints)
                
                value := fmt.Sprintf("$%.2f", reward.DiscountValue)
                if reward.IsPercentage {
                        value = fmt.Sprintf("%.0f%%", reward.DiscountValue)
                }
                
                fmt.Printf("Reward value: %s, Valid for: %d days\n", value, reward.ValidDays)
                fmt.Println("Use during checkout to apply the reward")
        },
}

// customerLoyaltyStatusCmd shows a customer's loyalty program status
var customerLoyaltyStatusCmd = &cobra.Command{
        Use:   "loyalty-status [id]",
        Short: "Show loyalty program status",
        Long:  `Display a customer's loyalty program status, including tier, points, available rewards, and tier benefits.`,
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:read"); err != nil {
                        fmt.Println("Error: You don't have permission to view customer loyalty status")
                        return
                }
                
                id, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: ID must be a number")
                        return
                }
                
                // Get customer
                customer, err := db.GetCustomer(id)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                }
                
                // Get available rewards
                rewards, err := db.GetLoyaltyRewards(true)
                if err != nil {
                        fmt.Printf("Error retrieving loyalty rewards: %v\n", err)
                        return
                }
                
                // Display loyalty status in a nice format
                fmt.Println("==================================")
                fmt.Printf("LOYALTY STATUS: %s\n", customer.Name)
                fmt.Println("==================================")
                fmt.Printf("Loyalty Tier: %s\n", customer.LoyaltyTier)
                fmt.Printf("Points Balance: %d points\n", customer.LoyaltyPoints)
                
                // Display tier benefits
                fmt.Println("\nTier Benefits:")
                fmt.Printf("- Discount: %.0f%% off purchases\n", models.GetLoyaltyTierDiscount(customer.LoyaltyTier)*100)
                fmt.Printf("- Points Multiplier: %.1fx points on purchases\n", models.GetLoyaltyTierMultiplier(customer.LoyaltyTier))
                
                // Display next tier if not at highest
                if customer.LoyaltyTier != "Platinum" {
                        var nextTier string
                        var pointsNeeded int
                        
                        switch customer.LoyaltyTier {
                        case "Bronze":
                                nextTier = "Silver"
                                pointsNeeded = 200 - customer.LoyaltyPoints
                        case "Silver":
                                nextTier = "Gold"
                                pointsNeeded = 500 - customer.LoyaltyPoints
                        case "Gold":
                                nextTier = "Platinum"
                                pointsNeeded = 1000 - customer.LoyaltyPoints
                        }
                        
                        fmt.Printf("\nNext Tier: %s (need %d more points)\n", nextTier, pointsNeeded)
                }
                
                // Display available rewards
                fmt.Println("\nAvailable Rewards:")
                if len(rewards) == 0 {
                        fmt.Println("No rewards currently available")
                } else {
                        table := tablewriter.NewWriter(os.Stdout)
                        table.SetHeader([]string{"ID", "NAME", "COST", "VALUE", "ELIGIBLE"})
                        table.SetBorder(false)
                        
                        for _, r := range rewards {
                                eligible := "No"
                                if customer.LoyaltyPoints >= r.PointsCost {
                                        eligible = "Yes"
                                }
                                
                                value := fmt.Sprintf("$%.2f", r.DiscountValue)
                                if r.IsPercentage {
                                        value = fmt.Sprintf("%.0f%%", r.DiscountValue*100)
                                }
                                
                                table.Append([]string{
                                        fmt.Sprintf("%d", r.ID),
                                        r.Name,
                                        fmt.Sprintf("%d pts", r.PointsCost),
                                        value,
                                        eligible,
                                })
                        }
                        
                        table.Render()
                }
                
                // Show recent activity
                fmt.Println("\nRecent Activity:")
                history, err := db.GetCustomerPurchaseHistory(id, 5)
                if err != nil {
                        fmt.Printf("Error retrieving purchase history: %v\n", err)
                } else if len(history) == 0 {
                        fmt.Println("No recent activity")
                } else {
                        table := tablewriter.NewWriter(os.Stdout)
                        table.SetHeader([]string{"DATE", "POINTS +/-", "DETAILS"})
                        table.SetBorder(false)
                        
                        for _, p := range history {
                                pointsNet := p["points_earned"].(int) - p["points_used"].(int)
                                pointsStr := fmt.Sprintf("%+d", pointsNet)
                                
                                details := fmt.Sprintf("%s x%d ($%.2f)", 
                                        p["product_name"].(string), 
                                        p["quantity"].(int),
                                        p["total"].(float64))
                                
                                table.Append([]string{
                                        p["sale_date"].(string),
                                        pointsStr,
                                        details,
                                })
                        }
                        
                        table.Render()
                }
                
                fmt.Println("==================================")
        },
}

// customerLinkSaleCmd links a sale to a customer
var customerLinkSaleCmd = &cobra.Command{
        Use:   "link-sale [customer_id] [sale_id] [points_earned]",
        Short: "Link a sale to a customer",
        Long:  `Link an existing sale to a customer and award loyalty points.`,
        Args:  cobra.MinimumNArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
                // Check permissions
                if err := auth.RequirePermission("customer:update"); err != nil {
                        fmt.Println("Error: You don't have permission to link sales")
                        return
                }
                
                customerID, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Println("Error: Customer ID must be a number")
                        return
                }
                
                saleID, err := strconv.Atoi(args[1])
                if err != nil {
                        fmt.Println("Error: Sale ID must be a number")
                        return
                }
                
                // Get points earned (optional argument)
                pointsEarned := 0
                if len(args) > 2 {
                        if p, err := strconv.Atoi(args[2]); err == nil && p > 0 {
                                pointsEarned = p
                        }
                }
                
                // If points not specified, calculate based on sale amount
                if pointsEarned == 0 {
                        // Get sale amount
                        var total float64
                        err = db.DB.QueryRow("SELECT total FROM sales WHERE id = ?", saleID).Scan(&total)
                        if err != nil {
                                fmt.Printf("Error retrieving sale: %v\n", err)
                                return
                        }
                        
                        // Get customer's tier for points multiplier
                        customer, err := db.GetCustomer(customerID)
                        if err != nil {
                                fmt.Printf("Error retrieving customer: %v\n", err)
                                return
                        }
                        
                        // Calculate points based on tier multiplier
                        multiplier := models.GetLoyaltyTierMultiplier(customer.LoyaltyTier)
                        pointsEarned = models.CalculatePointsForPurchase(total, multiplier)
                }
                
                // Link the sale
                err = db.LinkSaleToCustomer(saleID, customerID, pointsEarned, 0, 0)
                if err != nil {
                        fmt.Printf("Error linking sale to customer: %v\n", err)
                        return
                }
                
                // Get updated customer for display
                customer, _ := db.GetCustomer(customerID)
                
                fmt.Printf("Sale ID %d linked to %s\n", saleID, customer.Name)
                fmt.Printf("Points earned: %d, New balance: %d\n", pointsEarned, customer.LoyaltyPoints)
                fmt.Printf("Current tier: %s\n", customer.LoyaltyTier)
        },
}

// displayCustomerDetails shows detailed information for a customer
func displayCustomerDetails(customer models.Customer) {
        fmt.Println("\nCustomer Details:")
        fmt.Printf("ID: %d\n", customer.ID)
        fmt.Printf("Name: %s\n", customer.Name)
        
        if customer.Email != "" {
                fmt.Printf("Email: %s\n", customer.Email)
        }
        
        if customer.Phone != "" {
                fmt.Printf("Phone: %s\n", customer.Phone)
        }
        
        if customer.Address != "" {
                fmt.Printf("Address: %s\n", customer.Address)
        }
        
        fmt.Printf("Join Date: %s\n", customer.JoinDate.Format("2006-01-02"))
        
        if !customer.LastPurchaseDate.IsZero() {
                fmt.Printf("Last Purchase: %s\n", customer.LastPurchaseDate.Format("2006-01-02"))
        }
        
        fmt.Printf("Total Spent: $%.2f\n", customer.TotalPurchases)
        fmt.Printf("Loyalty Points: %d\n", customer.LoyaltyPoints)
        fmt.Printf("Loyalty Tier: %s\n", customer.LoyaltyTier)
        
        if customer.Birthday != "" {
                fmt.Printf("Birthday: %s\n", customer.Birthday)
        }
        
        if customer.Notes != "" {
                fmt.Printf("Notes: %s\n", customer.Notes)
        }
        
        if customer.PreferredProducts != "" {
                fmt.Printf("Preferred Products: %s\n", customer.PreferredProducts)
        }
}

func init() {
        rootCmd.AddCommand(customerCmd)
        
        // Add subcommands
        customerCmd.AddCommand(customerAddCmd)
        customerCmd.AddCommand(customerGetCmd)
        customerCmd.AddCommand(customerUpdateCmd)
        customerCmd.AddCommand(customerDeleteCmd)
        customerCmd.AddCommand(customerListCmd)
        customerCmd.AddCommand(customerHistoryCmd)
        customerCmd.AddCommand(customerAddPointsCmd)
        customerCmd.AddCommand(loyaltyRewardsCmd)
        customerCmd.AddCommand(customerRedeemRewardCmd)
        customerCmd.AddCommand(customerLinkSaleCmd)
        customerCmd.AddCommand(customerLoyaltyStatusCmd)
        
        // Add flags for add command
        customerAddCmd.Flags().StringVar(&customerEmail, "email", "", "Customer email address")
        customerAddCmd.Flags().StringVar(&customerPhone, "phone", "", "Customer phone number")
        customerAddCmd.Flags().StringVar(&customerAddress, "address", "", "Customer address")
        customerAddCmd.Flags().StringVar(&customerNotes, "notes", "", "Additional notes about the customer")
        customerAddCmd.Flags().StringVar(&customerBirthday, "birthday", "", "Customer birthday (YYYY-MM-DD)")
        customerAddCmd.Flags().StringVar(&customerPreferredProd, "preferred-products", "", "Customer's preferred products")
        customerAddCmd.Flags().IntVar(&loyaltyPoints, "loyalty-points", 0, "Initial loyalty points")
        customerAddCmd.Flags().StringVar(&loyaltyTier, "loyalty-tier", "", "Initial loyalty tier")
        
        // Add flags for update command
        customerUpdateCmd.Flags().StringVar(&customerEmail, "email", "", "Customer email address")
        customerUpdateCmd.Flags().StringVar(&customerPhone, "phone", "", "Customer phone number")
        customerUpdateCmd.Flags().StringVar(&customerAddress, "address", "", "Customer address")
        customerUpdateCmd.Flags().StringVar(&customerNotes, "notes", "", "Additional notes about the customer")
        customerUpdateCmd.Flags().StringVar(&customerBirthday, "birthday", "", "Customer birthday (YYYY-MM-DD)")
        customerUpdateCmd.Flags().StringVar(&customerPreferredProd, "preferred-products", "", "Customer's preferred products")
        customerUpdateCmd.Flags().IntVar(&loyaltyPoints, "loyalty-points", 0, "Loyalty points")
        customerUpdateCmd.Flags().StringVar(&loyaltyTier, "loyalty-tier", "", "Loyalty tier")
}