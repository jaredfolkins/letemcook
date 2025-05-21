package yeschef

import (
	"os"

	"github.com/docker/docker/client"
)

// CheckJobImages verifies that all step images for the given job
// exist locally. It returns a slice of image names that are missing.
func CheckJobImages(job *JobRecipe) ([]string, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(os.Getenv("LEMC_DOCKER_HOST")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, st := range job.Recipe.Steps {
		if !imageExists(cli, st.Image) {
			missing = append(missing, st.Image)
		}
	}
	return missing, nil
}
