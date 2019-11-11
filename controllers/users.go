package controllers

import (
	"fmt"
	"net/http"

	"github.com/mrpineapples/lenslocked/models"
	"github.com/mrpineapples/lenslocked/views"
)

// NewUsers is used to create a new Users controller.
// It will panic if templates are not parsed correctly
// and should only be used during setup
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		service: us,
	}
}

type Users struct {
	NewView *views.View
	service *models.UserService
}

// New is used to render the form where a user can create a new account
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form and creates a new account
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	user := models.User{
		Name:  form.Name,
		Email: form.Email,
	}
	if err := u.service.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, form)
}
