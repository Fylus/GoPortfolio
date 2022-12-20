package main

import (
	"fmt"
)

func main() {
	//st := flag.Bool("static", false, "Static")
	//flag.Parse()
	//
	//if *st {
	//	static()
	//} else {
	//	dynamic()
	//}

	buildDatabase()
	project := Project{Date: 2345}
	fmt.Println("Project ", project)

}

func static() {
	fmt.Print("Static")
}

func dynamic() {
	fmt.Print("Dynamic")
	//connect with mongodb

}
