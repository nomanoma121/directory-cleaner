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
	archiveAll        bool
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive old directories by copying or zipping them",
	Run: func(cmd *cobra.Command, args []string) {
		threshold := time.Now().AddDate(0, 0, -daysForArchive)

		var totalOriginalSize int64
		var totalArchivedSize int64

		entries, err := os.ReadDir(archivePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read directory %q: %v\n", archivePath, err)
			return
		}

		// カレントディレクトリ直下のエントリを一つずつ調査
		fmt.Printf("Scanning path: %s\n", archivePath)
		targetDirs := make([]os.DirEntry, 0)
		for _, entry := range entries {
			if entry.Name() == "archive" {
				continue
			}
			if entry.IsDir() {
				fullPath := filepath.Join(archivePath, entry.Name())
				// そのディレクトリの中を再帰的に調べて、最新の更新日を取得
				lastMod, err := utils.GetLatestModTime(fullPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", fullPath, err)
					continue
				}
				// 古いと判断されたら、そのトップレベルの名前だけ出力
				if lastMod.Before(threshold) || archiveAll {
					targetDirs = append(targetDirs, entry)
				} else {
					fmt.Printf("Skipping %s (last modified: %s)\n", entry.Name(), lastMod.Format(time.RFC3339))
				}
			}
		}

		// アーカイブ対象のディレクトリを表示
		fmt.Printf("%d directories found for archiving:\n", len(targetDirs))
		for _, target := range targetDirs {
			fmt.Println(" -", target.Name())
		}
		if len(targetDirs) == 0 {
			fmt.Println("No directories to archive.")
			return
		}

		// アーカイブを作成するか確認
		fmt.Println()
		fmt.Print("Do you want to proceed with archiving? [Y/n]: ")
		fmt.Scanln(&response)
		fmt.Println()
		if response != "y" && response != "Y" {
			fmt.Println("Aborting archiving.")
			return
		}

		// アーカイブディレクトリを作成
		archiveDir := filepath.Join(archivePath, "archive")
		// tmpディレクトリと競合しないように隠しファイルにする
		tmpDir := "./archive/.tmp"
		os.MkdirAll(tmpDir, 0755)
		// この関数が終了する前にtmpDirを確実に削除する
		defer os.RemoveAll(tmpDir)

		successDirs := make([]os.DirEntry, 0)
		// 古いディレクトリを順番にアーカイブ
		for _, target := range targetDirs {
			fmt.Printf("Archiving %s...\n", target.Name())
			fullPath := filepath.Join(archivePath, target.Name())
			// tmpディレクトリに一時的にコピー
			err = utils.CopyDir(fullPath, tmpDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy %q to %q: %v\n", fullPath, tmpDir, err)
				continue
			}

			// アーカイブ前のディレクトリサイズを取得
			originalSize, err := utils.GetDirSize(tmpDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get original directory size %q: %v\n", tmpDir, err)
				// エラーが発生しても処理を続行する originalSize は 0 のままになる
			}

			if removeNodeModules {
				if err := utils.RemoveNodeModules(tmpDir); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to remove node_modules from %q: %v\n", tmpDir, err)
					continue
				}
			}

			// アーカイブを作成
			if useZip {
				zipPath := filepath.Join(archiveDir, target.Name()+".zip")
				err = utils.ZipDir(tmpDir, zipPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to create zip %q: %v\n", zipPath, err)
					continue
				}
				fmt.Println("✅ Created zip:", zipPath)

				// zipファイルサイズを取得
				zipInfo, err := os.Stat(zipPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get zip file size %q: %v\n", zipPath, err)
				} else {
					fmt.Printf("Original: %s, Zipped: %s, Saved: %s (%.2f%%)\n",
						formatBytes(originalSize),
						formatBytes(zipInfo.Size()),
						formatBytes(originalSize-zipInfo.Size()),
						float64(originalSize-zipInfo.Size())/float64(originalSize)*100)
				}
				totalOriginalSize += originalSize
				totalArchivedSize += zipInfo.Size()
			} else {
				copyPath := filepath.Join(archiveDir, target.Name())
				err = utils.CopyDir(tmpDir, copyPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to copy %q to %q: %v\n", tmpDir, copyPath, err)
					continue
				}
				fmt.Println("✅ Copied to:", copyPath)

				// コピー後のディレクトリサイズを取得
				archivedDirSize, err := utils.GetDirSize(copyPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get archived directory size %q: %v\n", copyPath, err)
				} else {
					// originalSize が 0 の場合は除算エラーを避ける
					if originalSize > 0 {
						fmt.Printf("Original: %s, Copied: %s, Saved: %s (%.2f%%)\n",
							formatBytes(originalSize),
							formatBytes(archivedDirSize),
							formatBytes(originalSize-archivedDirSize),
							float64(originalSize-archivedDirSize)/float64(originalSize)*100)
					} else {
						fmt.Printf("Original: %s, Copied: %s, Saved: %s\n",
							formatBytes(originalSize),
							formatBytes(archivedDirSize),
							formatBytes(originalSize-archivedDirSize))
					}
				}
				totalOriginalSize += originalSize
				totalArchivedSize += archivedDirSize
			}

			// 各ターゲットの処理が終わった後にtmpDirをクリーンアップする
			// ただし、Renameで移動した場合はtmpDirは既に存在しないのでエラーになる場合がある
			// そのため、エラーをチェックせずに単純に削除を試みる
			os.RemoveAll(tmpDir)      // tmpDirを次のループのために空にする
			os.MkdirAll(tmpDir, 0755) // 再度tmpDirを作成
			successDirs = append(successDirs, target)
		}

		// アーカイブの合計サイズを表示
		fmt.Println()
		if totalOriginalSize > 0 {
			fmt.Printf("Total original: %s\n", formatBytes(totalOriginalSize))
			fmt.Printf("Total archived: %s\n", formatBytes(totalArchivedSize))
			fmt.Printf("Total saved: %s (%.2f%%)\n",
				formatBytes(totalOriginalSize-totalArchivedSize),
				float64(totalOriginalSize-totalArchivedSize)/float64(totalOriginalSize)*100)
		} else {
			fmt.Println("No directories were archived.")
		}

		fmt.Println()
		fmt.Println("Archiving completed!!")

		// アーカイブに成功したディレクトリを表示
		fmt.Println()
		fmt.Println("Successfully archived directories:")
		for _, successDir := range successDirs {
			fmt.Println(" -", successDir.Name())
		}

		// 元のディレクトリを削除するか確認
		fmt.Println()
		fmt.Print("Delete original dirs? (failures will be skipped) [Y/n]: ")
		fmt.Scanln(&response)
		fmt.Println()
		if response == "y" || response == "Y" {
			for _, successDir := range successDirs {
				fullPath := filepath.Join(archivePath, successDir.Name())
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

		fmt.Println()
		fmt.Println("All done ✨")
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
	archiveCmd.Flags().StringVarP(&archivePath, "path", "p", ".", "Directory to scan for archiving")
	archiveCmd.Flags().IntVarP(&daysForArchive, "days", "d", 30, "Days threshold for last modification")
	archiveCmd.Flags().BoolVarP(&useZip, "no-zip", "n", true, "Use zip archive instead of copy (default: true)")
	archiveCmd.Flags().BoolVarP(&removeNodeModules, "remove-node-modules", "r", false, "Remove node_modules directories before archiving")
	archiveCmd.Flags().BoolVarP(&archiveAll, "all", "a", false, "Archive all directories without checking last modification date")
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
