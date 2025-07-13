package main

import (
	"fmt"
	"log"
	"net/http"

	"flashcards/config"
	"flashcards/db"
	"flashcards/handlers"
	"flashcards/services"
	"flashcards/services/agent"
	"flashcards/services/docindex"
	"flashcards/services/quiz"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	if cfg.PineconeAPIKey == "" {
		log.Fatal("PINECONE_API_KEY environment variable is required")
	}

	if cfg.AnthropicAPIKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	noteRepo, err := db.NewPostgresNoteRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize note database: %v", err)
	}
	defer noteRepo.Close()

	quizRepo, err := db.NewPostgresQuizRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize quiz database: %v", err)
	}
	defer quizRepo.Close()

	docindexService, err := docindex.NewService(cfg.PineconeAPIKey, cfg.OpenAIAPIKey, cfg.PineconeIndexName)
	if err != nil {
		log.Fatalf("Failed to initialize document index service: %v", err)
	}

	noteService := services.NewNoteService(noteRepo)
	noteHandler := handlers.NewNoteHandler(noteService)

	quizStoreService := services.NewQuizStoreService(quizRepo, docindexService)
	quizStoreHandler := handlers.NewQuizStoreHandler(quizStoreService)

	quizService := quiz.NewService(noteService, quizStoreService, cfg.OpenAIAPIKey)
	quizHandler := handlers.NewQuizHandler(quizService)

	agentService, err := agent.NewService(cfg.AnthropicAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize agent service: %v", err)
	}
	agentHandler := handlers.NewAgentHandler(agentService)

	router := mux.NewRouter()

	router.Use(corsMiddleware)
	router.Use(jsonMiddleware)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	noteHandler.RegisterRoutes(router)
	quizStoreHandler.RegisterRoutes(router)
	quizHandler.RegisterRoutes(router)
	agentHandler.RegisterRoutes(router)

	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	addr := ":" + cfg.Port
	fmt.Printf("Server starting on port %s\n", cfg.Port)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
