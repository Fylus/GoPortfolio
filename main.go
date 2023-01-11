package main

import (
	"flag"
	"fmt"
)

func main() {
	st := flag.Bool("static", false, "Static")
	flag.Parse()

	if *st {
		static()
	} else {
		dynamic()
	}

	buildDatabase()
	//static()
}

func static() {
	fmt.Print("Static")
}

func dynamic() {
	fmt.Print("Dynamic")
	startWebServer()
}
