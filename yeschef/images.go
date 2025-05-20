package yeschef

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"         // Import AWS core
	"github.com/aws/aws-sdk-go-v2/config"      // Import AWS config loader
	"github.com/aws/aws-sdk-go-v2/credentials" // Import static credentials provider
	"github.com/aws/aws-sdk-go-v2/service/ecr" // Import ECR service client

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/jaredfolkins/letemcook/util"

	dockerTypes "github.com/docker/docker/api/types" // Alias to avoid conflict
)

// ImageSpec defines the structure for image specifications in the recipe YAML.
type ImageSpec struct {
	Name string `yaml:"image"`
	// RegistryAuth specifies authentication credentials. Prefixes determine the type:
	// - "aws:ACCESS_KEY:SECRET_KEY" or "aws:b64(ACCESS_KEY:SECRET_KEY)" for AWS ECR.
	// - "gcp:_json_key_base64:B64_KEY_JSON" for GCP GCR/Artifact Registry.
	// - "azr:APP_ID:PASSWORD" for Azure ACR (Service Principal).
	// - "basic:USER:PASSWORD" or "basic:b64(USER:PASSWORD)" for basic auth.
	// - If no prefix, defaults to basic auth (USER:PASSWORD or b64(USER:PASSWORD)).
	RegistryAuth string `yaml:"registry_auth,omitempty"`
}

// getECRAuthToken fetches an ECR authorization token using provided AWS credentials.
func getECRAuthToken(ctx context.Context, accessKeyID, secretAccessKey, region string) (*registry.AuthConfig, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region), // Use the region derived from the ECR URL or a default
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)

	log.Println("Requesting ECR authorization token...")
	tokenOutput, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR authorization token: %w", err)
	}

	if len(tokenOutput.AuthorizationData) == 0 {
		return nil, fmt.Errorf("no ECR authorization data received")
	}

	authData := tokenOutput.AuthorizationData[0]
	if authData.AuthorizationToken == nil || authData.ProxyEndpoint == nil {
		return nil, fmt.Errorf("invalid ECR authorization data received")
	}

	tokenBytes, err := base64.StdEncoding.DecodeString(aws.ToString(authData.AuthorizationToken))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ECR authorization token: %w", err)
	}

	tokenParts := strings.SplitN(string(tokenBytes), ":", 2)
	if len(tokenParts) != 2 {
		return nil, fmt.Errorf("invalid ECR token format")
	}

	log.Printf("Successfully obtained ECR token for endpoint: %s", aws.ToString(authData.ProxyEndpoint))

	return &registry.AuthConfig{
		Username:      tokenParts[0], // Should be "AWS"
		Password:      tokenParts[1],
		ServerAddress: aws.ToString(authData.ProxyEndpoint),
	}, nil
}

// buildAuthHeader converts the RegistryAuth string into the base64 encoded Docker auth header.
func buildAuthHeader(spec ImageSpec) (string, error) {
	val := strings.TrimSpace(spec.RegistryAuth)
	if val == "" {
		return "", nil // No explicit auth provided, rely on local Docker config.
	}

	var authConfig registry.AuthConfig
	var err error

	switch {
	case strings.HasPrefix(val, "aws:"):
		log.Println("Processing AWS ECR credentials...")
		credsPart := strings.TrimPrefix(val, "aws:")
		if decoded, decodeErr := base64.StdEncoding.DecodeString(credsPart); decodeErr == nil && strings.Contains(string(decoded), ":") {
			credsPart = string(decoded)
		}
		credParts := strings.SplitN(credsPart, ":", 2)
		if len(credParts) != 2 {
			return "", fmt.Errorf("invalid AWS credentials format in registry_auth, expected aws:AccessKeyID:SecretAccessKey or aws:b64(AccessKeyID:SecretAccessKey)")
		}
		accessKeyID := credParts[0]
		secretAccessKey := credParts[1]

		_, baseImageName, _, parseErr := util.NormalizeImageName(spec.Name)
		if parseErr != nil {
			return "", fmt.Errorf("cannot determine AWS region: failed to parse image name '%s': %w", spec.Name, parseErr)
		}
		region := deriveAWSRegion(baseImageName)
		if region == "" {
			log.Printf("Warning: Could not derive AWS region from image name '%s'. Using default AWS region resolution.", spec.Name)
		}

		awsAuthConfig, ecrErr := getECRAuthToken(context.Background(), accessKeyID, secretAccessKey, region)
		if ecrErr != nil {
			return "", fmt.Errorf("failed to get ECR auth token: %w", ecrErr)
		}
		authConfig = *awsAuthConfig // Use the AuthConfig returned by getECRAuthToken

	case strings.HasPrefix(val, "gcp:"):
		log.Println("Processing GCP GCR/Artifact Registry credentials...")
		credsPart := strings.TrimPrefix(val, "gcp:")
		// Expecting format like "_json_key_base64:BASE64_KEY_JSON"
		parts := strings.SplitN(credsPart, ":", 2)
		if len(parts) != 2 || !(parts[0] == "_json_key_base64" || parts[0] == "_json_key") {
			return "", fmt.Errorf("invalid GCP credentials format in registry_auth, expected gcp:_json_key_base64:B64_KEY_JSON or gcp:_json_key:B64_KEY_JSON")
		}
		authConfig = registry.AuthConfig{
			Username: parts[0],
			Password: parts[1],
			// ServerAddress is usually derived from the image name for GCP.
		}

	case strings.HasPrefix(val, "azr:"):
		log.Println("Processing Azure ACR credentials...")
		credsPart := strings.TrimPrefix(val, "azr:")
		// Expecting format like "APP_ID:PASSWORD"
		parts := strings.SplitN(credsPart, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid Azure credentials format in registry_auth, expected azr:APP_ID:PASSWORD")
		}
		authConfig = registry.AuthConfig{
			Username: parts[0],
			Password: parts[1],
			// ServerAddress is usually derived from the image name for Azure.
		}

	case strings.HasPrefix(val, "basic:"):
		log.Println("Processing basic credentials...")
		authStr := strings.TrimPrefix(val, "basic:")
		if decoded, decodeErr := base64.StdEncoding.DecodeString(authStr); decodeErr == nil && strings.Contains(string(decoded), ":") {
			authStr = string(decoded)
		}
		parts := strings.SplitN(authStr, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid basic auth format in registry_auth, expected basic:USER:PASSWORD or basic:b64(USER:PASSWORD)")
		}
		authConfig = registry.AuthConfig{
			Username: parts[0],
			Password: parts[1],
		}

	default: // Treat as basic auth if no recognized prefix
		log.Println("No recognized prefix found, processing as basic credentials...")
		authStr := val
		if decoded, decodeErr := base64.StdEncoding.DecodeString(authStr); decodeErr == nil && strings.Contains(string(decoded), ":") {
			authStr = string(decoded)
		}
		parts := strings.SplitN(authStr, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("unprefixed registry_auth format not recognized, expected USER:PASSWORD or b64(USER:PASSWORD)")
		}
		authConfig = registry.AuthConfig{
			Username: parts[0],
			Password: parts[1],
		}
	}

	// Marshal the final AuthConfig
	blob, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth config: %w", err)
	}
	return base64.URLEncoding.EncodeToString(blob), nil
}

// deriveAWSRegion attempts to extract the AWS region from an ECR image name.
func deriveAWSRegion(baseImageName string) string {
	// Example: 123456789012.dkr.ecr.us-west-2.amazonaws.com/my-repo
	if strings.Contains(baseImageName, ".dkr.ecr.") && strings.Contains(baseImageName, ".amazonaws.com") {
		parts := strings.Split(baseImageName, ".")
		// Example parts: [123456789012, dkr, ecr, us-west-2, amazonaws, com]
		if len(parts) >= 5 && parts[1] == "dkr" && parts[2] == "ecr" {
			log.Printf("Derived AWS region '%s' from image name '%s'", parts[3], baseImageName)
			return parts[3]
		}
	}
	return "" // Region could not be derived
}

// handleImagePull checks if an image exists locally and pulls if necessary, incorporating registry auth.
func handleImagePull(cli *client.Client, spec ImageSpec) error {
	pullRef, baseImageName, tag, err := util.NormalizeImageName(spec.Name)
	if err != nil {
		return fmt.Errorf("failed to parse image name '%s': %w", spec.Name, err)
	}
	isLatestTag := tag == "latest"

	log.Printf("Checking image: %s (Resolved Ref: %s, Base Name: %s, Tag: %s, IsLatest: %t)", spec.Name, pullRef, baseImageName, tag, isLatestTag)

	ctx := context.Background()
	localImageDigest := ""
	imageFoundLocally := false

	// Check local cache first
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		log.Printf("Warning: Failed to list images while checking for %s: %v", pullRef, err)
		// Continue attempt to pull
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

	needsPull := false
	if isLatestTag && imageFoundLocally {
		log.Printf("Local image '%s' found. Checking remote registry for updates...", pullRef)
		remoteDigest, err := getRemoteImageDigest(cli, spec) // Pass ImageSpec here
		if err != nil {
			log.Printf("Warning: Failed to check remote registry for '%s': %v. Will attempt pull.", pullRef, err)
			needsPull = true
		} else {
			localDigestClean := strings.TrimPrefix(localImageDigest, "sha256:")
			log.Printf("Comparing local digest (%s) with remote digest (%s)", localDigestClean, remoteDigest)
			if remoteDigest != "" && strings.HasPrefix(localDigestClean, remoteDigest) {
				log.Printf("Local image '%s' is up-to-date.", pullRef)
				return nil // Image is local and up-to-date
			}
			log.Printf("Remote registry has a newer version for '%s'. Pulling update.", pullRef)
			needsPull = true
		}
	} else if imageFoundLocally {
		log.Printf("Image '%s' found locally. Using existing image.", pullRef)
		return nil // Image is local (non-latest tag), use it
	} else {
		log.Printf("Image '%s' not found in local cache. Pulling required.", pullRef)
		needsPull = true
	}

	if needsPull {
		log.Printf("Attempting to pull image reference: %s", pullRef)

		// Build auth header for the pull operation using the potentially derived token
		header, err := buildAuthHeader(spec) // Pass the full spec
		if err != nil {
			return fmt.Errorf("failed to prepare authentication for image pull %s: %w", spec.Name, err)
		}

		opts := dockerTypes.ImagePullOptions{}
		if header != "" {
			log.Printf("Using generated registry credentials header for image pull: %s", pullRef)
			opts.RegistryAuth = header
		} else {
			log.Printf("No specific registry credentials provided or derived for image pull: %s. Relying on local Docker config.", pullRef)
		}

		stream, err := cli.ImagePull(ctx, pullRef, opts)
		if err != nil {
			if errdefs.IsUnauthorized(err) {
				log.Printf("Authentication failed for image pull %s. Check registry_auth or local Docker config.", pullRef)
				if header != "" {
					return fmt.Errorf("authentication failed using provided/derived credentials for image pull %s: %w", pullRef, err)
				}
				return fmt.Errorf("authentication failed for image pull %s (check local Docker config): %w", pullRef, err)
			}
			if errdefs.IsNotFound(err) {
				return fmt.Errorf("image %s not found in registry: %w", pullRef, err)
			}
			return fmt.Errorf("failed to pull image %s: %w", pullRef, err)
		}
		defer stream.Close()

		pullOutput, err := io.ReadAll(stream)
		if err != nil {
			log.Printf("Warning: Error reading image pull stream for %s: %v", pullRef, err)
			return fmt.Errorf("error occurred during image pull stream for %s: %w", pullRef, err)
		}

		outputStr := string(pullOutput)
		if strings.Contains(outputStr, "\"errorDetail\"") || strings.Contains(outputStr, "\"error\"") {
			log.Printf("Image pull stream for %s contained error details: %s", pullRef, outputStr)
			return fmt.Errorf("image pull for %s completed with errors reported in stream", pullRef)
		}

		log.Printf("Successfully initiated pull for image: %s", pullRef)
	}

	return nil
}

// getRemoteImageDigest checks the remote registry for the image digest, using auth if provided.
func getRemoteImageDigest(cli *client.Client, spec ImageSpec) (string, error) {
	pullRef, _, _, err := util.NormalizeImageName(spec.Name)
	if err != nil {
		log.Printf("Error normalizing image name '%s' for digest check: %v", spec.Name, err)
		return "", fmt.Errorf("invalid image name '%s' for digest check: %w", spec.Name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	header, err := buildAuthHeader(spec) // Reuse buildAuthHeader logic
	if err != nil {
		log.Printf("Warning: could not build/fetch registry auth header for digest check %s: %v. Proceeding without explicit auth.", pullRef, err)
		header = ""
	}

	log.Printf("Inspecting remote distribution for %s (using generated auth header: %t)", pullRef, header != "")
	distInspect, err := cli.DistributionInspect(ctx, pullRef, header)
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Printf("Image '%s' not found in remote registry during digest check.", pullRef)
			return "", nil
		}
		if errdefs.IsUnauthorized(err) {
			log.Printf("Authentication failed during remote digest check for %s. Check registry_auth or local Docker config.", pullRef)
			if header != "" {
				return "", fmt.Errorf("authentication failed using provided/derived credentials for remote digest check %s: %w", pullRef, err)
			}
			return "", fmt.Errorf("authentication failed for remote digest check %s (check local Docker config): %w", pullRef, err)
		}
		return "", fmt.Errorf("failed to inspect remote image '%s': %w", pullRef, err)
	}

	if distInspect.Descriptor.Digest == "" {
		log.Printf("Warning: Remote distribution inspection for %s returned no digest.", pullRef)
		return "", nil
	}

	return string(distInspect.Descriptor.Digest), nil
}

// PullImage creates a Docker client and pulls the specified image if needed.
func PullImage(spec ImageSpec) error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(os.Getenv("LEMC_DOCKER_HOST")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	return handleImagePull(cli, spec)
}

// Deprecated functions commented out
// func pullImageV2(cli *client.Client, imagename string, authConfig string) error { ... }
// func createAuthConfig(username, password, registryname string, useToken bool) string { ... }
