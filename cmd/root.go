package cmd

import (
	"fmt"
	"os"
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	logo string 
)

var rootCmd = &cobra.Command{
	Use:   "dclean",
	Short: "A simple and powerful directory archiving tool",
	Run: func(cmd *cobra.Command, args []string) {
		println()
		fmt.Println(logo)
		fmt.Println()
		fmt.Println("A simple and powerful directory archiving tool.\n")

		fmt.Println("Usage:")
		fmt.Println("  dclean [command]\n")

		fmt.Println("Available Commands:")
		fmt.Printf("  %-10s %s\n", "archive", "Archive old directories by date")
		fmt.Printf("  %-10s %s\n", "clean", "Remove files or directories based on rules")
		fmt.Printf("  %-10s %s\n", "help", "Show this help message")
		fmt.Println()

		fmt.Println("Flags:")
		fmt.Printf("  %-13s %s\n", "-h, --help", "Show help for dclean")
		fmt.Printf("  %-13s %s\n", "-v, --version", "Show version information")
		fmt.Println()

		fmt.Println("Run 'dclean [command] --help' for more information on a command.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
