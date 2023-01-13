package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"path/filepath"
)

var (
	tmpDir = flag.String("tmp", "templates", "Template -Dir.")
)

const (
	statDir = "./static" // statische Seiten: html , css , js , etc.
	//srcDir    = "./seiten" // Verzeichnis für Blog -Beiträge
	tmplDir      = "./templates/" // HTML -Template Verzeichnis
	templFile    = "*.templ.html"
	homeTempl    = "home"
	productTempl = "product"
	impTempl     = "impressum"
	errorTempl   = "error"
)

func startWebServer() {
	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join(tmplDir, "**/", templFile))
	router.Static("/static", statDir)
	router.NoRoute(pageNotFound)
	router.GET("/", makeHomeHandler)
	router.GET("/impressum", impressumHandler)
	router.GET("/project/:projectID", projectHandler)
	router.GET("/tool/:toolID", toolHandler)
	log.Print("Listening on :9000 ....")
	err := router.Run(":9000")
	if err != nil {
		log.Fatal("Error starting web server: ", err)
	}
}

func toolHandler(context *gin.Context) {
	tool, status := getToolFromDatabase(context.Param("toolID"))
	context.HTML(status, productTempl, tool)
}

func projectHandler(context *gin.Context) {
	product, status := getProjectFromDatabase(context.Param("projectID"))
	context.HTML(status, productTempl, product)
}

func impressumHandler(context *gin.Context) {
	ps := Page{Title: "Impressum"}
	context.HTML(http.StatusOK, impTempl, ps)
}

func pageNotFound(c *gin.Context) {
	ps := Page{Title: "Page not found"}
	c.HTML(http.StatusNotFound, errorTempl, ps)
}

func makeHomeHandler(c *gin.Context) {
	type Home struct {
		Page
		Categories  map[string][]bson.M
		Education   []bson.M
		ProgLang    []bson.M
		Software    []bson.M
		OtherSkills []bson.M
		Languages   []bson.M
	}
	getSkillsFromDatabase()

	home := Home{
		Page: Page{
			Title: "Portfolio",
			CSS:   "home",
		},
		Categories:  getAllProjectsInCategories(),
		Education:   getEducationFromDatabase(),
		ProgLang:    getProgLangFromDatabase(),
		Software:    getSoftwareFromDatabase(),
		OtherSkills: getOtherSkillsFromDatabase(),
		Languages:   getLanguageFromDatabase(),
	}
	// fill struct Home
	c.HTML(http.StatusOK, homeTempl, home)
}

type Page struct {
	Title string
	CSS   string
	Product
}

type Product struct {
	Description string
	Image       string
	Table       map[string][]bson.M
	Type        string
	External    string
	Noproduct   bool
}

type TableMap struct {
	Name    string
	Content []string
}
