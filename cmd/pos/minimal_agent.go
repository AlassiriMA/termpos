package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// Initialize minimal agent server command
func initMinimalAgentCommand() {
	var minAgentCmd = &cobra.Command{
		Use:   "agent-minimal",
		Short: "Start the minimal agent server",
		Long:  `Start a minimal HTTP server with a health check endpoint.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Starting minimal agent server on port %d...\n", port)
			// Run the server in a goroutine
			go startMinimalAgentServer(port)

			// Wait for termination signal
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			fmt.Println("Shutting down server...")
		},
	}

	rootCmd.AddCommand(minAgentCmd)
}

// Start minimal agent server
func startMinimalAgentServer(port int) {
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Start the server
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("Minimal server listening on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}