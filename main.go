package main

import (
	"flag"
	"fmt"
)

func main() {
	st := flag.Bool("static", false, "Static")
	flag.Parse()

	buildDatabase()

	if *st {
		static()
	} else {
		dynamic()
	}

	//static()
}

func static() {
	fmt.Print("Static")
}

func dynamic() {
	fmt.Print("Dynamic")
	startWebServer()
}
