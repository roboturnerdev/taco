package main

import (
	"html/template"
	"log"
	"net/http"
)

type App struct {
	Name        string
	Description string
	RepoLink    string
	Notes       string
}

var apps = []App{
	{"CoolApp", "Deploys widgets", "https://github.com/example/coolapp", "Uses MSI for auth"},
	{"InfraTool", "Bicep deployments", "https://github.com/example/infratool", "Tests in prod"},
}

func main() {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", apps)
	})

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
