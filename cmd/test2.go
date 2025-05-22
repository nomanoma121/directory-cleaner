package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var test2Cmd = &cobra.Command{
	Use:   "test2",
	Short: "Test command to demonstrate functionality",
	Run: func(cmd *cobra.Command, args []string) {
		validate := func(input string) error {
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return errors.New("Invalid number")
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "Number",
			Validate: validate,
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("You choose %q\n", result)
	},
}

func init() {
	rootCmd.AddCommand(test2Cmd)
}
