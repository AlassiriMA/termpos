package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"termpos/internal/assistant"
)

// initAssistantCommand sets up the AI assistant mode command
func initAssistantCommand() {
	var assistantCmd = &cobra.Command{
		Use:   "assistant",
		Short: "Start the AI assistant mode",
		Long:  `Start an interactive session that accepts natural language inputs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting AI Assistant Mode")
			fmt.Println("Enter commands in natural language (e.g., 'add 3 coffees at $4.50')")
			fmt.Println("Type 'exit' or 'quit' to exit")
			return startAssistantMode()
		},
	}

	rootCmd.AddCommand(assistantCmd)
}

func startAssistantMode() error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		input = strings.TrimSpace(input)

		if input == "exit" || input == "quit" {
			fmt.Println("Exiting assistant mode")
			return nil
		}

		if input == "" {
			continue
		}

		// Parse and execute the natural language command
		result, err := assistant.ProcessNaturalLanguage(input)
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
