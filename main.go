package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrpineapples/lenslocked/controllers"
	"github.com/mrpineapples/lenslocked/views"
)

var homeView *views.View
var contactView *views.View

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

func main() {
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	usersC := controllers.NewUsers()

	r := mux.NewRouter()
	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/contact", contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	http.ListenAndServe(":8000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
