// internal/handler.go
package internal

import (
	"encoding/json"
	"net/http"
	"strings"

	m "rooms/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PartHandler struct {
	service *partService
}

func NewHandler(service *partService) *PartHandler {
	return &PartHandler{service: service}
}

func (h *PartHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/hi", h.Hi).Methods("GET")
	r.HandleFunc("/user/init", h.InitUser).Methods("POST")
	r.HandleFunc("/rooms", h.CreateRoom).Methods("POST")
	r.HandleFunc("/rooms", h.GetAllRooms).Methods("GET")
	r.HandleFunc("/rooms/{id}/users", h.GetRoomUsers).Methods("GET")
}

func (h *PartHandler) Hi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("HI NIGGERS!")
}

func (h *PartHandler) InitUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	user, err := h.service.CreateUser(r.Context(), req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

func (h *PartHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req m.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "room name is required", http.StatusBadRequest)
		return
	}
	ownerIDStr := r.Header.Get("X-User-ID")
	if ownerIDStr == "" {
		http.Error(w, "X-User-ID header required", http.StatusBadRequest)
		return
	}
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	room, err := h.service.CreateRoom(r.Context(), req.Name, ownerID)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint") {
			http.Error(w, "owner not found", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         room.ID,
		"name":       room.Name,
		"owner_id":   room.OwnerID,
		"created_at": room.CreatedAt,
	})
}

func (h *PartHandler) GetAllRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.GetAllRooms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]map[string]interface{}, len(rooms))
	for i, rm := range rooms {
		resp[i] = map[string]interface{}{
			"id":         rm.ID,
			"name":       rm.Name,
			"owner_id":   rm.OwnerID,
			"created_at": rm.CreatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PartHandler) GetRoomUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["id"]
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}
	users, err := h.service.GetRoomUsers(r.Context(), roomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]map[string]interface{}, len(users))
	for i, u := range users {
		resp[i] = map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
