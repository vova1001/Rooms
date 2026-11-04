package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	repository "backend/api_server/internal/repository"
	service "backend/api_server/internal/service"
	m "backend/api_server/model"

	"github.com/google/uuid"
)

func (h *PartHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
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

	var req m.CreateEventReq
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.RoomID = roomID
	req.CreatorID = userID

	event, err := h.service.CreateEvent(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrRoomNotFound) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		log.Printf("create event: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(event); err != nil {
		log.Printf("encode event: %v", err)
		return
	}
}

func (h *PartHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	events, err := h.service.ListEvents(r.Context(), q.Get("date"), q.Get("time_zone"))
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("list events: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// фронт ждёт {items, next_cursor}; пагинации нет, поэтому next_cursor всегда null
	if err = json.NewEncoder(w).Encode(map[string]any{"items": events, "next_cursor": nil}); err != nil {
		log.Printf("encode events: %v", err)
		return
	}
}

func (h *PartHandler) JoinEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	eventID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}

	booking, err := h.service.JoinEvent(r.Context(), eventID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrEventStarted) {
			http.Error(w, "event already started", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrEventFull) {
			http.Error(w, "no seats left", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrAlreadyBooked) {
			http.Error(w, "already booked", http.StatusConflict)
			return
		}
		log.Printf("join event: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(booking); err != nil {
		log.Printf("encode booking: %v", err)
		return
	}
}

func (h *PartHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bookings, err := h.service.GetUserBookings(r.Context(), userID)
	if err != nil {
		log.Printf("get my bookings: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(bookings); err != nil {
		log.Printf("encode bookings: %v", err)
		return
	}
}

func (h *PartHandler) GetMyEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := h.service.GetOwnerEvents(r.Context(), userID)
	if err != nil {
		log.Printf("get my events: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(events); err != nil {
		log.Printf("encode events: %v", err)
		return
	}
}
