package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "dclean",
    Short: "Directory cleaner CLI",
    Run: func(cmd *cobra.Command, args []string) {
        // ルートコマンドで特に処理をしないなら空のままでもよい
        fmt.Println("Use a subcommand, e.g. 'dirc hello'")
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
