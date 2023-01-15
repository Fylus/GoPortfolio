/*
 This file contains all building functions to generate the static website

*/
package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// renderStaticPages renders all static pages and writes them to the buildDir
// it is called by main.go
func renderStaticPages() {
	// Parse and compile the templates
	tmpl := template.Must(template.ParseGlob("templates/**/*.templ.html"))

	//copy static files
	err := copyDir("static", buildDir+"/static")
	if err != nil {
		log.Fatalln("Error copying static files: ", err)
	}

	// generate pages
	generatePage(tmpl.Lookup("home"), homeData(), "index.html")
	generatePage(tmpl.Lookup("impressum"), impressumData(), "impressum.html")
	err = generateProductpages(projects, "project", tmpl)
	if err != nil {
		log.Fatalln("Error generating project pages: ", err)
	}
	err = generateProductpages(software, "tool", tmpl)
	if err != nil {
		log.Fatalln("Error generating tool pages: ", err)
	}
}

// generateProductpages generates all product pages of the category projects or software
func generateProductpages(category string, folder string, tmpl *template.Template) error {
	productIDs := getAllIDs(category)
	if len(productIDs) > 0 {
		// make project folder if doesn't exist
		if _, err := os.Stat(buildDir + "/" + folder); os.IsNotExist(err) {
			err := os.Mkdir(buildDir+"/"+folder, 0755)
			if err != nil {
				return err
			}
		}
		for _, productID := range productIDs {
			var page ProductPage
			if category == projects {
				page, _ = getProjectFromDatabase(productID)
			} else if category == software {
				page, _ = getToolFromDatabase(productID)
			} else {
				break
			}
			generatePage(tmpl.Lookup("product"), page, folder+"/"+productID+".html")
		}
	}
	return nil
}

// copyDir copies a directory recursively from src directory to dst directory
func copyDir(src string, dst string) error {

	sInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	_, err = os.Stat(dst)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dst, sInfo.Mode())
		if err != nil {
			return err
		}
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a file from src directory to dst directory
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		err := in.Close()
		if err != nil {
			return
		}
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			return
		}
	}(out)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

// generatePage generates a single page and writes it to the buildDir using the given template and data and the given filename
func generatePage(tmpl *template.Template, s interface{}, path string) {
	log.Println("Generating page: " + path)
	f, err := os.Create(buildDir + "/" + path)
	if err != nil {
		log.Fatalln("Error creating file: ", err)
	}
	err = tmpl.Execute(f, s)
	if err != nil {
		log.Fatalln("Error executing template: ", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalln("Error closing file: ", err)
	}
}
