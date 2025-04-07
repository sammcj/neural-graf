package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// ConceptRequest represents a request to create a concept
type ConceptRequest struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// LinkConceptsRequest represents a request to link two concepts
type LinkConceptsRequest struct {
	FromID           string                 `json:"fromId"`
	ToID             string                 `json:"toId"`
	RelationshipType string                 `json:"relationshipType"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// createConcept handles POST /api/v1/concepts
func (s *Server) createConcept(w http.ResponseWriter, r *http.Request) {
	var req ConceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Create concept
	id, err := s.service.CreateConcept(r.Context(), req.Name, req.Properties)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the concept ID
	respondWithJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// getConcept handles GET /api/v1/concepts/{id}
func (s *Server) getConcept(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get concept
	concept, err := s.service.GetConcept(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Concept not found")
		return
	}

	// Return the concept
	respondWithJSON(w, http.StatusOK, concept)
}

// linkConcepts handles POST /api/v1/concepts/link
func (s *Server) linkConcepts(w http.ResponseWriter, r *http.Request) {
	var req LinkConceptsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.FromID == "" || req.ToID == "" || req.RelationshipType == "" {
		respondWithError(w, http.StatusBadRequest, "FromID, ToID, and RelationshipType are required")
		return
	}

	// Link concepts
	id, err := s.service.LinkConcepts(r.Context(), req.FromID, req.ToID, req.RelationshipType, req.Properties)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the edge ID
	respondWithJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// searchConcepts handles GET /api/v1/concepts?query=...
func (s *Server) searchConcepts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Query parameter is required")
		return
	}

	// Search concepts
	concepts, err := s.service.SearchConcepts(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the concepts
	respondWithJSON(w, http.StatusOK, concepts)
}
