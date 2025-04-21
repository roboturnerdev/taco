package main

import (
	"log"
	"os"

	"taco/internal/server"
	"taco/internal/store"
)

func main() {

	logger := log.New(os.Stdout, "[TACO] ", log.LstdFlags)

	port := 9001

	logger.Print("Creating workstream store..")
	workstreamDb, err := store.NewWorkstreamStore("workstreams.db")
	if err != nil {
		logger.Fatalf("Error when creating workstream store: %s", err)
		os.Exit(1)
	}

	err = workstreamDb.CreateWorkstream(store.Workstream{
		Name:		 "Enclave",
		Code:		 "Bicep / Powershell",
		Location:	 "Azure",
		Description: "Three Word Acronym",
		Quote:		 "Managing the accredited boundary like no other; The Gold Standard.",
	})
	if err != nil {
		logger.Printf("Error adding sample workstream: %s", err)
	}

	logger.Print("Creating guest store..")
	guestDb := store.NewGuestStore(logger)
	guestDb.AddGuest(store.Guest{Name: "Rob", Email: "rob@taco.dev"})

	srv, err := server.NewServer(logger, port, guestDb, workstreamDb)
	if err != nil {
		logger.Fatalf("Error when creating server: %s", err)
		os.Exit(1)
	}
	if err := srv.Start(); err != nil {
		logger.Fatalf("Error when starting server: %s", err)
		os.Exit(1)
	}
}