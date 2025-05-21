package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func GetLatestModTime(dir string) (time.Time, error) {
	var latest time.Time

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}

		return nil
	})

	return latest, err
}

func CopyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}

func ZipDir(sourceDir, zipFileName string) error {
	// ZIPファイル作成
	zipfile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// ZIP書き込み用ライター作成
	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	// ディレクトリ内のファイルを再帰的に追加
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// ディレクトリはZIPの中で空ディレクトリとして追加したい場合はここで処理できる（省略可）
			return nil
		}

		// ZIP内でのファイルパス（sourceDirの相対パスを使う）
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// ZIPに新規ファイルヘッダー作成
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// ファイルを開いて中身をコピー
		fsFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fsFile.Close()

		_, err = io.Copy(zipFileWriter, fsFile)
		return err
	})

	return err
}

// 指定されたディレクトリ内を探索し、すべてのnode_modulesディレクトリを削除する
func RemoveNodeModules(dir string) error {
	// ディレクトリを再帰的に探索し、node_modulesを見つけたら削除
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// エラーが発生しても処理を継続する（例: 権限がないファイルなど）
			// ただし、特定のクリティカルなエラーの場合はここでリターンすることも検討
			fmt.Fprintf(os.Stderr, "Error accessing path %q: %v\n", path, err)
			return nil // エラーを無視して継続
		}

		if info.IsDir() && info.Name() == "node_modules" {
			// node_modules ディレクトリを削除
			if err := os.RemoveAll(path); err != nil {
				// 削除中にエラーが発生した場合はそのエラーを返す
				fmt.Fprintf(os.Stderr, "Error removing %q: %v\n", path, err)
				return err
			}
			// 削除に成功した場合、このディレクトリ配下のさらなる探索をスキップ
			return filepath.SkipDir
		}
		return nil
	})
}

// GetDirSize は指定されたディレクトリの合計サイズをバイト単位で返します。
func GetDirSize(dirPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
