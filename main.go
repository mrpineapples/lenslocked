package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	path := r.URL.Path
	if path == "/" {
		fmt.Fprint(w, "<h1>Welcome to my awesome site &#128293;</h1>")
	} else if path == "/contact" {
		fmt.Fprint(w, "To get in touch, please send an email to <a href=\"mailto:support@lenslocked.com\">support@lenslocked.com</a>.")
	}
}

func main() {
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":8000", nil)
}
