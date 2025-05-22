package main

import (
	"directory-cleaner/cmd"
	_ "embed" // embedパッケージをブランクインポート
)

//go:embed assets/logo.txt
var logoContent []byte

func main() {
	cmd.SetLogo(logoContent)
	cmd.Execute()
}
