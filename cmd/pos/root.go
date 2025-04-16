package main

import (
        "fmt"
        "os"
        "path/filepath"

        "github.com/spf13/cobra"
        "gopkg.in/yaml.v2"

        "termpos/internal/auth"
        "termpos/internal/db"
)

var (
        cfgFile    string
        dbPath     string
        mode       string
        port       int
        showBanner bool
        debug      bool

        rootCmd = &cobra.Command{
                Use:   "pos",
                Short: "Terminal-based Point of Sale system",
                Long: `A terminal-based Point of Sale (POS) system with three operating modes:
1. Classic CLI Mode — Command-based interface
2. Agent Mode — HTTP server for remote commands
3. AI Assistant Mode — Natural language interface`,
                PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
                        // Show welcome banner if enabled
                        if showBanner && cmd.Use != "version" && cmd.Use != "help" {
                                PrintWelcomeBanner()
                        }

                        // Check for first-run scenario
                        configFile := filepath.Join("./config", "config.json")
                        if _, err := os.Stat(configFile); os.IsNotExist(err) {
                                // Config file doesn't exist and it's not the init command
                                if cmd.Use != "init" && cmd.Use != "version" && cmd.Use != "help" {
                                        fmt.Println(cyan("Welcome to TermPOS!"))
                                        fmt.Println(cyan("It looks like this is your first time running the application."))
                                        fmt.Println(yellow("Run './termpos init' to set up your system with the interactive configuration wizard."))
                                        fmt.Println()
                                }
                        }
                        
                        // Check if database directories exist
                        ensureDirectoriesExist()
                        
                        // Initialize database
                        if err := db.Initialize(dbPath); err != nil {
                                return fmt.Errorf("failed to initialize database: %w", err)
                        }
                        
                        // Try to load session if it exists
                        if err := auth.LoadSession(); err != nil {
                                fmt.Printf("Warning: Failed to load session: %v\n", err)
                                // Non-fatal error, continue without a session
                        }
                        
                        return nil
                },
        }
)

// Execute adds all child commands to the root command and sets flags appropriately.
func init() {
        cobra.OnInitialize(initConfig)

        // Define flags for the root command
        rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.termpos.yaml)")
        rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./pos.db", "database path")
        rootCmd.PersistentFlags().StringVar(&mode, "mode", "classic", "operating mode (classic, agent, assistant)")
        rootCmd.PersistentFlags().IntVar(&port, "port", 8000, "port for agent mode server")
        rootCmd.PersistentFlags().BoolVar(&showBanner, "banner", true, "show welcome banner")
        rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

        // Initialize the CLI commands
        initClassicCommands()
        initAgentCommand()
        initMinimalAgentCommand()  // Add the minimal agent command
        initAssistantCommand()
        initUserCommands()
        initStaffCommands()
}

// ensureDirectoriesExist makes sure required directories are available
func ensureDirectoriesExist() {
        // Define required directories
        directories := []string{
                "./data",
                "./config",
                "./backups",
        }
        
        // Create directories if they don't exist
        for _, dir := range directories {
                if _, err := os.Stat(dir); os.IsNotExist(err) {
                        if debug {
                                fmt.Printf("Creating directory: %s\n", dir)
                        }
                        os.MkdirAll(dir, 0755)
                }
        }
}

// initConfig reads in config file if set
func initConfig() {
        if cfgFile != "" {
                // Use config file from the flag
                return
        }

        // Find home directory
        home, err := os.UserHomeDir()
        if err != nil {
                fmt.Println(err)
                return
        }

        // Search for config in home directory
        cfgFile = filepath.Join(home, ".termpos.yaml")
        if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
                // Create default config if it doesn't exist
                createDefaultConfig(cfgFile)
        }

        // Read the config file
        data, err := os.ReadFile(cfgFile)
        if err != nil {
                fmt.Printf("Warning: Cannot read config file: %v\n", err)
                return
        }

        var config struct {
                DBPath string `yaml:"db_path"`
                Mode   string `yaml:"mode"`
                Port   int    `yaml:"port"`
        }

        if err := yaml.Unmarshal(data, &config); err != nil {
                fmt.Printf("Warning: Cannot parse config file: %v\n", err)
                return
        }

        // Only set values if they weren't explicitly provided via flags
        if !rootCmd.PersistentFlags().Changed("db") && config.DBPath != "" {
                dbPath = config.DBPath
        }
        if !rootCmd.PersistentFlags().Changed("mode") && config.Mode != "" {
                mode = config.Mode
        }
        if !rootCmd.PersistentFlags().Changed("port") && config.Port != 0 {
                port = config.Port
        }
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) {
        config := map[string]interface{}{
                "db_path": "./pos.db",
                "mode":    "classic",
                "port":    8000,
        }

        data, err := yaml.Marshal(config)
        if err != nil {
                fmt.Printf("Warning: Cannot create default config: %v\n", err)
                return
        }

        if err := os.WriteFile(path, data, 0644); err != nil {
                fmt.Printf("Warning: Cannot write default config: %v\n", err)
        }
}
