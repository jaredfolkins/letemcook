package embedded

import (
	"embed"
	"io/fs"
)

//go:embed assets/themes/default/public/*
//go:embed assets/themes/banilla/public/*
//go:embed assets/heckle/public/*
//go:embed migrations/*.sql
var embedAssets embed.FS

const AssetsRoot = "assets"
const MigrationsRoot = "migrations"

// GetAssetsFS returns a sub-filesystem rooted at the AssetsRoot directory
// within the embedded assets.
func GetAssetsFS() (fs.FS, error) {
	return fs.Sub(embedAssets, AssetsRoot)
}

// GetMigrationsFS returns a sub-filesystem rooted at the MigrationsRoot directory
// within the embedded assets.
func GetMigrationsFS() (fs.FS, error) {
	return fs.Sub(embedAssets, MigrationsRoot)
}

// ReadAsset reads a file from the embedded assets filesystem relative to the AssetsRoot.
func ReadAsset(name string) ([]byte, error) {
	assetsFS, err := GetAssetsFS()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(assetsFS, name)
}
