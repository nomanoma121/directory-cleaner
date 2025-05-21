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
	response          string
	removeNodeModules bool
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive old directories by copying or zipping them",
	Run: func(cmd *cobra.Command, args []string) {
		threshold := time.Now().AddDate(0, 0, -daysForArchive)
		archiveDir := filepath.Join(archivePath, "archive")
		// tmpディレクトリと競合しないように隠しファイルにする
		tmpDir := "./archive/.tmp"
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
			if !entry.IsDir() {
				continue
			}
			targetDirs = append(targetDirs, entry)
		}

		// 古いディレクトリを順番にアーカイブ
		for _, target := range targetDirs {
			fullPath := filepath.Join(archivePath, target.Name())
			fmt.Println("Checking:", fullPath)
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
					if err := utils.RemoveNodeModules(tmpDir); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove node_modules from %q: %v\n", tmpDir, err)
						continue
					}
					fmt.Println("Removed node_modules from:", tmpDir)
				}

				// アーカイブ前のディレクトリサイズを取得
				originalSize, err := utils.GetDirSize(tmpDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get original directory size %q: %v\n", tmpDir, err)
					// エラーが発生しても処理を続行する
				}

				// アーカイブを作成
				if useZip {
					zipPath := filepath.Join(archiveDir, target.Name()+".zip")
					err = utils.ZipDir(tmpDir, zipPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to create zip %q: %v\n", zipPath, err)
						continue
					}
					fmt.Println("Created zip:", zipPath)

					// zipファイルサイズを取得
					zipInfo, err := os.Stat(zipPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to get zip file size %q: %v\n", zipPath, err)
					} else {
						fmt.Printf("Original size: %s, Zipped size: %s, Saved: %s (%.2f%%)\n",
							formatBytes(originalSize),
							formatBytes(zipInfo.Size()),
							formatBytes(originalSize-zipInfo.Size()),
							float64(originalSize-zipInfo.Size())/float64(originalSize)*100)
					}
				} else {
					copyPath := filepath.Join(archiveDir, target.Name())
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
			}
		}

		fmt.Println("Archiving completed.")
		fmt.Print("Do you want to delete the original directories? [Y/n]: ")
		fmt.Scanln(&response)
		if response == "y" || response == "Y" {
			for _, target := range targetDirs {
				fullPath := filepath.Join(archivePath, target.Name())
				err = os.RemoveAll(fullPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to remove original directory %q: %v\n", fullPath, err)
					continue
				}
				fmt.Println("Removed original directory:", fullPath)
			}
		} else {
			fmt.Println("Original directories not removed.")
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

// formatBytes はバイト数を人間が読みやすい形式（KB, MB, GB）に変換します。
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	dim := "KMGTPE"
	div := int64(unit)
	for i := 0; i < len(dim); i++ {
		if bytes < div*unit {
			return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), dim[i])
		}
		div *= unit
	}
	return fmt.Sprintf("%.2f EB", float64(bytes)/float64(div))
}
