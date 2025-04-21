package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"taco/internal/store"
	"taco/internal/templates"
)

type server struct {
	logger     			*log.Logger
	port      			int
	httpServer 			*http.Server
	workstreamDb  		*store.WorkstreamStore
}

func NewServer(logger *log.Logger, port int, workstreamDb *store.WorkstreamStore) (*server, error) {

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if workstreamDb == nil {
		return nil, fmt.Errorf("guestDb is required")
	}

	return &server{
		logger:  logger,
		port:    port,
		workstreamDb: workstreamDb,
	}, nil
}

func (s *server) Start() error {

	s.logger.Printf("Starting server on port %d", s.port)
	var stopChan chan os.Signal

	router := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	router.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	router.HandleFunc("GET /", s.defaultHandler)
	router.HandleFunc("GET /about", s.aboutHandler)
	router.HandleFunc("GET /health", s.healthCheckHandler)
	router.HandleFunc("GET /workstreams", s.workstreamsHandler)
	router.HandleFunc("GET /workstreams/new", s.workstreamsNewFormHandler)
	router.HandleFunc("POST /workstreams/new", s.addWorkstreamHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: router}

	stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error when running server: %s", err)
		}
	}()

	<-stopChan

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("Error when shutting down server: %v", err)
		return err
	}
	return nil
}

// GET /workstreams
func (s *server) workstreamsHandler(w http.ResponseWriter, r *http.Request) {

	workstreams, err := s.workstreamDb.GetAllWorkstreams()
	if err != nil {
		http.Error(w, "No workstreams", http.StatusInternalServerError)
		return
	}

	wsTemplate := templates.WorkstreamList(workstreams)
	err = templates.Layout(wsTemplate, "TACO", "/workstreams").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering workstreams: %v", err)
	}
}

// GET /workstreams/new - Render the form to create a new workstream
func (s *server) workstreamsNewFormHandler(w http.ResponseWriter, r *http.Request) {

	newWorkstreamTemplate := templates.NewWorkstreamForm()
	err := templates.Layout(newWorkstreamTemplate, "New Workstream", "/workstreams/new").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering new workstream form: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// POST /workstreams/new - Try add to database
func (s *server) addWorkstreamHandler(w http.ResponseWriter, r *http.Request) {

	// do not do this unless we are posting
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Workstream name is required", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Workstream code is required", http.StatusBadRequest)
		return
	}

	location := r.FormValue("location")
	if location == "" {
		http.Error(w, "Workstream location is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "Workstream description is required", http.StatusBadRequest)
		return
	}

	quote := r.FormValue("quote")
	if quote == "" {
		http.Error(w, "Workstream quote is required", http.StatusBadRequest)
		return
	}
	
	workstream := store.Workstream{
		Name:			name,
		Code: 			code,
		Location: 		location,
		Description: 	description,
		Quote: 			quote,
	}
	err = s.workstreamDb.CreateWorkstream(workstream)
	if err != nil {
		http.Error(w, "Failed to create workstream", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workstreams", http.StatusFound)
}

// GET /
func (s *server) defaultHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	homeTemplate := templates.Home()
	err := templates.Layout(homeTemplate, "TACO", "/").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering home: %v", err)
	}
}

// GET /about
func (s *server) aboutHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	aboutTemplate := templates.About()
	err := templates.Layout(aboutTemplate, "About", "/about").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering about: %v", err)
	}
}

// GET /health - HealthCheckHandler is a simple handler to check the health of the server
func (s *server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}