package util

import (
	"fmt"
	"strings"
)

func NormalizeImageName(imageRef string) (normalizedName string, imageName string, tag string, err error) {
	imageName = imageRef
	tag = "latest" // Default tag

	if strings.Contains(imageName, ":") {
		parts := strings.SplitN(imageName, ":", 2)
		imageName = parts[0]
		tag = parts[1]
	}

	parts := strings.SplitN(imageName, "/", 2)
	hasDomain := false
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		hasDomain = true
	}

	if !strings.Contains(imageName, "/") {
		normalizedName = fmt.Sprintf("docker.io/library/%s:%s", imageName, tag)
	} else if !hasDomain {
		normalizedName = fmt.Sprintf("docker.io/%s:%s", imageName, tag)
	} else {
		normalizedName = fmt.Sprintf("%s:%s", imageName, tag)
	}

	imageNameOnly := imageName
	if strings.Contains(imageName, "/") {
		parts := strings.Split(imageName, "/")
		imageNameOnly = parts[len(parts)-1]
	}

	return normalizedName, imageNameOnly, tag, nil
}
