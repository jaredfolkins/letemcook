package yeschef

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/jaredfolkins/letemcook/util"
)

func handleImagePull(cli *client.Client, imageRef string) error {
	normalizedName, _, _, err := util.NormalizeImageName(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image name '%s': %w", imageRef, err)
	}

	ctx := context.Background()
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		log.Printf("Warning: Failed to list images while checking for %s: %v", normalizedName, err)
	} else {
		for _, img := range images {
			for _, repoTag := range img.RepoTags {
				existingNormalized, _, _, parseErr := util.NormalizeImageName(repoTag)
				if parseErr == nil && existingNormalized == normalizedName {
					log.Printf("Image %s is already present locally (matched tag: %s). Using existing image.", normalizedName, repoTag)
					return nil // Image found locally
				}
			}
		}
		log.Printf("Image %s not found in local cache.", normalizedName)
	}

	pullRef := imageRef
	if !strings.Contains(pullRef, ":") {
		pullRef += ":latest"
	}
	log.Printf("Attempting to pull image reference: %s (normalized: %s)", pullRef, normalizedName)

	pullErr := pullImageV2(cli, pullRef, "")
	if pullErr == nil {
		log.Printf("Successfully pulled image: %s", pullRef)
		return nil // Success
	}

	log.Printf("Failed to pull image reference %s: %v", pullRef, pullErr)

	return fmt.Errorf("failed to pull image %s: %w", pullRef, pullErr)
}

func pullImageV2(cli *client.Client, imagename string, authConfig string) error {
	ctx := context.Background()
	options := image.PullOptions{}
	if authConfig != "" {
		options.RegistryAuth = authConfig
	}

	reader, err := cli.ImagePull(ctx, imagename, options)
	if err != nil {
		return err // Return error from ImagePull directly
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("error reading image pull output for %s: %w", imagename, err)
	}
	return nil // Success if copy completes without error
}

func createAuthConfig(username, password, registryname string, useToken bool) string {
	authConfig := registry.AuthConfig{
		Username:      username,
		ServerAddress: registryname,
	}

	if useToken {
		authConfig.IdentityToken = password // Use the password as a token
	} else {
		authConfig.Password = password // Use the password as a password
	}

	authBytes, _ := json.Marshal(authConfig)
	return base64.URLEncoding.EncodeToString(authBytes)
}
