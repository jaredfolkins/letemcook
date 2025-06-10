package util

import "testing"

func TestGenerateContainerName(t *testing.T) {
	jm := &JobMeta{
		UUID:       "abc123",
		PageID:     "1",
		StepID:     "2",
		UserID:     "42",
		Username:   "alice",
		RecipeName: "My Recipe",
		Scope:      "individual",
	}
	got := jm.GenerateContainerName(jm.RecipeName, jm.Username)
	expect := "uuid-abc123-page-1-recipe-my-recipe-step-2-scope-individual-username-alice"
	if got != expect {
		t.Errorf("expected %q, got %q", expect, got)
	}
}

func TestNewJobMetaFromEnv(t *testing.T) {
	env := []string{
		"LEMC_UUID=uuid1",
		"LEMC_PAGE_ID=p1",
		"LEMC_USER_ID=u1",
		"LEMC_USERNAME=bob",
		"LEMC_RECIPE_NAME=demo",
		"LEMC_STEP_ID=s1",
		"LEMC_SCOPE=test",
		"UNRELATED=skip",
	}

	jm := NewJobMetaFromEnv(env)

	if jm.UUID != "uuid1" || jm.PageID != "p1" || jm.UserID != "u1" ||
		jm.Username != "bob" || jm.RecipeName != "demo" || jm.StepID != "s1" || jm.Scope != "test" {
		t.Errorf("parsed job meta mismatch: %#v", jm)
	}
}

func TestAlphaNumHyphen(t *testing.T) {
	in := "Hello World@2024"
	expected := "hello-world-2024"
	if out := AlphaNumHyphen(in); out != expected {
		t.Errorf("AlphaNumHyphen(%q) = %q, want %q", in, out, expected)
	}
}
func TestAlphaNumHyphenNormalization(t *testing.T) {
	in := "User+tag@Gmail.com"
	expected := "user-gmail-com"
	if out := AlphaNumHyphen(in); out != expected {
		t.Errorf("AlphaNumHyphen(%q) = %q, want %q", in, out, expected)
	}
}
