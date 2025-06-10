package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	fmt.Printf("lemc.html.trunc; <h3>This is Step 1 and is printing env vars...</h3><br>\n\n")

	fmt.Printf("lemc.html.append; <h4>LEMC ENV VARS</h4><br>\n")

	envVars := os.Environ()
	sort.Strings(envVars)
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "LEMC_") {
			printLemc(envVar)
		}
	}

	fmt.Printf("lemc.html.append; <br><h4>SYSTEM env vars</h4><br>\n")

	for _, envVar := range envVars {
		if !strings.HasPrefix(envVar, "LEMC_") {
			printLemc(envVar)
		}
	}
}

func printLemc(s string) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		fmt.Printf("lemc.html.append; %s=%s<br>\n", parts[0], parts[1])
	} else {
		fmt.Printf("lemc.html.append; Invalid env var: %s<br>\n", s)
	}
}
