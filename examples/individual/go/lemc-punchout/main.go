package main

import (
	"fmt"
	"os"
	"time"
)

func htmlTrunc(s string) {
	time.Sleep(10 * time.Millisecond)
	templ := `<div class="flex items-center justify-center">%s</div>`
	fmt.Printf("lemc.html.trunc; %s\n", fmt.Sprintf(templ, s))
}

func cssTrunc(s string) {
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("lemc.css.trunc; %s\n", s)
}

func main() {

	css := fmt.Sprintf("#%s { font-size: 72px; overflow: auto !important; background: white; }", os.Getenv("LEMC_HTML_ID"))
	cssTrunc(css)

	htmlTrunc("<pre>😇               🥊👿</pre>")

	for i := 0; i < 5; i++ {
		htmlTrunc("<pre>😇               🥊👿</pre>")
		time.Sleep(500 * time.Millisecond)
		htmlTrunc("<pre>😇              🥊 👿</pre>")
		htmlTrunc("<pre>😇             🥊 -👿</pre>")
		htmlTrunc("<pre>😇            🥊 --👿</pre>")
		htmlTrunc("<pre>😇           🥊 ---👿</pre>")
		htmlTrunc("<pre>😇          🥊 ----👿</pre>")
		htmlTrunc("<pre>😇         🥊 -----👿</pre>")
		htmlTrunc("<pre>😇        🥊 ------👿</pre>")
		htmlTrunc("<pre>😇       🥊 -------👿</pre>")
		htmlTrunc("<pre>😇      🥊 --------👿</pre>")
		htmlTrunc("<pre>😇     🥊 ---------👿</pre>")
		htmlTrunc("<pre>😇    🥊 ----------👿</pre>")
		htmlTrunc("<pre>😇   🥊 -----------👿</pre>")
		htmlTrunc("<pre>😇  🥊 ------------👿</pre>")
		htmlTrunc("<pre>😇 🥊 -------------👿</pre>")
		for i1 := 0; i1 < 20; i1++ {
			htmlTrunc("<pre>😫🥊 --------------😈</pre>")
			htmlTrunc("<pre>😫 🥊 -------------😈</pre>")
			htmlTrunc("<pre>😩🥊 --------------😈</pre>")
			htmlTrunc("<pre>😩 🥊 -------------😈</pre>")
		}
		htmlTrunc("<pre>😇 🥊 -------------👿</pre>")
		htmlTrunc("<pre>😇  🥊 ------------👿</pre>")
		htmlTrunc("<pre>😇   🥊 -----------👿</pre>")
		htmlTrunc("<pre>😇    🥊 ----------👿</pre>")
		htmlTrunc("<pre>😇     🥊 ---------👿</pre>")
		htmlTrunc("<pre>😇      🥊 --------👿</pre>")
		htmlTrunc("<pre>😇       🥊 -------👿</pre>")
		htmlTrunc("<pre>😇        🥊 ------👿</pre>")
		htmlTrunc("<pre>😇         🥊 -----👿</pre>")
		htmlTrunc("<pre>😇          🥊 ----👿</pre>")
		htmlTrunc("<pre>😇           🥊 ---👿</pre>")
		htmlTrunc("<pre>😇            🥊 --👿</pre>")
		htmlTrunc("<pre>😇             🥊 -👿</pre>")
		htmlTrunc("<pre>😇              🥊 👿</pre>")
		htmlTrunc("<pre>😇               🥊👿</pre>")
		time.Sleep(500 * time.Millisecond)
	}
	htmlTrunc("<pre>🤕               🥊👿</pre>")
}
