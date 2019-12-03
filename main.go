package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrpineapples/lenslocked/controllers"
	"github.com/mrpineapples/lenslocked/middleware"
	"github.com/mrpineapples/lenslocked/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "Michael"
	password = "not-necessary"
	dbname   = "lenslocked_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	// declare router first so controllers can use it
	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, r)

	// lint can be ignored for middleware
	requireUserMw := middleware.RequireUser{services.User}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/signup", usersC.NewView).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// Gallery routes
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesC.NewView)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGalleryName)

	fmt.Println("Server running on port 8000 visit: http://localhost:8000/")
	http.ListenAndServe(":8000", r)
}
