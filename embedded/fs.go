package embedded

import (
	"embed"
	"io/fs"
)

//go:embed assets/themes/default/public/*
//go:embed assets/themes/banilla/public/*
//go:embed assets/heckle/public/*
//go:embed migrations/*.sql
//go:embed seeds/*.sql
var embedAssets embed.FS

const AssetsRoot = "assets"
const MigrationsRoot = "migrations"
const SeedRoot = "seeds"

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

// GetSeedFS returns a sub-filesystem rooted at the SeedRoot directory
func GetSeedFS() (fs.FS, error) {
	return fs.Sub(embedAssets, SeedRoot)
}

// ReadAsset reads a file from the embedded assets filesystem relative to the AssetsRoot.
func ReadAsset(name string) ([]byte, error) {
	assetsFS, err := GetAssetsFS()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(assetsFS, name)
}
