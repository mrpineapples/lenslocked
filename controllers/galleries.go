package controllers

import (
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
