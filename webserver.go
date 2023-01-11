package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"html/template"
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
	//router.GET("/impressum/:topic", blogHandler)
	log.Print("Listening on :9000 ....")
	err := router.Run(":9000")
	if err != nil {
		log.Fatal("Error starting web server: ", err)
	}
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
	ps := Page{Title: "Home", CSS: "home"}
	c.HTML(http.StatusNotFound, homeTempl, ps)
}

type Page struct {
	Title   string
	CSS     string
	Content template.HTML
}

type Product struct {
	Title   string
	Content template.HTML
	Image   string
	Table   []string
}
