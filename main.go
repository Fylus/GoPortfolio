/*
 This application starts a web server or serves it as static pages.
It uses a zip file with the json files and the static files.
The zip file is loaded from the input folder and the static pages are saved to the output folder.
The json files are saved to the json folder and then loaded into the database.
*/
package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	statDir = "./static"
	jsonDir = "./json"
)

var (
	outputDir = "output"
	buildDir  = outputDir + "/webapp_build"
	// check if static build is requested
	st       = os.Getenv("BUILD_STATIC") == "1"
	inputDir = "input"
)

// main is the entry point for the application.
func main() {
	log.Println("Starting application")

	checkInputFolder()
	loadZip()
	buildDatabase()

	// check if static build is requested and build static pages or start web server
	if st {
		static()
	} else {
		dynamic()
	}
}

// checkInputFolder checks if input folder exists. It's needed for the zip file
func checkInputFolder() {
	log.Println("Input folder: ", inputDir)
	//print all input files
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalln("Error reading input folder: ", err)
	}
	for _, f := range files {
		log.Println("Input file: ", f.Name())
	}
}

// loadZip loads the zip file from the input folder and extracts the json files to the json folder
func loadZip() {
	var zipName string

	// check the environment variable for the zip file name
	custom := os.Getenv("ZIP_NAME")
	if custom != "" {
		zipName = custom
	} else {
		zipName = "resources.zip"
	}

	path := inputDir + "/" + zipName
	log.Println("Opening zip file: ", path)

	// open a zip archive for reading
	r, err := zip.OpenReader(path)
	if err != nil {
		log.Fatalln("Error opening zip file: ", err)
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Fatalln("Error closing zip file: ", err)
		}
	}(r)

	//check if jsonDir exists otherwise create it
	if _, err := os.Stat(jsonDir); os.IsNotExist(err) {
		err := os.Mkdir(jsonDir, os.ModePerm)
		if err != nil {
			log.Fatalln("Error creating jsonDir: ", err)
		}
	}

	// Iterate through the files in the archive
	for _, f := range r.File {
		// Check if the current file is in the json folder copy it to the json folder
		if filepath.Dir(f.Name) == "json" && !f.FileInfo().IsDir() {
			copyZipFile(f, filepath.Join(jsonDir, filepath.Base(f.Name)))
		}

		// Check if the current file is in the static folder and copy it to the static folder
		if strings.HasPrefix(f.Name, "images/") && f.Name != "images/" {
			path := filepath.Join(statDir, f.Name)
			if f.FileInfo().IsDir() {
				err := os.MkdirAll(path, f.Mode())
				if err != nil {
					log.Fatalln("Error creating imageDir: ", err)
				}
			} else {
				copyZipFile(f, filepath.Join(statDir, f.Name))
			}
		}

	}
}

// copyZipFile copies a file from the zip archive to the given path
func copyZipFile(f *zip.File, path string) {
	// Open the current file in the zip file
	rc, err := f.Open()
	if err != nil {
		log.Fatalln("Error opening file in zip: ", err)
	}
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			log.Fatalln("Error closing file in zip: ", err)
		}
	}(rc)

	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		log.Fatalln("Error creating file in json folder: ", err)
	}
	defer func(fw *os.File) {
		err := fw.Close()
		if err != nil {
			log.Fatalln("Error closing file in json folder: ", err)
		}
	}(fw)

	// Write the contents of the file to the target file
	_, err = io.Copy(fw, rc)
	if err != nil {
		log.Fatalln(err)
	}
}

// static builds the static pages and saves them to the output folder
func static() {
	log.Println("Static build")

	//delete buildDir if exists
	if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
		err = os.RemoveAll(buildDir)
		if err != nil {
			log.Fatalln("Error deleting buildDir: ", err)
		}
	}

	log.Println("Output folder: ", outputDir)
	renderStaticPages()
	log.Println("Static build complete")
}

// startWebServer starts the web server
func dynamic() {
	log.Println("Dynamic start")
	startWebServer()
}
