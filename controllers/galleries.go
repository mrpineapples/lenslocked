package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mrpineapples/lenslocked/context"
	"github.com/mrpineapples/lenslocked/models"
	"github.com/mrpineapples/lenslocked/views"
)

const (
	ShowGalleryName = "show_gallery"
)

func NewGalleries(gs models.GalleryService, r *mux.Router) *Galleries {
	return &Galleries{
		NewView:  views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		service:  gs,
		router:   r,
	}
}

type Galleries struct {
	NewView  *views.View
	ShowView *views.View
	service  models.GalleryService
	router   *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// Show is used to show users their respective galleries.
// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return
	}

	gallery, err := g.service.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		}
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, vd)
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

	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := g.service.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	url, err := g.router.Get(ShowGalleryName).URL("id", fmt.Sprint(gallery.ID))
	if err != nil {
		// TODO: make this go to the index page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}
