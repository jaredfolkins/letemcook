package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Printf("lemc.html.trunc; This is Step 3 and yet another totally different container...<br>\n")
	fmt.Printf("lemc.html.append; Printing new env vars and we should see NEW_VAR in output...<br>\n")
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "BIG_VAR" {
				fmt.Printf("lemc.html.append; <div style='color:purple;'>Key: %s, Value: %s</style><br>\n", parts[0], parts[1])
			}

			/*
				if parts[0] == "NEW_VAR" {
					fmt.Printf("lemc.html.append; <div style='color:red;'>Key: %s, Value: %s</style><br>\n", parts[0], parts[1])
				} else if parts[0] == "BIG_VAR" {
					fmt.Printf("lemc.html.append; <div style='color:green;'>Key: %s, Value: %s</style><br>\n", parts[0], parts[1])
				} else if parts[0] == "BIG_VAR_SIZE" {
					fmt.Printf("lemc.html.append; <div style='color:purple;'>Key: %s, Value: %s</style><br>\n", parts[0], parts[1])
				} else {
					fmt.Printf("lemc.html.append; Key: %s, Value: %s<br>\n", parts[0], parts[1])
				}
			*/
		} else {
			fmt.Printf("lemc.html.append; Invalid env var: %s<br>\n", envVar)
		}
	}
}
