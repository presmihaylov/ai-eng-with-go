package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"flashcards/mcp/tools"
	"flashcards/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	mcpServer            *mcp.Server
	noteService          *services.NoteService
	memoryService        *services.MemoryService
	knowledgeCheckService *services.KnowledgeCheckService
}

func NewServer(
	noteService *services.NoteService,
	memoryService *services.MemoryService,
	knowledgeCheckService *services.KnowledgeCheckService,
) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "flashcards-mcp-server",
		Version: "1.0.0",
	}, nil)

	server := &Server{
		mcpServer:            mcpServer,
		noteService:          noteService,
		memoryService:        memoryService,
		knowledgeCheckService: knowledgeCheckService,
	}

	server.registerAllTools()

	return server
}

func (s *Server) registerAllTools() {
	// Note tools (read-only)
	s.registerNoteTools()

	// Memory tools (full CRUD)
	s.registerMemoryTools()

	// Knowledge check tools (full CRUD)
	s.registerKnowledgeCheckTools()
}

func (s *Server) Run(ctx context.Context, port string, apiKey string) error {
	log.Printf("[INFO] Starting MCP HTTP server on port %s...", port)

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	// Wrap with auth middleware if API key is provided
	var finalHandler http.Handler = handler
	if apiKey != "" {
		finalHandler = authMiddleware(apiKey, handler)
		log.Printf("[INFO] MCP HTTP server authentication enabled")
	}

	addr := ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: finalHandler,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] MCP HTTP server failed: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	log.Println("[INFO] Shutting down MCP HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*http.DefaultClient.Timeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] MCP HTTP server shutdown error: %v", err)
		return fmt.Errorf("MCP HTTP server shutdown failed: %w", err)
	}

	log.Println("[INFO] MCP HTTP server stopped")
	return nil
}

func authMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[ERROR] Unauthorized MCP request from %s: missing Authorization header", r.RemoteAddr)
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		expectedAuth := "Bearer " + apiKey
		if authHeader != expectedAuth {
			log.Printf("[ERROR] Unauthorized MCP request from %s: invalid token", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) registerNoteTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_notes",
		Description: "Get all notes metadata (ID, 200-char preview, timestamps) without full content. Preview shows first 200 characters to help identify note topics. Use get_note with specific ID to retrieve full or partial content.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListNotesInput) (*mcp.CallToolResult, tools.ListNotesOutput, error) {
		return tools.ListNotes(ctx, req, input, s.noteService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_note",
		Description: "Get a specific note by ID. Supports optional offset (line number to start from, 1-based) and limit (number of lines) to retrieve partial content and avoid context bloat. Response includes total_lines, lines_returned, and offset_used for pagination.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GetNoteInput) (*mcp.CallToolResult, tools.GetNoteOutput, error) {
		return tools.GetNote(ctx, req, input, s.noteService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "search_notes",
		Description: "Search notes by content keywords. Returns matching note metadata (ID, 200-char preview, timestamps) without full content. Preview shows first 200 characters to help identify note topics. Use get_note with specific ID to retrieve full or partial content.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SearchNotesInput) (*mcp.CallToolResult, tools.SearchNotesOutput, error) {
		return tools.SearchNotes(ctx, req, input, s.noteService)
	})
}

func (s *Server) registerMemoryTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_memory",
		Description: "Get the agent's memory content",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GetMemoryInput) (*mcp.CallToolResult, tools.GetMemoryOutput, error) {
		return tools.GetMemory(ctx, req, input, s.memoryService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_memory",
		Description: "Update the agent's memory content",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateMemoryInput) (*mcp.CallToolResult, tools.UpdateMemoryOutput, error) {
		return tools.UpdateMemory(ctx, req, input, s.memoryService)
	})
}

func (s *Server) registerKnowledgeCheckTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_knowledge_checks",
		Description: "Get all knowledge checks, optionally filtered by date range",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListKnowledgeChecksInput) (*mcp.CallToolResult, tools.ListKnowledgeChecksOutput, error) {
		return tools.ListKnowledgeChecks(ctx, req, input, s.knowledgeCheckService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_knowledge_check",
		Description: "Get a specific knowledge check by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GetKnowledgeCheckInput) (*mcp.CallToolResult, tools.GetKnowledgeCheckOutput, error) {
		return tools.GetKnowledgeCheck(ctx, req, input, s.knowledgeCheckService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_knowledge_check",
		Description: "Create a new knowledge check for a note section",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateKnowledgeCheckInput) (*mcp.CallToolResult, tools.CreateKnowledgeCheckOutput, error) {
		return tools.CreateKnowledgeCheck(ctx, req, input, s.knowledgeCheckService)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_knowledge_check",
		Description: "Update an existing knowledge check (state, score, explanation, or topic)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateKnowledgeCheckInput) (*mcp.CallToolResult, tools.UpdateKnowledgeCheckOutput, error) {
		return tools.UpdateKnowledgeCheck(ctx, req, input, s.knowledgeCheckService)
	})
}
