package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mrpineapples/lenslocked/models"
	"github.com/mrpineapples/lenslocked/views"
)

func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		NewView: views.NewView("bootstrap", "galleries/new"),
		service: gs,
	}
}

type Galleries struct {
	NewView *views.View
	service models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// Create is used to process the gallery form and creates a new gallery.
// POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		Title: form.Title,
	}
	if err := g.service.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	fmt.Fprintln(w, gallery)
}
