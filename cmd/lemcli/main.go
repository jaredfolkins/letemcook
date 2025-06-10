package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
