package yeschef

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/jaredfolkins/letemcook/util"
)

func TestMsgLogsCorrectStepID(t *testing.T) {
	env := []string{
		"LEMC_UUID=test-uuid",
		"LEMC_PAGE_ID=1",
		"LEMC_USER_ID=42",
		"LEMC_USERNAME=testuser",
		"LEMC_RECIPE_NAME=test-recipe",
		"LEMC_STEP_ID=3",
		"LEMC_SCOPE=individual",
	}

	jm := util.NewJobMetaFromEnv(env)

	var buf bytes.Buffer
	lf := &util.LogFile{EventID: "event", Writer: bufio.NewWriter(&buf)}

	// Minimal ChefsKiss to avoid nil pointer in msg
	XoxoX = &ChefsKiss{apps: make(map[int64]*CmdServer)}

	job := &JobRecipe{Scope: "individual", UserID: "42"}

	msg("hello", "abcdef12", "testimg", job, jm, &util.ContainerFiles{}, lf)
	lf.Writer.Flush()

	if !strings.Contains(buf.String(), "[step:3]") {
		t.Fatalf("expected log to contain step id 3, got %s", buf.String())
	}
}
