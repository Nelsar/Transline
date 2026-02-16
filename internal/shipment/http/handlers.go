package http

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/google/uuid"
	shservice "transline.kz/internal/shipment/service"
)

// Handler — HTTP слой, НЕ содержит бизнес-логики
type Handler struct {
	service *shservice.Service
}

// New создаёт HTTP handler
func New(service *shservice.Service) *Handler {
	return &Handler{service: service}
}

// ===== DTO =====

type createShipmentRequest struct {
	Route    string  `json:"route"`
	Price    float64 `json:"price"`
	Customer struct {
		IDN string `json:"idn"`
	} `json:"customer"`
}

type createShipmentResponse struct {
	ID         uuid.UUID `json:"id"`
	Status     string    `json:"status"`
	CustomerID uuid.UUID `json:"customerId"`
}

// ===== Handlers =====

// Create — POST /api/v1/shipments
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	result, err := h.service.CreateShipment(
		r.Context(),
		shservice.CreateShipmentInput{
			Route: req.Route,
			Price: req.Price,
			IDN:   req.Customer.IDN,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := createShipmentResponse{
		ID:         result.ID,
		Status:     result.Status,
		CustomerID: result.CustomerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("error encoding response", "err", err)
	}
}
