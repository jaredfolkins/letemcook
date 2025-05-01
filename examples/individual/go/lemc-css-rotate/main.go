package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	var htmlid string
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "LEMC_HTML_ID":
				htmlid = parts[1]
			}
		}
	}

	arr := [8]string{"#ffadad", "#ffd6a5", "#fdffb6", "#caffbf", "#9bf6ff", "#a0c4ff", "#bdb2ff", "#ffc6ff"}

	fmt.Println("lemc.html.trunc;")

	shuffleArray(arr[:])
	for i := 0; i < 3; i++ {
		for _, color := range arr {
			fmt.Printf("lemc.css.trunc; #%s { background-color: %s; }\n", htmlid, color)
			fmt.Printf("lemc.html.append; %s<br>\n", color)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func shuffleArray(arr []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}
