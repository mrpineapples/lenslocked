package views

import (
	"html/template"
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
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}

// Render is used to render the view with the predefined layout
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	switch data.(type) {
	case Data:
		// do nothing
	default:
		data = Data{
			Yield: data,
		}
	}
	return v.Template.ExecuteTemplate(w, v.Layout, data)
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
