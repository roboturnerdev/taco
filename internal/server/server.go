package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taco/internal/handler"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type server struct {
	logger       		*log.Logger
	port         		int
	httpServer   		*http.Server
	workstreamHandler 	*handler.WorkstreamHandler
	pageHandler			*handler.PageHandler
}

func NewServer(logger *log.Logger, port int, handler *handler.WorkstreamHandler, ph *handler.PageHandler) (*server, error) {

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if handler == nil || ph == nil {
		return nil, fmt.Errorf("handlers are required")
	}

	return &server{
		logger:       		logger,
		port:         		port,
		workstreamHandler: 	handler,
		pageHandler: 		ph,
	}, nil
}

func (s *server) Start() error {

	s.logger.Printf("Starting server on port %d", s.port)

	var stopChan chan os.Signal

	// chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// page handler
	r.Get("/", s.pageHandler.Home)
	r.Get("/about", s.pageHandler.About)
	r.Get("/health", s.pageHandler.Health)

	// workstream handler
	r.Route("/workstreams", func(r chi.Router) {
		r.Get("/", s.workstreamHandler.List)
		r.Get("/new", s.workstreamHandler.CreateNewGet)
		r.Post("/new", s.workstreamHandler.CreateNewPost)
		r.Get("/{id}", s.workstreamHandler.GetByID)
		r.Post("/{id}/delete", s.workstreamHandler.Delete)
	})

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
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

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("Error when shutting down server: %v", err)
		return err
	}
	return nil
}