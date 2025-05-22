package util

import "io/fs"

const (
	DefaultTheme             = "default"
	DirPerm      fs.FileMode = 0o777
	FilePerm     fs.FileMode = 0o666
)
