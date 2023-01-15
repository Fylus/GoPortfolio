/*
 	This file contains all the data structures used in the project.
	webserver.go and webbuilder.go use these data structures to render the webpages.
*/
package main

import "go.mongodb.org/mongo-driver/bson"

// Page data structure for the header and footer of every page
type Page struct {
	Title string
	CSS   string
	HTML  string
}

// Home data structure for the home page
type Home struct {
	Page
	Categories  map[string][]bson.M
	Education   []bson.M
	ProgLang    []bson.M
	Software    []bson.M
	OtherSkills []bson.M
	Languages   []bson.M
}

// ProductPage data structure for the project and tool pages
type ProductPage struct {
	Page
	Description string
	Image       string
	Table       map[string][]bson.M
	Type        string
	External    string
	Noproduct   bool
}

// getHTML returns the HTML returns the suffix for HTML-links if the pages are served statically
func getHTML() string {
	html := ""
	if st {
		html = ".html"
	}
	return html
}

// HomeData returns the data for the home page using the database
func homeData() Home {
	home := Home{
		Page: Page{
			Title: "Portfolio",
			HTML:  getHTML(),
			CSS:   "home",
		},
		Categories:  getAllProjectsInCategories(),
		Education:   getEducationFromDatabase(),
		ProgLang:    getProgLangFromDatabase(),
		Software:    getSoftwareFromDatabase(),
		OtherSkills: getOtherSkillsFromDatabase(),
		Languages:   getLanguageFromDatabase(),
	}
	return home
}

// impressumData returns the data for the impressum page
func impressumData() Page {
	return Page{
		Title: "Impressum",
		HTML:  getHTML(),
	}
}
