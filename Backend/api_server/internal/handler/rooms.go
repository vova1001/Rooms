package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	repository "backend/api_server/internal/repository"
	service "backend/api_server/internal/service"

	"github.com/google/uuid"
)

func (h *PartHandler) UnlockRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	roomID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.service.UnlockRoom(r.Context(), roomID, userID, body.Password)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidRoomCode) {
			http.Error(w, "invalid room code", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrTooManyAttempts) {
			http.Error(w, "too many attempts", http.StatusTooManyRequests)
			return
		}
		log.Printf("unlock room: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
