package cmd

import (
	"directory-cleaner/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	archivePath       string
	daysForArchive    int
	useZip            bool
	removeNodeModules bool
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive old directories by copying or zipping them",
	Run: func(cmd *cobra.Command, args []string) {
		threshold := time.Now().AddDate(0, 0, -daysForArchive)
		archiveDir := "./archive"
		os.MkdirAll(archiveDir, 0755)

		entries, err := os.ReadDir(archivePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read directory %q: %v\n", archivePath, err)
			return
		}

		for _, entry := range entries {
			if entry.Name() == "archive" {
				continue
			}
			if !entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(archivePath, entry.Name())

			lastMod, err := utils.GetLatestModTime(fullPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get mod time for %q: %v\n", fullPath, err)
				continue
			}

			if lastMod.Before(threshold) {
				if removeNodeModules {
					err := utils.RemoveNodeModules(fullPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove node_modules in %s: %v\n", fullPath, err)
					} else {
						fmt.Printf("Removed node_modules in %s\n", fullPath)
					}
				}
				if useZip {
					zipName := filepath.Join(archiveDir, entry.Name()+".zip")
					fmt.Printf("Zipping: %s -> %s\n", fullPath, zipName)
					if err := utils.ZipDir(fullPath, zipName); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to zip %s: %v\n", fullPath, err)
					}
				} else {
					dest := filepath.Join(archiveDir, entry.Name())
					fmt.Printf("Copying: %s -> %s\n", fullPath, dest)
					if err := utils.CopyDir(fullPath, dest); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to copy %s: %v\n", fullPath, err)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
	archiveCmd.Flags().StringVarP(&archivePath, "path", "p", ".", "Directory to scan for archiving")
	archiveCmd.Flags().IntVarP(&daysForArchive, "days", "d", 30, "Days threshold for last modification")
	archiveCmd.Flags().BoolVarP(&useZip, "zip", "z", true, "Use zip archive instead of copy (default: true)")
	archiveCmd.Flags().BoolVar(&removeNodeModules, "remove-node-modules", false, "Remove node_modules directories before archiving")
}
