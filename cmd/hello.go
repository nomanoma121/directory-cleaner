package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Print hello, world",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello, world")
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
}
