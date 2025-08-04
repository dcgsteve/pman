package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/steve/pman/backend/database"
	"github.com/steve/pman/backend/handlers"
	"github.com/steve/pman/shared/config"
)

var (
	Version = "1.0.1"
)

func main() {
	if err := config.ValidateEnvVars(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()
	handlers.SetupRoutes(api, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("pman backend API listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
