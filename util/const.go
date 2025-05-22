package util

import "io/fs"

const (
	DefaultTheme             = "default"
	DirPerm      fs.FileMode = 0o755
	FilePerm     fs.FileMode = 0o666
)
