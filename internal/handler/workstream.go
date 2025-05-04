package handler

import (
	"log"
	"net/http"
	"strconv"
	"taco/internal/models"
	"taco/internal/store"
	"taco/internal/templates"

	"github.com/go-chi/chi"
)

type WorkstreamHandler struct {
	Logger 	*log.Logger
	DB		store.WorkstreamReader
}

func NewWorkstreamHandler(logger *log.Logger, db store.WorkstreamReader) *WorkstreamHandler {

	return &WorkstreamHandler{
		Logger: logger,
		DB: db,
	}
}

// GET /workstreams
func (h *WorkstreamHandler) List(w http.ResponseWriter, r *http.Request) {

	workstreams, err := h.DB.GetAllWorkstreams()
	if err != nil {
		http.Error(w, "No workstreams", http.StatusInternalServerError)
		return
	}

	if err := templates.Layout(templates.WorkstreamList(workstreams), "TACO", "/workstreams").Render(r.Context(), w); err != nil {
		h.Logger.Printf("Template error when rendering workstreams: %v", err)
	}
}

// GET /workstreams/new
func (h *WorkstreamHandler) CreateNewGet(w http.ResponseWriter, r *http.Request) {

	if err := templates.Layout(templates.NewWorkstreamForm(), "New Workstream", "/workstreams/new").
		Render(r.Context(), w); err != nil {
		h.Logger.Printf("Error when rendering new workstream form: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// GET /workstreams/{id}
func (h *WorkstreamHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr) // ensure id is converted to integer
	if err != nil {
		http.NotFound(w, r)
		return
	}

	workstream, err := h.DB.GetWorkstreamByID(id)
	if err != nil {
		http.Error(w, "Workstream not found", http.StatusNotFound)
		return
	}

	workstreamPath := "/workstreams/" + strconv.Itoa(workstream.ID)
	err = templates.Layout(templates.WorkstreamDetailPage(workstream), "TACO", workstreamPath).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// POST /workstreams/new - Try add to database
func (h *WorkstreamHandler) CreateNewPost(w http.ResponseWriter, r *http.Request) {

	// do not do this unless we are posting
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if errParse := r.ParseForm(); errParse != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// name is the only value in the database that is required / NOT NULL
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Workstream name is required", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	location := r.FormValue("location")
	description := r.FormValue("description")
	identity := r.FormValue("identity")
	quote := r.FormValue("quote")

	workstream := models.Workstream{
		Name:        name,
		Code:        code,
		Location:    location,
		Description: description,
		Identity:    identity,
		Quote:       quote,
	}

	if err := h.DB.CreateWorkstream(workstream); err != nil {
		http.Error(w, "Failed to create workstream", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workstreams", http.StatusFound)
}

// POST /workstreams/{id}/delete
func (h *WorkstreamHandler) Delete(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, idErr := strconv.Atoi(idStr)
	if idErr != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.DB.DeleteWorkstream(id); err != nil {
		http.Error(w, "Failed to delete workstream", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workstreams", http.StatusSeeOther)
}