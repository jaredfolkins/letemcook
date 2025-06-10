package models

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestStepGetEnvironment tests the GetEnvironment method with both old and new field formats
func TestStepGetEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		step     Step
		expected []string
	}{
		{
			name: "legacy env field only",
			step: Step{
				Env: []string{"LEGACY_VAR=old_value", "TEST_VAR=test"},
			},
			expected: []string{"LEGACY_VAR=old_value", "TEST_VAR=test"},
		},
		{
			name: "new environment field only",
			step: Step{
				Environment: []string{"NEW_VAR=new_value", "CONFIG_VAR=config"},
			},
			expected: []string{"NEW_VAR=new_value", "CONFIG_VAR=config"},
		},
		{
			name: "both fields present - environment takes precedence",
			step: Step{
				Env:         []string{"LEGACY_VAR=old_value"},
				Environment: []string{"NEW_VAR=new_value", "PREFERRED_VAR=preferred"},
			},
			expected: []string{"NEW_VAR=new_value", "PREFERRED_VAR=preferred"},
		},
		{
			name:     "no environment fields",
			step:     Step{},
			expected: nil,
		},
		{
			name: "empty environment field falls back to env",
			step: Step{
				Env:         []string{"FALLBACK_VAR=fallback"},
				Environment: []string{},
			},
			expected: []string{"FALLBACK_VAR=fallback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.step.GetEnvironment()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetEnvironment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestYAMLParsingWithLegacyEnvField tests YAML parsing with the old env field
func TestYAMLParsingWithLegacyEnvField(t *testing.T) {
	yamlContent := `
cookbook:
  environment:
    public:
      - TEST_ENV=test
  pages:
    - page: 1
      name: Test Page
      recipes:
        - recipe: test recipe
          description: Test recipe with legacy env field
          form: []
          steps:
            - step: 1
              name: test step
              image: test:latest
              env:
                - LEGACY_VAR=old_value
                - CONFIG_VAR=legacy_config
              do: now
              timeout: 5.minutes
`

	var yamlDefault YamlDefault
	err := yaml.Unmarshal([]byte(yamlContent), &yamlDefault)
	if err != nil {
		t.Fatalf("Failed to parse YAML with legacy env field: %v", err)
	}

	if len(yamlDefault.Cookbook.Pages) != 1 {
		t.Fatalf("Expected 1 page, got %d", len(yamlDefault.Cookbook.Pages))
	}

	page := yamlDefault.Cookbook.Pages[0]
	if len(page.Recipes) != 1 {
		t.Fatalf("Expected 1 recipe, got %d", len(page.Recipes))
	}

	recipe := page.Recipes[0]
	if len(recipe.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(recipe.Steps))
	}

	step := recipe.Steps[0]
	envVars := step.GetEnvironment()
	expected := []string{"LEGACY_VAR=old_value", "CONFIG_VAR=legacy_config"}

	if !reflect.DeepEqual(envVars, expected) {
		t.Errorf("Expected environment variables %v, got %v", expected, envVars)
	}
}

// TestYAMLParsingWithNewEnvironmentField tests YAML parsing with the new environment field
func TestYAMLParsingWithNewEnvironmentField(t *testing.T) {
	yamlContent := `
cookbook:
  environment:
    public:
      - TEST_ENV=test
  pages:
    - page: 1
      name: Test Page
      recipes:
        - recipe: test recipe
          description: Test recipe with new environment field
          form: []
          steps:
            - step: 1
              name: test step
              image: test:latest
              environment:
                - NEW_VAR=new_value
                - ENHANCED_CONFIG=modern_config
              do: now
              timeout: 5.minutes
`

	var yamlDefault YamlDefault
	err := yaml.Unmarshal([]byte(yamlContent), &yamlDefault)
	if err != nil {
		t.Fatalf("Failed to parse YAML with new environment field: %v", err)
	}

	step := yamlDefault.Cookbook.Pages[0].Recipes[0].Steps[0]
	envVars := step.GetEnvironment()
	expected := []string{"NEW_VAR=new_value", "ENHANCED_CONFIG=modern_config"}

	if !reflect.DeepEqual(envVars, expected) {
		t.Errorf("Expected environment variables %v, got %v", expected, envVars)
	}
}

// TestYAMLParsingWithMixedEnvironmentFields tests YAML parsing with both old and new fields
func TestYAMLParsingWithMixedEnvironmentFields(t *testing.T) {
	yamlContent := `
cookbook:
  environment:
    public:
      - TEST_ENV=test
  pages:
    - page: 1
      name: Test Page
      recipes:
        - recipe: mixed format recipe
          description: Recipe with mixed environment field formats
          form: []
          steps:
            - step: 1
              name: legacy step
              image: legacy:latest
              env:
                - LEGACY_VAR=legacy_value
              do: now
              timeout: 5.minutes
            - step: 2
              name: modern step
              image: modern:latest
              environment:
                - MODERN_VAR=modern_value
                - ENHANCED_VAR=enhanced
              do: now
              timeout: 5.minutes
            - step: 3
              name: precedence test step
              image: test:latest
              env:
                - OLD_VAR=should_not_be_used
              environment:
                - NEW_VAR=should_be_used
              do: now
              timeout: 5.minutes
`

	var yamlDefault YamlDefault
	err := yaml.Unmarshal([]byte(yamlContent), &yamlDefault)
	if err != nil {
		t.Fatalf("Failed to parse YAML with mixed environment fields: %v", err)
	}

	recipe := yamlDefault.Cookbook.Pages[0].Recipes[0]
	if len(recipe.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(recipe.Steps))
	}

	// Test legacy step
	legacyStep := recipe.Steps[0]
	legacyEnv := legacyStep.GetEnvironment()
	expectedLegacy := []string{"LEGACY_VAR=legacy_value"}
	if !reflect.DeepEqual(legacyEnv, expectedLegacy) {
		t.Errorf("Legacy step: expected %v, got %v", expectedLegacy, legacyEnv)
	}

	// Test modern step
	modernStep := recipe.Steps[1]
	modernEnv := modernStep.GetEnvironment()
	expectedModern := []string{"MODERN_VAR=modern_value", "ENHANCED_VAR=enhanced"}
	if !reflect.DeepEqual(modernEnv, expectedModern) {
		t.Errorf("Modern step: expected %v, got %v", expectedModern, modernEnv)
	}

	// Test precedence step (environment should take precedence over env)
	precedenceStep := recipe.Steps[2]
	precedenceEnv := precedenceStep.GetEnvironment()
	expectedPrecedence := []string{"NEW_VAR=should_be_used"}
	if !reflect.DeepEqual(precedenceEnv, expectedPrecedence) {
		t.Errorf("Precedence step: expected %v, got %v", expectedPrecedence, precedenceEnv)
	}
}

// TestYAMLMarshalUnmarshalRoundTrip tests that YAML can be marshaled and unmarshaled correctly
func TestYAMLMarshalUnmarshalRoundTrip(t *testing.T) {
	original := YamlDefault{
		Cookbook: Book{
			Environment: Environment{
				Public:  []string{"PUBLIC_VAR=public"},
				Private: []string{"PRIVATE_VAR=private"},
			},
			Pages: []Page{
				{
					PageID: 1,
					Name:   "Test Page",
					Recipes: []Recipe{
						{
							Name:        "test recipe",
							Description: "Test recipe for round trip",
							Form:        []FormField{},
							Steps: []Step{
								{
									Step:    1,
									Name:    "test step with legacy env",
									Image:   "test:latest",
									Env:     []string{"LEGACY_VAR=legacy"},
									Do:      "now",
									Timeout: "5.minutes",
								},
								{
									Step:        2,
									Name:        "test step with new environment",
									Image:       "test:latest",
									Environment: []string{"NEW_VAR=new", "CONFIG_VAR=config"},
									Do:          "now",
									Timeout:     "5.minutes",
								},
							},
						},
					},
				},
			},
			Storage: Storage{
				Files: map[string]string{},
				Wikis: map[int]string{},
			},
		},
	}

	// Marshal to YAML
	marshaled, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	// Unmarshal back
	var unmarshaled YamlDefault
	err = yaml.Unmarshal(marshaled, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Verify the data is preserved
	if len(unmarshaled.Cookbook.Pages) != 1 {
		t.Fatalf("Expected 1 page after round trip, got %d", len(unmarshaled.Cookbook.Pages))
	}

	steps := unmarshaled.Cookbook.Pages[0].Recipes[0].Steps
	if len(steps) != 2 {
		t.Fatalf("Expected 2 steps after round trip, got %d", len(steps))
	}

	// Test legacy step environment
	legacyEnv := steps[0].GetEnvironment()
	expectedLegacy := []string{"LEGACY_VAR=legacy"}
	if !reflect.DeepEqual(legacyEnv, expectedLegacy) {
		t.Errorf("Round trip legacy step: expected %v, got %v", expectedLegacy, legacyEnv)
	}

	// Test new environment step
	newEnv := steps[1].GetEnvironment()
	expectedNew := []string{"NEW_VAR=new", "CONFIG_VAR=config"}
	if !reflect.DeepEqual(newEnv, expectedNew) {
		t.Errorf("Round trip new environment step: expected %v, got %v", expectedNew, newEnv)
	}
}

// TestStepEnvironmentFieldsInYAMLOutput tests that both fields are preserved in YAML output
func TestStepEnvironmentFieldsInYAMLOutput(t *testing.T) {
	step := Step{
		Step:        1,
		Name:        "test step",
		Image:       "test:latest",
		Env:         []string{"OLD_VAR=old"},
		Environment: []string{"NEW_VAR=new"},
		Do:          "now",
		Timeout:     "5.minutes",
	}

	marshaled, err := yaml.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal step: %v", err)
	}

	yamlStr := string(marshaled)

	// Verify both fields are present in the YAML output
	if !strings.Contains(yamlStr, "env:") {
		t.Error("Expected 'env:' field in YAML output")
	}
	if !strings.Contains(yamlStr, "environment:") {
		t.Error("Expected 'environment:' field in YAML output")
	}
	if !strings.Contains(yamlStr, "OLD_VAR=old") {
		t.Error("Expected 'OLD_VAR=old' in YAML output")
	}
	if !strings.Contains(yamlStr, "NEW_VAR=new") {
		t.Error("Expected 'NEW_VAR=new' in YAML output")
	}
}
