package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flashcards/config"
	"flashcards/db"
	"flashcards/handlers"
	"flashcards/mcp"
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

	memoryRepo, err := db.NewPostgresMemoryRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize memory database: %v", err)
	}
	defer memoryRepo.Close()

	noteService := services.NewNoteService(noteRepo)
	noteHandler := handlers.NewNoteHandler(noteService)

	memoryService := services.NewMemoryService(memoryRepo)

	knowledgeCheckRepo, err := db.NewPostgresKnowledgeCheckRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize knowledge check database: %v", err)
	}
	defer knowledgeCheckRepo.Close()

	knowledgeCheckService := services.NewKnowledgeCheckService(knowledgeCheckRepo)

	quizStoreService := services.NewQuizStoreService(quizRepo, docindexService)
	quizStoreHandler := handlers.NewQuizStoreHandler(quizStoreService)

	quizService := quiz.NewService(noteService, quizStoreService, cfg.OpenAIAPIKey)
	quizHandler := handlers.NewQuizHandler(quizService)

	agentService, err := agent.NewService(cfg.AnthropicAPIKey, noteService, memoryService, knowledgeCheckService)
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

	// Initialize MCP server
	mcpServer := mcp.NewServer(noteService, memoryService, knowledgeCheckService)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start HTTP API server
	httpAddr := ":" + cfg.Port
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	go func() {
		log.Printf("[INFO] Starting HTTP server on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] HTTP server failed: %v", err)
		}
	}()

	// Start MCP server in goroutine
	go func() {
		if err := mcpServer.Run(ctx, cfg.MCPPort, cfg.MCPAPIKey); err != nil {
			log.Printf("[ERROR] MCP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("[INFO] Shutdown signal received, initiating graceful shutdown...")

	// Cancel context to trigger MCP server shutdown
	cancel()

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] HTTP server shutdown error: %v", err)
	}

	log.Println("[INFO] All servers stopped gracefully")
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
