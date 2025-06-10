package models

import (
	"encoding/base64"
	"strings"

	"golang.org/x/net/html"
)

type YamlDefault struct {
	UUID     string `yaml:"-"`
	Cookbook Book   `yaml:"cookbook"`
}

type YamlDefaultNoStorage struct {
	UUID     string        `yaml:"-"`
	Cookbook BookNoStorage `yaml:"cookbook"`
}

type Environment struct {
	Public  []string `yaml:"public"`
	Private []string `yaml:"private"`
}

type Configuration struct {
	Environment Environment `yaml:"environment"`
}

type BookNoStorage struct {
	Environment Environment `yaml:"environment"`
	Pages       []Page      `yaml:"pages"`
}

type Book struct {
	Environment Environment `yaml:"environment"`
	Pages       []Page      `yaml:"pages"`
	Storage     Storage     `yaml:"storage"`
}

type Thumbnail struct {
	Type      string `yaml:"type"`
	B64       string `yaml:"b64"`
	Timestamp string `yaml:"timestamp"`
}
type Storage struct {
	Thumbnail Thumbnail         `yaml:"thumbnail"`
	Files     map[string]string `yaml:"files"`
	Wikis     map[int]string    `yaml:"wikis"`
}

func extractImgSrcs(htmlStr string) ([]string, error) {
	var srcs []string
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					srcs = append(srcs, attr.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return srcs, nil
}

func getLastParameter(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func (storage *Storage) PurgeUnusedFiles() error {

	for name := range storage.Files {
		deleteThis := true

		for _, b64wiki := range storage.Wikis {
			byteswiki, err := base64.StdEncoding.DecodeString(b64wiki)
			if err != nil {
				return err
			}

			imgsrc, err := extractImgSrcs(string(byteswiki))
			if err != nil {
				return err
			}

			for _, src := range imgsrc {
				last := getLastParameter(src)
				if name == last {
					deleteThis = false
				}
			}
		}

		if deleteThis {
			delete(storage.Files, name)
		}
	}

	return nil
}

type Wiki struct {
	Page int    `yaml:"page"`
	B64  string `yaml:"b64"`
}

type Page struct {
	PageID    int      `yaml:"page"`
	Name      string   `yaml:"name"`
	Recipes   []Recipe `yaml:"recipes"`
	HtmlCache string   `yaml:"-"`
	CssCache  string   `yaml:"-"`
	JsCache   string   `yaml:"-"`
}

type Recipe struct {
	IsShared    bool        `yaml:"-"` // used for telling the job how to run, as a user or admin
	Name        string      `yaml:"recipe"`
	Description string      `yaml:"description"`
	Form        []FormField `yaml:"form,omitempty"`
	Steps       []Step      `yaml:"steps"`
}

type FormField struct {
	Name        string            `yaml:"name,omitempty"`        // Deprecated: use Variable instead
	Variable    string            `yaml:"variable,omitempty"`    // New field for environment variable name
	Description string            `yaml:"description,omitempty"` // New field for display label
	Type        string            `yaml:"type"`                  // input, text, radio, select, password, textarea
	Defaults    []string          `yaml:"defaults,omitempty"`
	Options     []FormFieldOption `yaml:"options,omitempty"`
}

type FormFieldOption struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// GetOptions returns form field options, using Options if available, otherwise falling back to Defaults
func (f *FormField) GetOptions() []FormFieldOption {
	if len(f.Options) > 0 {
		return f.Options
	}

	// Fallback to Defaults for backward compatibility
	var options []FormFieldOption
	for _, d := range f.Defaults {
		options = append(options, FormFieldOption{
			Label: d,
			Value: d,
		})
	}
	return options
}

// GetPlaceholder returns the first default value as placeholder for text inputs
func (f *FormField) GetPlaceholder() string {
	if len(f.Defaults) > 0 {
		return f.Defaults[0]
	}
	return ""
}

// IsSelectType returns true if the field type is radio or select
func (f *FormField) IsSelectType() bool {
	return f.Type == "radio" || f.Type == "select"
}

// GetVariable returns the variable name, preferring Variable over Name for backward compatibility
func (f *FormField) GetVariable() string {
	if f.Variable != "" {
		return f.Variable
	}
	return f.Name // Fallback to Name for backward compatibility
}

// GetDisplayName returns the description if available, otherwise falls back to variable name
func (f *FormField) GetDisplayName() string {
	if f.Description != "" {
		return f.Description
	}
	return f.GetVariable() // Fallback to variable name
}

func (r *Recipe) UsernameOrAdmin() string {
	if r.IsShared {
		return "shared"
	}
	return r.Name
}

type Step struct {
	Step         int      `yaml:"step"`
	Name         string   `yaml:"name"`
	Image        string   `yaml:"image"`
	RegistryAuth string   `yaml:"registry_auth,omitempty"`
	Entrypoint   []string `yaml:"entrypoint,omitempty"`
	Env          []string `yaml:"env,omitempty"`         // Deprecated: use Environment instead
	Environment  []string `yaml:"environment,omitempty"` // New field for environment variables
	Do           string   `yaml:"do"`
	Timeout      string   `yaml:"timeout"`
}

// GetEnvironment returns environment variables, preferring Environment over Env for backward compatibility
func (s *Step) GetEnvironment() []string {
	if len(s.Environment) > 0 {
		return s.Environment
	}
	return s.Env // Fallback to Env for backward compatibility
}

func NewYamlIndividual() *YamlDefault {
	return &YamlDefault{
		Cookbook: Book{
			Environment: Environment{},
			Pages:       []Page{},
			Storage: Storage{
				Files: map[string]string{},
				Wikis: map[int]string{},
			},
		},
	}
}
