package main

import (
        "fmt"
        "runtime"
        "time"

        "github.com/spf13/cobra"
)

// ANSI color codes
const (
        ColorReset  = "\033[0m"
        ColorRed    = "\033[31m"
        ColorGreen  = "\033[32m"
        ColorYellow = "\033[33m"
        ColorBlue   = "\033[34m"
        ColorPurple = "\033[35m"
        ColorCyan   = "\033[36m"
        ColorWhite  = "\033[37m"
)

// Build information - to be set at compile time
var (
        Version     = "0.1.0"
        BuildDate   = ""
        GitCommit   = ""
        BuildOS     = runtime.GOOS
        BuildArch   = runtime.GOARCH
        GoVersion   = runtime.Version()
)

// Returns formatted build information
func getBuildInfo() string {
        if BuildDate == "" {
                // If build date wasn't set at compile time, use current time
                BuildDate = time.Now().Format("2006-01-02")
        }
        
        if GitCommit == "" {
                GitCommit = "development"
        }
        
        return fmt.Sprintf("Version:      %s\nBuild Date:   %s\nGit Commit:   %s\nGo Version:   %s\nOS/Arch:      %s/%s",
                Version, BuildDate, GitCommit, GoVersion, BuildOS, BuildArch)
}

// Returns ASCII art banner for the application
func getBanner() string {
        return `
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║   ████████╗███████╗██████╗ ███╗   ███╗██████╗  ██████╗ ███████╗║
║   ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║██╔══██╗██╔═══██╗██╔════╝║
║      ██║   █████╗  ██████╔╝██╔████╔██║██████╔╝██║   ██║███████╗║
║      ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║██╔═══╝ ██║   ██║╚════██║║
║      ██║   ███████╗██║  ██║██║ ╚═╝ ██║██║     ╚██████╔╝███████║║
║      ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝      ╚═════╝ ╚══════╝║
║                                                                ║
║                Terminal Point of Sale System                   ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`
}

// Colorize functions
func blue(s string) string {
        return ColorBlue + s + ColorReset
}

func green(s string) string {
        return ColorGreen + s + ColorReset
}

func yellow(s string) string {
        return ColorYellow + s + ColorReset
}

func cyan(s string) string {
        return ColorCyan + s + ColorReset
}

// PrintWelcomeBanner prints the welcome banner with version information
func PrintWelcomeBanner() {
        fmt.Println(blue(getBanner()))
        fmt.Printf("%s %s\n", green("Version:"), Version)
        fmt.Printf("%s %s\n\n", green("Build Date:"), BuildDate)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
        Use:   "version",
        Short: "Display version information",
        Long:  `Display detailed version and build information about the application`,
        Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("\nTermPOS - Terminal Point of Sale System")
                fmt.Println(getBuildInfo())
        },
}

func init() {
        rootCmd.AddCommand(versionCmd)
}