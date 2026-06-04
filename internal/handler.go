package internal

import (
	"encoding/json"
	"log"
	"net/http"
)

type PartHandler struct {
	service *partService
}

func NewHandler(service *partService) *PartHandler {
	return &PartHandler{service: service}
}

func (h *PartHandler) RegisterRote(mux *http.ServeMux) {
	mux.HandleFunc("GET /hi", h.Hi)
}

func (h *PartHandler) Hi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("HI NIGGERS!"); err != nil {
		log.Printf("encode failed: %v", err)
		return
	}
}
