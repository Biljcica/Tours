package handler

import (
	"encoding/json"
	"net/http"

	"database-example/service"
	"time" 
	"github.com/gorilla/mux"
)

type TourHandler struct {
	TourService *service.TourService
}

// CreateTour -> POST /tours
func (h *TourHandler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AuthorID    string   `json:"author_id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Difficulty  string   `json:"difficulty"`
		Tags        []string `json:"tags"`
		Status      string   `json:"status"`
		Price       float64  `json:"price" bson:"price"`      
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tour, err := h.TourService.CreateTour(body.AuthorID, body.Name, body.Description, body.Difficulty, body.Tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tour)
}

// GetTour -> GET /tours/{id}
func (h *TourHandler) GetTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	tour, err := h.TourService.GetTour(id)
	if err != nil {
		http.Error(w, "tour not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tour)
}

// GetToursByAuthor -> GET /authors/{authorID}/tours
func (h *TourHandler) GetToursByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authorID := vars["authorID"]

	tours, err := h.TourService.GetToursByAuthor(authorID)
	if err != nil {
		http.Error(w, "could not fetch tours", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tours)
}

// PublishTour -> PUT /tours/{id}/publish
func (h *TourHandler) PublishTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.TourService.PublishTour(id)
	if err != nil {
		http.Error(w, "could not publish tour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Tour published successfully"}`))
}

// GetAllTours -> GET /tours
func (h *TourHandler) GetAllTours(w http.ResponseWriter, r *http.Request) {
	tours, err := h.TourService.GetAllTours()
	if err != nil {
		http.Error(w, "could not fetch tours", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tours)
}

