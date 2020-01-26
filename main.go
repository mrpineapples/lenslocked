package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	llctx "github.com/mrpineapples/lenslocked/context"
	"github.com/mrpineapples/lenslocked/controllers"
	"github.com/mrpineapples/lenslocked/email"
	"github.com/mrpineapples/lenslocked/middleware"
	"github.com/mrpineapples/lenslocked/models"
	"github.com/mrpineapples/lenslocked/rand"
	"golang.org/x/oauth2"
)

func main() {
	boolPtr := flag.Bool("prod", false, "Provide this flag in production. This ensures that a .config file is provided to the application")
	flag.Parse()

	appConfig := LoadConfig(*boolPtr)
	dbConfig := appConfig.Database
	services, err := models.NewServices(
		models.WithGorm(dbConfig.Dialect(), dbConfig.ConnectionInfo()),
		models.WithLogMode(!appConfig.IsProd()),
		models.WithUser(appConfig.Pepper, appConfig.HMACKey),
		models.WithGallery(),
		models.WithImage(),
		models.WithOAuth(),
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	mgConfig := appConfig.Mailgun
	emailer := email.NewClient(
		email.WithSender("lens-locked support", "support@lens-locked.com"),
		email.WithMailgun(mgConfig.Domain, mgConfig.APIKey, mgConfig.PublicAPIKey),
	)

	// declare router first so controllers can use it
	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User, emailer)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMw := csrf.Protect(b, csrf.Secure(appConfig.IsProd()))
	// lint can be ignored for middleware
	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{userMw}

	dbxOAuth := &oauth2.Config{
		ClientID:     appConfig.Dropbox.ID,
		ClientSecret: appConfig.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  appConfig.Dropbox.AuthURL,
			TokenURL: appConfig.Dropbox.TokenURL,
		},
		RedirectURL: "http://localhost:8000/oauth/dropbox/callback",
	}

	dbxRedirect := func(w http.ResponseWriter, r *http.Request) {
		state := csrf.Token(r)
		cookie := http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
		url := dbxOAuth.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}
	r.HandleFunc("/oauth/dropbox/connect", requireUserMw.ApplyFn(dbxRedirect))
	dbxCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.FormValue("state")
		cookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if cookie == nil || cookie.Value != state {
			http.Error(w, "Invalid state provided", http.StatusBadRequest)
			return
		}
		cookie.Value = ""
		cookie.Expires = time.Now()
		http.SetCookie(w, cookie)

		code := r.FormValue("code")
		token, err := dbxOAuth.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user := llctx.User(r.Context())
		existingToken, err := services.OAuth.Find(user.ID, models.OAuthDropbox)
		if err == models.ErrNotFound {
			// noop
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			services.OAuth.Delete(existingToken.ID)
		}
		userOAuth := models.OAuth{
			UserID:  user.ID,
			Token:   *token,
			Service: models.OAuthDropbox,
		}
		err = services.OAuth.Create(&userOAuth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%+v", token)
	}
	r.HandleFunc("/oauth/dropbox/callback", requireUserMw.ApplyFn(dbxCallback))

	dbxQuery := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		path := r.FormValue("path")
		user := llctx.User(r.Context())
		userOAuth, err := services.OAuth.Find(user.ID, models.OAuthDropbox)
		if err != nil {
			panic(err)
		}
		token := userOAuth.Token

		data := struct {
			Path string `json:"path"`
		}{
			Path: path,
		}
		dataBytes, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		client := dbxOAuth.Client(context.TODO(), &token)
		req, err := http.NewRequest(http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", bytes.NewReader(dataBytes))
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	}
	r.HandleFunc("/oauth/dropbox/test", requireUserMw.ApplyFn(dbxQuery))

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/logout", requireUserMw.ApplyFn(usersC.Logout)).Methods("POST")
	r.Handle("/forgot", usersC.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", usersC.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", usersC.ResetPw).Methods("GET")
	r.HandleFunc("/reset", usersC.CompleteReset).Methods("POST")

	// Assets routes
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Gallery routes
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesC.NewView)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGalleryName)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).Methods("GET").Name(controllers.EditGalleryName)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleriesC.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleriesC.ImageUpload)).Methods("POST")
	// route to delete individual images
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleriesC.ImageDelete)).Methods("POST")

	fmt.Printf("Server running on port %[1]d visit: http://localhost:%[1]d/\n", appConfig.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Port), csrfMw(userMw.Apply(r)))
}
