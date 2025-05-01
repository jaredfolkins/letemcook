package yeschef

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/jaredfolkins/letemcook/util"
)

func handleImagePull(cli *client.Client, imageRef string) error {
	pullRef, baseImageName, tag, err := util.NormalizeImageName(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image name '%s': %w", imageRef, err)
	}
	isLatestTag := tag == "latest"

	log.Printf("Checking image: %s (Resolved Ref: %s, Base Name: %s, Tag: %s, IsLatest: %t)", imageRef, pullRef, baseImageName, tag, isLatestTag)

	ctx := context.Background()
	localImageDigest := ""
	imageFoundLocally := false

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		log.Printf("Warning: Failed to list images while checking for %s: %v", pullRef, err)
	} else {
		for _, img := range images {
			for _, repoTag := range img.RepoTags {
				if repoTag == pullRef {
					imageFoundLocally = true
					localImageDigest = img.ID
					log.Printf("Exact match found locally: %s with ID %s", repoTag, localImageDigest)
					break
				}
			}
			if imageFoundLocally {
				break
			}
		}
	}

	if isLatestTag && imageFoundLocally {
		log.Printf("Local image '%s' found. Checking remote registry for updates...", pullRef)
		remoteDigest, err := getRemoteImageDigest(cli, pullRef)
		if err != nil {
			log.Printf("Warning: Failed to check remote registry for '%s': %v. Will attempt pull.", pullRef, err)
		} else {
			localDigestClean := strings.TrimPrefix(localImageDigest, "sha256:")
			log.Printf("Comparing local digest (%s) with remote digest (%s)", localDigestClean, remoteDigest)
			if remoteDigest != "" && strings.HasPrefix(localDigestClean, remoteDigest) {
				log.Printf("Local image '%s' is up-to-date.", pullRef)
				return nil
			}
			log.Printf("Remote registry has a newer version for '%s'. Pulling update.", pullRef)
		}
	} else if imageFoundLocally {
		log.Printf("Image '%s' found locally. Using existing image.", pullRef)
		return nil
	} else {
		log.Printf("Image '%s' not found in local cache.", pullRef)
	}

	log.Printf("Attempting to pull image reference: %s", pullRef)
	pullErr := pullImageV2(cli, pullRef, "")
	if pullErr == nil {
		log.Printf("Successfully pulled image: %s", pullRef)
		return nil
	}

	log.Printf("Failed to pull image reference %s: %v", pullRef, pullErr)
	return fmt.Errorf("failed to pull image %s: %w", pullRef, pullErr)
}

func getRemoteImageDigest(cli *client.Client, imageRef string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	distInspect, err := cli.DistributionInspect(ctx, imageRef, "")
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Printf("Image '%s' not found in remote registry.", imageRef)
			return "", nil
		}
		return "", fmt.Errorf("failed to inspect remote image '%s': %w", imageRef, err)
	}

	return string(distInspect.Descriptor.Digest), nil
}

func pullImageV2(cli *client.Client, imagename string, authConfig string) error {
	ctx := context.Background()
	options := image.PullOptions{}
	if authConfig != "" {
		options.RegistryAuth = authConfig
	}

	reader, err := cli.ImagePull(ctx, imagename, options)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("error reading image pull output for %s: %w", imagename, err)
	}
	return nil
}

func createAuthConfig(username, password, registryname string, useToken bool) string {
	authConfig := registry.AuthConfig{
		Username:      username,
		ServerAddress: registryname,
	}

	if useToken {
		authConfig.IdentityToken = password
	} else {
		authConfig.Password = password
	}

	authBytes, _ := json.Marshal(authConfig)
	return base64.URLEncoding.EncodeToString(authBytes)
}
