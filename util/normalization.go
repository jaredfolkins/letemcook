package util

import (
	"fmt"
	"regexp"
	"strings"

	normalizer "github.com/dimuska139/go-email-normalizer"
)

func ReplaceSpecialCharsWithDashes(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return re.ReplaceAllString(input, "-")
}

func AlphaNumHyphen(s string) string {
	n := normalizer.NewNormalizer()
	s = strings.ToLower(s)
	normalized := n.Normalize(s)

	re := regexp.MustCompile(`\W+`)
	normalized = re.ReplaceAllString(s, "-")

	return normalized
}

type JobMeta struct {
	UUID       string
	PageID     string
	StepID     string
	UserID     string
	Username   string
	RecipeName string
	Scope      string
}

func (jm *JobMeta) GenerateContainerName(recipeName, adminOrUsername string) string {
	return fmt.Sprintf("uuid-%s-page-%s-recipe-%s-step-%s-scope-%s-username-%s", jm.UUID, jm.PageID, AlphaNumHyphen(recipeName), jm.StepID, jm.Scope, adminOrUsername)
}

func (jm *JobMeta) createTagMap() map[string]string {
	tagMap := make(map[string]string)
	tagMap["UUID"] = jm.UUID
	tagMap["PAGE_ID"] = jm.PageID
	tagMap["USER_ID"] = jm.UserID
	tagMap["USERNAME"] = jm.Username
	tagMap["RECIPE_NAME"] = jm.RecipeName
	tagMap["STEP_ID"] = jm.StepID
	tagMap["SCOPE"] = jm.Scope
	return tagMap
}

func NewJobMetaFromEnv(env []string) *JobMeta {
	jm := &JobMeta{}
	for _, envVar := range env {
		keyValue := strings.SplitN(envVar, "=", 2)
		if len(keyValue) == 2 {
			key := keyValue[0]
			v := keyValue[1]
			switch key {
			case "LEMC_UUID":
				jm.UUID = v
			case "LEMC_PAGE_ID":
				jm.PageID = v
			case "LEMC_USER_ID":
				jm.UserID = v
			case "LEMC_USERNAME":
				jm.Username = v
			case "LEMC_RECIPE_NAME":
				jm.RecipeName = v
			case "LEMC_STEP_ID":
				jm.StepID = v
			case "LEMC_SCOPE":
				jm.Scope = v
			}
		}
	}
	return jm
}
