package api

import (
	"encoding/json"
	"net/http"
)

// QueryRequest represents a request to query the knowledge graph
type QueryRequest struct {
	Query  string                 `json:"query"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// SchemaRequest represents a request to update the schema
type SchemaRequest struct {
	Schema string `json:"schema"`
}

// query handles POST /api/v1/query
func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Query == "" {
		respondWithError(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Execute query directly against the graph store
	results, err := s.graph.Query(r.Context(), req.Query, req.Params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the results
	respondWithJSON(w, http.StatusOK, results)
}

// upsertSchema handles POST /api/v1/schema
func (s *Server) upsertSchema(w http.ResponseWriter, r *http.Request) {
	var req SchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Schema == "" {
		respondWithError(w, http.StatusBadRequest, "Schema is required")
		return
	}

	// Update schema
	err := s.service.InitialiseSchema(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}
