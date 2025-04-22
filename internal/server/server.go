package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"taco/internal/store"
	"taco/internal/templates"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
		return nil, fmt.Errorf("workstreamDb is required")
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

	// chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.Get("/", s.homeHandler)
	r.Get("/about", s.aboutHandler)
	r.Get("/health", s.healthCheckHandler)

	r.Route("/workstreams", func(r chi.Router) {
		r.Get("/", s.workstreamsHandler)
		r.Get("/new", s.workstreamsNewHandler)
		r.Post("/new", s.workstreamsPostNewHandler)
		r.Get("/{id}", s.workstreamIdHandler)
	})
	
	s.httpServer = &http.Server{
		Addr:			fmt.Sprintf(":%d", s.port),
		Handler:		r,
		ReadTimeout:	5 * time.Second,
		WriteTimeout:	10 * time.Second,
		IdleTimeout:	15 * time.Second,
	}

	stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Goroutine
	// httpServer.ListenAndServe() blocks the process
	// putting it in a goroutine it prevents it from blocking shutdown logic
	// if the server crashes for any reason other than being manually shutdown it logs fatal
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error when running server: %s", err)
		}
	}()

	// the main function blocks here until a signal is received on stopChan
	<-stopChan
	s.logger.Println("Shutting down [TACO] server...")

	// This example of Go works to still be readable and explicit
	// There is only "error" coming back from the function
	// The lifecycle of err here is only the if block
	// Consideration: variables inside if blocks are scoped to it, and lost after
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("Error when shutting down server: %v", err)
		return err
	}
	return nil
}

// GET /workstreams
func (s *server) workstreamsHandler(w http.ResponseWriter, r *http.Request) {

	// Try to avoid mixing control flow and success logic inside a conditional block like this
	// It is against Go's usual clean separation of logic.
	// The lifecycle of workstreams here is beyond the error handling for the db call
	workstreams, err := s.workstreamDb.GetAllWorkstreams()
	if err != nil {
		http.Error(w, "No workstreams", http.StatusInternalServerError)
		return
	}

	err = templates.Layout(templates.WorkstreamList(workstreams), "TACO", "/workstreams").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering workstreams: %v", err)
	}
}

// GET /workstreams/new - Render the form to create a new workstream
func (s *server) workstreamsNewHandler(w http.ResponseWriter, r *http.Request) {

	err := templates.Layout(templates.NewWorkstreamForm(), "New Workstream", "/workstreams/new").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering new workstream form: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}


// POST /workstreams/new - Try add to database
func (s *server) workstreamsPostNewHandler(w http.ResponseWriter, r *http.Request) {

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
	location := r.FormValue("location")
	description := r.FormValue("description")
	identity := r.FormValue("identity")
	quote := r.FormValue("quote")
	
	workstream := store.Workstream{
		Name:			name,
		Code: 			code,
		Location: 		location,
		Description: 	description,
		Identity:		identity,
		Quote: 			quote,
	}
	err = s.workstreamDb.CreateWorkstream(workstream)
	if err != nil {
		http.Error(w, "Failed to create workstream", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workstreams", http.StatusFound)
}

// GET /workstreams/{id}
func (s *server) workstreamIdHandler(w http.ResponseWriter, r *http.Request) {
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)	// ensure id is converted to integer
	if err != nil {
		http.NotFound(w, r)
		return
	}

	workstream, err := s.workstreamDb.GetWorkstreamByID(id)
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

// GET /
func (s *server) homeHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	err := templates.Layout(templates.Home(), "TACO", "/").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering home: %v", err)
	}
}

// GET /about
func (s *server) aboutHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	err := templates.Layout(templates.About(), "About", "/about").Render(r.Context(), w)
	if err != nil {
		s.logger.Printf("Error when rendering about: %v", err)
	}
}

// GET /health - HealthCheckHandler is a simple handler to check the health of the server
func (s *server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}