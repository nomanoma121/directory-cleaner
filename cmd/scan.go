package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"directory-cleaner/utils"

	"github.com/spf13/cobra"
)

var (
	scanPath string
	days     int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan top-level directories and list old ones",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Scanning path: %s\n", scanPath)
		fmt.Printf("Finding directories not updated in last %d days\n\n", days)

		// 日付のしきい値を計算
		threshold := time.Now().AddDate(0, 0, -days)

		// カレントディレクトリ直下のエントリ一覧
		entries, err := os.ReadDir(scanPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read directory %q: %v\n", scanPath, err)
			return
		}
		oldDirs := make([]string, 0)
		// カレント直下のディレクトリを一つずつ調査
		for i, entry := range entries {
			if entry.Name() == "archive" {
				continue
			}
			fmt.Printf("Checking %s... %d / %d \n", entry.Name(), i+1, len(entries))
			if entry.IsDir() {
				topLevelPath := filepath.Join(scanPath, entry.Name())

				// そのディレクトリの中を再帰的に調べて、最新の更新日を取得
				lastMod, err := utils.GetLatestModTime(topLevelPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", topLevelPath, err)
					continue
				}

				// 古いと判断されたら、そのトップレベルの名前だけ出力
				if lastMod.Before(threshold) {
					oldDirs = append(oldDirs, entry.Name())
				}
			}
		}
		for _, dir := range oldDirs {
			fmt.Printf("Old directory: %s\n", dir)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVarP(&scanPath, "path", "p", ".", "Directory path to scan")
	scanCmd.Flags().IntVarP(&days, "days", "d", 30, "Days threshold for last update")
}
