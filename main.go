package main

import (
	"html/template"
	"net/http"
	"time"
	"github.com/google/uuid"
)

type Application struct {
	ID            string
	Name          string
	Customer      string
	Environment   string
	InfraRepo     string
	AppRepo       string
	HelmChart     string
	Identities    []string
	ResourceGroup string
	Notes         string
	Tags          []string
	LastUpdated   time.Time
}

var apps = map[string]Application{}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/app/view", viewAppHandler)
	http.HandleFunc("/app/edit", editAppHandler)
	http.HandleFunc("/app/update", updateAppHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	seedDummyApps()
	http.ListenAndServe(":8080", nil)
}

