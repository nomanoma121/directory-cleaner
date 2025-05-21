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
		tmpDir :=  "./archive/.tmp"
		os.MkdirAll(tmpDir, 0755)

		entries, err := os.ReadDir(archivePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read directory %q: %v\n", archivePath, err)
			return
		}

		// スキャン対象のディレクトリを取得(archiveディレクトリは除外
		targetDirs := make([]os.DirEntry, 0)
		for _, entry := range entries {
			if entry.Name() == "archive" {
				continue
			}
			targetDirs = append(targetDirs, entry)
		}

		// 古いディレクトリを順番にアーカイブ
		for _, target := range targetDirs {
			fullPath := filepath.Join(archivePath, target.Name())

			lastMod, err := utils.GetLatestModTime(fullPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get mod time for %q: %v\n", fullPath, err)
				continue
			}

			if lastMod.Before(threshold) {
				fmt.Println("Archiving:", target.Name())
				// tmpディレクトリに一時的にコピー
				err = utils.CopyDir(fullPath, tmpDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to copy %q to %q: %v\n", fullPath, tmpDir, err)
					continue
				}

				if removeNodeModules {
					nodeModulesPath := filepath.Join(tmpDir, "node_modules")
					if _, err := os.Stat(nodeModulesPath); !os.IsNotExist(err) {
						err = os.RemoveAll(nodeModulesPath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to remove %q: %v\n", nodeModulesPath, err)
						}
					}
				}

				// アーカイブを作成
				if useZip {
					zipPath := filepath.Join(archivePath, target.Name()+".zip")
					err = utils.ZipDir(tmpDir, zipPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to create zip %q: %v\n", zipPath, err)
						continue
					}
					fmt.Println("Created zip:", zipPath)
				} else {
					copyPath := filepath.Join(archivePath, target.Name())
					err = os.Rename(tmpDir, copyPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to move %q to %q: %v\n", tmpDir, copyPath, err)
						continue
					}
					fmt.Println("Copied to:", copyPath)
				}
				
				// tmpディレクトリを削除
				err = os.RemoveAll(tmpDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to remove tmp dir %q: %v\n", tmpDir, err)
					continue
				}
				fmt.Println("Removed tmp dir:", tmpDir)
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
