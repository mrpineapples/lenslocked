package controllers

import (
	"net/http"

	"github.com/mrpineapples/lenslocked/views"
)

// NewUsers is used to create a new Users controller.
// It will panic if templates are not parsed correctly
// and should only be used during setup
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

type Users struct {
	NewView *views.View
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}
