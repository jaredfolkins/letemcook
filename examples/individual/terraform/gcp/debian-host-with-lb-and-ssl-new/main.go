package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	sourceDir  = "terraform-config"
	dotEnvPath = "dotenv"
)

const (
	destDir        = "private"
	tfVarsFilename = "terraform.tfvars"
)

// Parses a .env file and returns a map of key-value pairs.
func parseDotEnv(filePath string) (map[string]string, error) {
	envVars := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if present
			if len(value) > 1 && ((strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) || (strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`))) {
				value = value[1 : len(value)-1]
			}
			envVars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	return envVars, nil
}

// Copies a single file from src to dst.
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("Running in directory: %s", cwd)

	log.Println("Starting Terraform setup...")

	dotEnvPath = filepath.Join(cwd, dotEnvPath)
	sourceDir = filepath.Join(cwd, sourceDir)

	// 1. Parse .env file
	log.Printf("Parsing %s...", dotEnvPath)
	envMap, err := parseDotEnv(dotEnvPath)
	if err != nil {
		log.Fatalf("Error parsing .env file: %v", err)
	}
	log.Printf("Parsed %d variables from %s", len(envMap), dotEnvPath)

	// Check for GCP key file from command line arguments
	var keyFilePath string
	flag.StringVar(&keyFilePath, "key", "", "Path to GCP JSON key file")
	flag.Parse()
	log.Printf("Key file path: %s", keyFilePath)

	// Exit if key file path was not provided
	if keyFilePath == "" {
		log.Fatalf("Error: -key flag is required. Please provide the path to the GCP JSON key file.")
	}

	// If a key file was provided, decode it and set GCP_PROJECT_ID
	if keyFilePath != "" {
		log.Printf("Reading GCP key file from: %s", keyFilePath)
		keyData, err := os.ReadFile(keyFilePath)
		if err != nil {
			log.Fatalf("Error reading GCP key file: %v", err)
		}

		// Define a struct to decode the JSON key file
		type GCPCredentials struct {
			ProjectID string `json:"project_id"`
		}

		var credentials GCPCredentials
		if err := json.Unmarshal(keyData, &credentials); err != nil {
			log.Fatalf("Error parsing GCP key file: %v", err)
		}

		if credentials.ProjectID == "" {
			log.Fatalf("No project_id found in GCP key file")
		}

		log.Printf("Setting GCP_PROJECT_ID to: %s", credentials.ProjectID)
		envMap["GCP_PROJECT_ID"] = credentials.ProjectID
	}

	// Use placeholder if LEMC_UUID is commented out or missing
	lemcUUID, ok := envMap["LEMC_UUID"]
	if !ok || lemcUUID == "" {
		log.Println("LEMC_UUID not found in .env, using placeholder 'default-uuid'")
		lemcUUID = "default-uuid" // Provide a default or generate one
	}

	// 2. Create destination directory
	log.Printf("Ensuring destination directory '%s' exists...", destDir)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory %s: %v", destDir, err)
	}

	// 3. Copy files from source directory
	log.Printf("Copying files from %s to %s...", sourceDir, destDir)
	filesToCopy := []string{"main.tf", "outputs.tf", "variables.tf"}
	for _, filename := range filesToCopy {
		srcPath := filepath.Join(sourceDir, filename)
		dstPath := filepath.Join(destDir, filename)
		log.Printf("  Copying %s to %s", srcPath, dstPath)
		if err := copyFile(srcPath, dstPath); err != nil {
			log.Fatalf("Error copying file %s: %v", filename, err)
		}
	}
	log.Println("Finished copying base Terraform files.")

	// 4. Map .env vars to Terraform vars and prepare tfvars content
	log.Println("Generating terraform.tfvars content...")
	var tfVarsContent strings.Builder

	// Mapping from .env key to terraform variable name
	varMap := map[string]string{
		"GCP_PROJECT_ID": "gcp_project_id",
		"GCP_REGION":     "gcp_region",
		"GCP_ZONE":       "gcp_zone",
		"LEMC_SCOPE":     "lemc_scope", // Note: TF variable is uppercase
		"LEMC_USERNAME":  "lemc_username",
		"LEMC_USER_ID":   "lemc_user_id", // Added mapping for user id
	}

	// --- Construct Prefix --- moved calculation earlier
	var prefix, uuid, user, scope, userid, rootDomain string
	uuid = envMap["LEMC_UUID"]
	user = envMap["LEMC_USERNAME"]
	scope = envMap["LEMC_SCOPE"]
	userid = envMap["LEMC_USER_ID"]
	rootDomain = envMap["ROOT_DOMAIN"] // Keep for domain construction
	uuidPrefix := ""
	if len(uuid) >= 8 {
		uuidPrefix = uuid[:8]
	} else {
		log.Printf("Warning: LEMC_UUID ('%s') is shorter than 8 characters.", uuid)
		uuidPrefix = uuid // Use the whole short UUID
	}

	if user == "" {
		log.Printf("Warning: LEMC_USERNAME not found in %s, required for resource prefix.", dotEnvPath)
	}
	if scope == "" {
		log.Printf("Warning: LEMC_SCOPE not found in %s, required for resource prefix.", dotEnvPath)
	}
	if userid == "" {
		log.Printf("Warning: LEMC_USER_ID not found in %s, required for resource prefix.", dotEnvPath)
	}

	scopePrefix := strings.ToLower(scope)
	if scopePrefix == "individual" {
		scopePrefix = "ind"
	} else if scopePrefix == "shared" {
		scopePrefix = "shd"
	}

	if len(envMap["DEFAULT_STATIC_PREFIX"]) < 4 {
		prefix = fmt.Sprintf("lemc-%s-%s-%s-%s", uuidPrefix, user, userid, scopePrefix)
		log.Printf("Constructed resource prefix: %s", prefix)
	} else {
		prefix = envMap["DEFAULT_STATIC_PREFIX"]
		log.Printf("Using default static prefix: %s", prefix)
	}

	// Add explicitly mapped variables (including project ID from key)
	for envKey, tfKey := range varMap {
		if value, ok := envMap[envKey]; ok {
			tfVarsContent.WriteString(fmt.Sprintf("%s = \"%s\"\n", tfKey, value))
		} else {
			log.Printf("Warning: Expected variable %s not found in %s", envKey, dotEnvPath)
			// Optionally add a default or skip
			// tfVarsContent.WriteString(fmt.Sprintf("%s = \"\"\n", tfKey))
		}
	}

	// Add lemc_uuid separately
	tfVarsContent.WriteString(fmt.Sprintf("lemc_uuid = \"%s\"\n", lemcUUID))

	// Add the constructed resource prefix
	tfVarsContent.WriteString(fmt.Sprintf("resource_prefix = \"%s\"\n", prefix))

	// --- Get machine_type ---
	if machineType, ok := envMap["FORM_MACHINE_TYPE"]; ok {
		tfVarsContent.WriteString(fmt.Sprintf("machine_type = \"%s\"\n", machineType))
	} else {
		log.Printf("Warning: FORM_MACHINE_TYPE not found in %s. Terraform will use its default.", dotEnvPath)
		// You might want to set a default in variables.tf instead of here
	}

	// --- Get image ---
	if imageName, ok := envMap["FORM_IMAGE"]; ok {
		tfVarsContent.WriteString(fmt.Sprintf("image = \"%s\"\n", imageName))
	} else {
		log.Printf("Warning: FORM_IMAGE not found in %s. Terraform will use its default.", dotEnvPath)
	}

	// --- Construct domain_name ---
	// Use the already calculated prefix
	constructedDomainName := ""
	if rootDomain != "" && prefix != "" {
		constructedDomainName = fmt.Sprintf("%s.%s", prefix, rootDomain)
		log.Printf("  Constructed domain_name: %s", constructedDomainName)
	} else {
		if rootDomain == "" {
			log.Printf("Warning: ROOT_DOMAIN not found in %s, cannot construct domain_name.", dotEnvPath)
		}
		if prefix == "" {
			log.Printf("Warning: Could not construct resource prefix, cannot construct domain_name.")
		}
	}

	if constructedDomainName == "" {
		log.Printf("Warning: Could not construct domain_name from .env variables. Ensure LEMC_UUID, LEMC_USERNAME, LEMC_USER_ID, LEMC_SCOPE, and ROOT_DOMAIN are set.")
		// Provide a default or handle error if domain_name is mandatory
		tfVarsContent.WriteString(fmt.Sprintf("domain_name = \"%s\"\n", "default.example.com")) // Placeholder
	} else {
		// Add the constructed domain name if successful
		tfVarsContent.WriteString(fmt.Sprintf("domain_name = \"%s\"\n", constructedDomainName))
	}

	// --- Get dns_zone_name ---
	if rootZone, ok := envMap["ROOT_ZONE"]; ok {
		tfVarsContent.WriteString(fmt.Sprintf("dns_zone_name = \"%s\"\n", rootZone))
	} else {
		log.Printf("Warning: ROOT_ZONE not found in %s, required for dns_zone_name.", dotEnvPath)
		// Provide a default or handle error if dns_zone_name is mandatory
		tfVarsContent.WriteString(fmt.Sprintf("dns_zone_name = \"%s\"\n", "default-zone-name")) // Placeholder
	}

	// 5. Write terraform.tfvars file
	tfVarsPath := filepath.Join(destDir, tfVarsFilename)
	log.Printf("Writing %s...", tfVarsPath)
	err = os.WriteFile(tfVarsPath, []byte(tfVarsContent.String()), 0644)
	if err != nil {
		log.Fatalf("Error writing %s: %v", tfVarsPath, err)
	}

	log.Printf("Successfully created %s", tfVarsPath)
	log.Println("Terraform setup complete. You can now run 'terraform plan/apply' in the 'private' directory.")
}
