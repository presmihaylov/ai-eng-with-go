package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"flashcards/models"
	"flashcards/services/agent"

	"github.com/gorilla/mux"
)

type AgentHandler struct {
	service *agent.Service
}

func NewAgentHandler(service *agent.Service) *AgentHandler {
	return &AgentHandler{service: service}
}

func (h *AgentHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/agent/chat", h.ProcessMessage).Methods("POST")
}

func (h *AgentHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received agent chat request")

	var req models.AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode agent request JSON: %v", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if len(req.Messages) == 0 {
		log.Printf("[ERROR] No messages provided in agent request")
		h.writeErrorResponse(w, http.StatusBadRequest, "At least one message is required")
		return
	}

	result, err := h.service.ProcessMessage(req.Messages)
	if err != nil {
		log.Printf("[ERROR] Agent message processing failed: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Agent message processing completed successfully")
	h.writeJSONResponse(w, http.StatusOK, result)
}

func (h *AgentHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AgentHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}