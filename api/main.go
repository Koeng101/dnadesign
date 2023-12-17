package main

import (
	"log"
	"net/http"
	"os"

	"github.com/koeng101/dnadesign/api/api"
)

func main() {
	app := api.InitializeApp()
	// Serve application
	s := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: app.Router,
	}
	log.Fatal(s.ListenAndServe())
}
