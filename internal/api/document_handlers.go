package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// DocumentRequest represents a request to create or update a document
type DocumentRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// createDocument handles POST /api/v1/documents
func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Title == "" || req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Create document
	id, err := s.service.CreateDocument(r.Context(), req.Title, req.Content, req.Metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the document ID
	respondWithJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// getDocument handles GET /api/v1/documents/{id}
func (s *Server) getDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get document
	doc, err := s.service.GetDocument(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Document not found")
		return
	}

	// Return the document
	respondWithJSON(w, http.StatusOK, doc)
}

// updateDocument handles PUT /api/v1/documents/{id}
func (s *Server) updateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Title == "" || req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Update document
	err := s.service.UpdateDocument(r.Context(), id, req.Title, req.Content, req.Metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// deleteDocument handles DELETE /api/v1/documents/{id}
func (s *Server) deleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete document
	err := s.service.DeleteDocument(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// searchDocuments handles GET /api/v1/documents?query=...
func (s *Server) searchDocuments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Query parameter is required")
		return
	}

	// Search documents
	docs, err := s.service.SearchDocuments(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the documents
	respondWithJSON(w, http.StatusOK, docs)
}
