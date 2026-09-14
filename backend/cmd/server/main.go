package main

import (
	"log"
	"net/http"
	"os"

	"media-sequencer/backend/internal/handler"
	"media-sequencer/backend/internal/middleware"
	"media-sequencer/backend/internal/service"
	"media-sequencer/backend/internal/store"
)

func main() {
	dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		dbPath = "./data/media_sequencer.db"
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	database, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	defer database.Close()

	appService := service.New(database)
	appHandler := handler.New(appService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", appHandler.Health)
	mux.HandleFunc("/windows", appHandler.GetWindows)
	mux.HandleFunc("/windows/", appHandler.AddMedia)
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			appHandler.SyncMedia(w, r)
			return
		}

		if r.Method == http.MethodGet {
			appHandler.GetSync(w, r)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: middleware.CORS(mux),
	}

	log.Printf("Media Sequencer backend running on port %s", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}