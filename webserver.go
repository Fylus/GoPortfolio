/*
 This file is the webbserver to serve the portfolio website on a specified port.
 The webserver is started by running the main.go file.
 The webserver uses the gin web framework.
 The webserver uses the data structures defined in webdata.go.
*/
package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	//srcDir    = "./seiten" // Verzeichnis für Blog -Beiträge
	tmplDir      = "./templates/" // HTML -Template Verzeichnis
	templFile    = "*.templ.html"
	homeTempl    = "home"
	productTempl = "product"
	impTempl     = "impressum"
	errorTempl   = "error"
)

// startWebserver starts the webserver on the specified port and sets up the routes
func startWebServer() {
	router := gin.Default()
	log.Println("Load templates from: ", tmplDir)
	router.LoadHTMLGlob(filepath.Join(tmplDir, "**/", templFile))
	log.Println("Load static files from: ", statDir)
	router.Static("/static", statDir)
	log.Println("Set up routes")
	router.NoRoute(pageNotFound)
	router.GET("/", homeHandler)
	router.GET("/impressum", impressumHandler)
	router.GET("/project/:projectID", projectHandler)
	router.GET("/tool/:toolID", toolHandler)
	port := ":" + os.Getenv("PORT")
	log.Printf("Listening on :%v ....", port)
	err := router.Run(port)
	if err != nil {
		log.Fatalln("Error starting web server: ", err)
	}
}

// toolHandler handles the request for a tool page, used from software-sites
func toolHandler(context *gin.Context) {
	tool, status := getToolFromDatabase(context.Param("toolID"))
	context.HTML(status, productTempl, tool)
}

// projectHandler handles the request for a project page, used from project-sites
func projectHandler(context *gin.Context) {
	product, status := getProjectFromDatabase(context.Param("projectID"))
	context.HTML(status, productTempl, product)
}

// impressumHandler handles the request for the impressum page
func impressumHandler(context *gin.Context) {
	ps := impressumData()
	context.HTML(http.StatusOK, impTempl, ps)
}

// pageNotFound handles the request for a page that does not exist
func pageNotFound(c *gin.Context) {
	ps := Page{Title: "Page not found"}
	c.HTML(http.StatusNotFound, errorTempl, ps)
}

// homeHandler handles the request for the home page
func homeHandler(c *gin.Context) {
	home := homeData()
	c.HTML(http.StatusOK, homeTempl, home)
}
