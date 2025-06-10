package main

import "fmt"

func main() {
	//fmt.Printf("lemc.html.trunc; This is Step 2 and a totally different container...<br>\n")
	//fmt.Printf("lemc.html.append; Updating env vars...<br>\n")
	//time.Sleep(1 * time.Second)
	//fmt.Printf("lemc.env; NEW_VAR=YAY\n")

	/*
		for i0 := 0; i0 <= 2; i0++ {
			var s string
			for i := 0; i <= 5000; i++ {
				s += fmt.Sprintf("%d %d<br>", i0, i)
			}
			fmt.Printf("<lemc-html-append>\n")
			fmt.Printf("%s\n", s)
			fmt.Printf("</lemc-html-append>\n")
		}
	*/

	fmt.Printf("<lemc-html-append onclick='alert(1)'><h1>HELLO</h1></lemc-html-append>\n")

}
