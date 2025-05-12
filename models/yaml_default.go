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

	for name, _ := range storage.Files {
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
	Form        []FormField `yaml:"form"`
	Steps       []Step      `yaml:"steps"`
}

type FormField struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"` // input, text, radio, select, password, textarea
	Defaults []string `yaml:"defaults"`
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
	Command      []string `yaml:"command,omitempty"`
	Do           string   `yaml:"do"`
	Timeout      string   `yaml:"timeout"`
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
