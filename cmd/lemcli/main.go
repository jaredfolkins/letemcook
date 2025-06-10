package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lemcli secrets <command> [args]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "secrets":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lemcli secrets [init|set|get] ...")
			os.Exit(1)
		}
		handleSecrets(os.Args[2:])
	case "cookbook":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lemcli cookbook init <name>")
			os.Exit(1)
		}
		handleCookbook(os.Args[2:])
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func secretsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Cannot determine home directory:", err)
		os.Exit(1)
	}
	dir := filepath.Join(home, ".lemc", "secrets")
	if err := os.MkdirAll(dir, 0700); err != nil {
		fmt.Println("Cannot create secrets directory:", err)
		os.Exit(1)
	}
	return dir
}

const cookbookPrefix = "lemc-"

const defaultCookbookYAML = `cookbook:
    environment:
        public:
            - USER_DEFINED_PUBLIC_ENV_VAR=somesillypublicvar
        private:
            - USER_DEFINED_PRIVATE_ENV_VAR=somesillyprivatevarthatyouwantsecret
    pages:
        - page: 1
          name: Hello World Page
          recipes:
            - recipe: hello world
              description: basic hello world lemc example
              form: []
              steps:
                - step: 1
                  image: docker.io/jfolkins/lemc-helloworld:latest
                  do: now
                  timeout: 10.minutes
`

func handleSecrets(args []string) {
	cmd := args[0]
	switch cmd {
	case "init":
		if len(args) != 2 {
			fmt.Println("Usage: lemcli secrets init <cookbook>")
			os.Exit(1)
		}
		dir := filepath.Join(secretsDir(), args[1])
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Initialized secrets directory:", dir)
	case "set":
		fs := flag.NewFlagSet("set", flag.ExitOnError)
		var value string
		fs.StringVar(&value, "value", "", "secret value")
		fs.Parse(args[1:])
		if fs.NArg() != 2 || value == "" {
			fmt.Println("Usage: lemcli secrets set <cookbook> <key> -value <value>")
			os.Exit(1)
		}
		cookbook := fs.Arg(0)
		key := fs.Arg(1)
		dir := filepath.Join(secretsDir(), cookbook)
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		path := filepath.Join(dir, key)
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			fmt.Println("Error writing secret:", err)
			os.Exit(1)
		}
		fmt.Println("Stored secret at:", path)
	case "get":
		if len(args) != 3 {
			fmt.Println("Usage: lemcli secrets get <cookbook> <key>")
			os.Exit(1)
		}
		dir := filepath.Join(secretsDir(), args[1])
		path := filepath.Join(dir, args[2])
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading secret:", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	default:
		fmt.Println("Usage: lemcli secrets [init|set|get]")
		os.Exit(1)
	}
}

func handleCookbook(args []string) {
	cmd := args[0]
	switch cmd {
	case "init":
		if len(args) != 2 {
			fmt.Println("Usage: lemcli cookbook init <name>")
			os.Exit(1)
		}
		name := args[1]
		if !strings.HasPrefix(name, cookbookPrefix) {
			name = cookbookPrefix + name
		}
		if err := os.MkdirAll(name, 0755); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		path := filepath.Join(name, "cookbook.yaml")
		if err := os.WriteFile(path, []byte(defaultCookbookYAML), 0644); err != nil {
			fmt.Println("Error writing cookbook.yaml:", err)
			os.Exit(1)
		}
		fmt.Println("Initialized cookbook at:", path)
	default:
		fmt.Println("Usage: lemcli cookbook init <name>")
		os.Exit(1)
	}
}
