package views

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

var layoutDir = "views/layouts/"
var templateDir = "views/"
var templateExt = ".gohtml"

// NewView creates a View object with a template and a layout
func NewView(layout string, files ...string) *View {
	addTemplatePathAndExt(files)
	files = append(files, layoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}

// View represents an HTML view for a route
type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, nil)
}

// Render is used to render the view with the predefined layout
func (v *View) Render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	switch data.(type) {
	case Data:
		// do nothing
	default:
		data = Data{
			Yield: data,
		}
	}
	var buf bytes.Buffer
	err := v.Template.ExecuteTemplate(&buf, v.Layout, data)
	if err != nil {
		http.Error(w, "Something went wrong. If the problem persists, please contact us.", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

func layoutFiles() []string {
	files, err := filepath.Glob(layoutDir + "*" + templateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// Prepends templateDir and appends templatExt to each string
// ("home" => "views/" + "home" + ".gohtml" => "views/home.gohtml")
func addTemplatePathAndExt(files []string) {
	for i, file := range files {
		files[i] = templateDir + file + templateExt
	}
}
