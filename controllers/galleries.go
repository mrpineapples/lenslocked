package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mrpineapples/lenslocked/context"
	"github.com/mrpineapples/lenslocked/models"
	"github.com/mrpineapples/lenslocked/views"
)

// Named routes
const (
	ShowGalleryName = "show_gallery"
	EditGalleryName = "edit_gallery"

	maxMultipartMem = 5 << 20 // 5 megabytes
)

func NewGalleries(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		IndexView:  views.NewView("bootstrap", "galleries/index"),
		NewView:    views.NewView("bootstrap", "galleries/new"),
		ShowView:   views.NewView("bootstrap", "galleries/show"),
		EditView:   views.NewView("bootstrap", "galleries/edit"),
		service:    gs,
		imgService: is,
		router:     r,
	}
}

type Galleries struct {
	IndexView  *views.View
	NewView    *views.View
	ShowView   *views.View
	EditView   *views.View
	service    models.GalleryService
	imgService models.ImageService
	router     *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// Index renders the default view where a user can view all their galleries.
// GET /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.service.ByUserID(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// Show is used to show users their respective galleries.
// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

// Edit is used to show the user a form which they can use to edit a gallery.
// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)
}

// Update is used to update a gallery with the data from the edit form.
// POST /galleries/:id/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Title = form.Title
	err = g.service.Update(gallery)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLevelSuccess,
		Message: "Gallery successfully updated!",
	}
	g.EditView.Render(w, r, vd)
}

// Create is used to process the gallery form and creates a new gallery.
// POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, r, vd)
		return
	}

	user := context.User(r.Context())
	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := g.service.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, r, vd)
		return
	}

	url, err := g.router.Get(EditGalleryName).URL("id", fmt.Sprint(gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// ImageUpload is used to upload images.
// POST /galleries/:id/images
func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		// open the uploaded file
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()

		err = g.imgService.Create(gallery.ID, file, f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}

	url, err := g.router.Get(EditGalleryName).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

// ImageViaLink is used to upload images via a link.
// POST /galleries/:id/images/link
func (g *Galleries) ImageViaLink(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yield = gallery

	if err := r.ParseForm(); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	files := r.PostForm["files"]

	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, fileURL := range files {
		go func(imgURL string) {
			defer wg.Done()
			resp, err := http.Get(imgURL)
			if err != nil {
				log.Println("Failed to download the image from:", imgURL)
			}
			defer resp.Body.Close()
			sections := strings.Split(imgURL, "/")
			filename := sections[len(sections)-1]
			decodedFile, err := url.QueryUnescape(filename)
			if err != nil {
				log.Fatal(err)
				return
			}
			if err := g.imgService.Create(gallery.ID, resp.Body, decodedFile); err != nil {
				log.Println("Failed to created the image from:", imgURL)
			}
		}(fileURL)
	}
	wg.Wait()

	url, err := g.router.Get(EditGalleryName).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// ImageDelete is used to delete individual images in a gallery.
// POST /galleries/:id/images/:filename/delete
func (g *Galleries) ImageDelete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	filename := mux.Vars(r)["filename"]
	decodedFile, err := url.QueryUnescape(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	img := models.Image{
		Filename:  decodedFile,
		GalleryID: gallery.ID,
	}

	err = g.imgService.Delete(&img)
	if err != nil {
		var vd views.Data
		vd.Yield = gallery
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.router.Get(EditGalleryName).URL("id", fmt.Sprint(gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// Delete is used to delete a gallery by its ID.
// POST /galleries/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	err = g.service.Delete(gallery.ID)
	if err != nil {
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.service.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}

	images, _ := g.imgService.ByGalleryID(gallery.ID)
	gallery.Images = images

	return gallery, nil
}
