package handler

import (
	"log"
	"net/http"
	"taco/internal/templates"
)

type PageHandler struct {
	Logger *log.Logger
}

func NewPageHandler(logger *log.Logger) *PageHandler {
	
	return &PageHandler{
		Logger: logger,
	}
}

// GET / - Home page
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	if err := templates.Layout(templates.Home(), "TACO", "/").
		Render(r.Context(), w); err != nil {
		h.Logger.Printf("Error when rendering home: %v", err)
	}
}

// GET /about - About page
func (h *PageHandler) About(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	if err := templates.Layout(templates.About(), "About", "/about").
		Render(r.Context(), w); err != nil {
		h.Logger.Printf("Error when rendering about: %v", err)
	}
}

// GET /health - HealthCheckHandler is a simple handler to check the health of the server
func (h *PageHandler) Health(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("we up"))
}