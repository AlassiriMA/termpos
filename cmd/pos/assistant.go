package main

import (
        "bufio"
        "fmt"
        "os"
        "strings"
        "time"

        "github.com/spf13/cobra"

        "termpos/internal/assistant"
        "termpos/internal/auth"
        "termpos/internal/db"
)

// initAssistantCommand sets up the AI assistant mode command
func initAssistantCommand() {
        var assistantCmd = &cobra.Command{
                Use:   "assistant",
                Short: "Start the AI assistant mode",
                Long:  `Start an interactive session that accepts natural language inputs with context-aware responses.`,
                RunE: func(cmd *cobra.Command, args []string) error {
                        fmt.Println("Starting AI Assistant Mode")
                        fmt.Println("-----------------------------------------------")
                        fmt.Println("Hello! I'm the TermPOS AI assistant.")
                        fmt.Println("You can ask me to:")
                        fmt.Println("• Add products: 'add coffee at $3.50'")
                        fmt.Println("• Sell products: 'sell 2 coffees'")
                        fmt.Println("• View inventory: 'show me the inventory'")
                        fmt.Println("• Generate reports: 'show sales report'")
                        fmt.Println("• And more - just ask naturally!")
                        fmt.Println("Type 'help' for a complete list of functions.")
                        fmt.Println("Type 'exit' or 'quit' to exit")
                        fmt.Println("-----------------------------------------------")
                        return startAssistantMode()
                },
        }

        rootCmd.AddCommand(assistantCmd)
}

func startAssistantMode() error {
        scanner := bufio.NewScanner(os.Stdin)
        lastCommandTime := time.Now()

        // Ask for login credentials if not already authenticated
        if !auth.IsAuthenticated() {
                fmt.Println("Please login to use the assistant mode:")
                
                var username, password string
                
                fmt.Print("Username: ")
                if !scanner.Scan() {
                        return fmt.Errorf("error reading username")
                }
                username = strings.TrimSpace(scanner.Text())
                
                fmt.Print("Password: ")
                if !scanner.Scan() {
                        return fmt.Errorf("error reading password")
                }
                password = strings.TrimSpace(scanner.Text())
                
                // Attempt login
                _, err := auth.Login(username, password, db.GetUserByUsername, db.UpdateLastLogin)
                if err != nil {
                        fmt.Printf("Login failed: %v\n", err)
                        return err
                }
                
                fmt.Println("Login successful!")
        }

        // Display welcome message for logged-in user
        user := auth.GetCurrentUser()
        fmt.Printf("Welcome, %s (%s)!\n", user.Username, user.Role)

        // Main input loop
        for {
                fmt.Print("\n> ")
                if !scanner.Scan() {
                        break
                }

                input := scanner.Text()
                input = strings.TrimSpace(input)

                if input == "exit" || input == "quit" {
                        fmt.Println("Exiting assistant mode. Have a great day!")
                        return nil
                }

                if input == "" {
                        continue
                }
                
                if input == "clear" || input == "reset" {
                        assistant.ClearContext()
                        fmt.Println("Conversation context has been reset.")
                        continue
                }
                
                if input == "logout" {
                        auth.Logout()
                        fmt.Println("You have been logged out.")
                        return nil
                }

                // Check for session timeout (context reset after 5 minutes of inactivity)
                now := time.Now()
                if now.Sub(lastCommandTime) > 5*time.Minute {
                        assistant.ClearContext()
                        fmt.Println("Session has been reset due to inactivity.")
                }
                lastCommandTime = now

                // Parse and execute the natural language command with context
                result, err := assistant.ProcessNaturalLanguageWithContext(input)
                if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        continue
                }

                fmt.Println(result)
        }

        if err := scanner.Err(); err != nil {
                return fmt.Errorf("input error: %w", err)
        }

        return nil
}
