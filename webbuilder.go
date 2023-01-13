package main

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

func renderPage(w io.Writer) error {
	data := Page{
		Title: "Home",
	}
	temp, err := template.ParseFiles(
		filepath.Join(*tmpDir, "base.templ.html"),
		filepath.Join(*tmpDir, "head.templ.html"),
		filepath.Join(*tmpDir, "header.templ.html"),
		filepath.Join(*tmpDir, "content.templ.html"),
		filepath.Join(*tmpDir, "links.templ.html"),
		filepath.Join(*tmpDir, "footer.templ.html"),
		//filepath.Join(*tmpDir, content),
	)
	if err != nil {
		return fmt.Errorf("renderPage.Parsefiles: %w", err)
	}
	err = temp.ExecuteTemplate(w, "base", data)
	if err != nil {
		return fmt.Errorf("renderPage.ExecuteTemplate: %w", err)
	}
	return nil
}
