package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	service "backend/api_server/internal/service"
	m "backend/api_server/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PartHandler struct {
	service *service.PartService
}

func NewHandler(service *service.PartService) *PartHandler {
	return &PartHandler{service: service}
}

func (h *PartHandler) RegisterRoutes(r *http.ServeMux) {
	r.Handle("GET /rooms", h.MiddlAuth(http.HandlerFunc(h.GetAllRooms)))
	r.Handle("POST /create-rooms", h.MiddlAuth(http.HandlerFunc(h.CreateRoom)))
	r.Handle("GET /rooms/{id}/users", h.MiddlAuth(http.HandlerFunc(h.GetRoomUsers)))
	r.HandleFunc("POST /avatars/unlock", h.UnlockAvatars)
	r.HandleFunc("POST /auth/email/send-code", h.SendCode)
	r.HandleFunc("POST /auth/email/verify-code", h.VerifyCode)
	r.HandleFunc("GET /auth/session", h.GetSession)
	r.HandleFunc("POST /create-user", h.CreateUser)
	r.HandleFunc("GET /avatars", h.GetAvatars)
	r.HandleFunc("GET /hi", h.Hi)

}

func (h *PartHandler) UnlockAvatars(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	const secretCode = "1122334455"

	if req.Code != secretCode {
		http.Error(w, "invalid code", http.StatusUnauthorized)
		return
	}

	avatars := []string{
		"/static/s1.png",
		"/static/s2.png",
		"/static/s3.png",
		"/static/s4.png",
		"/static/s5.png",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(avatars)
}

func (h *PartHandler) Hi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("HI NIGGERS!")
}

func (h *PartHandler) GetAvatars(w http.ResponseWriter, r *http.Request) {
	avatarsInfo, err := h.service.GetAvatars(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(avatarsInfo); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

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
		"name":       room.RoomName,
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
			"name":       rm.RoomName,
			"owner_id":   rm.OwnerID,
			"created_at": rm.CreatedAt,
			"users":      rm.RoomUsers,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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

func (h *PartHandler) SendCode(w http.ResponseWriter, r *http.Request) {

	type Request struct {
		EmailFC string `json:"email"`
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userA := r.UserAgent()

	sec := m.SendEmailCodeRequest{
		Email:     req.EmailFC,
		IP:        host,
		UserAgent: userA,
	}

	if err = h.service.SendCodeFromEmail(r.Context(), sec); err != nil {
		log.Printf("SendCodeFromEmail error: %v", err)

		http.Error(
			w,
			"send code error: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "code sent"})
}

func (h *PartHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	res, err := h.service.VerifyCode(r.Context(), req.Email, req.Code)
	if err != nil {
		log.Printf("verify code error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookiName := "auth_session"
	maxAge := int((30 * 24 * time.Hour).Seconds())

	if res.RequiresRegister {
		cookiName = "registration_session"
		maxAge = int((30 * time.Minute).Seconds())
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookiName,
		Value:    res.Token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"requires_register": res.RequiresRegister,
	})

}

func (h *PartHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	authCook, authErr := r.Cookie("auth_session")

	if authErr == nil {
		user, err := h.service.GetAuthSession(r.Context(), authCook.Value)

		if err != nil {
			clearCookie(w, "auth_session")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "authorized",
			"user":   user,
		})
		return
	}
	if authErr != nil && !errors.Is(authErr, http.ErrNoCookie) {
		http.Error(w, "invalid auth cookie", http.StatusBadRequest)
		return
	}
	registrationCookie, registrationErr := r.Cookie("registration_session")

	if registrationErr == nil {
		email, err := h.service.GetRegSession(r.Context(), registrationCookie.Value)
		if err != nil {
			clearCookie(w, "registration_session")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "registration",
			"email":  email,
		})
		return
	}

	if registrationErr != nil && !errors.Is(registrationErr, http.ErrNoCookie) {
		http.Error(w, "invalid registration cookie", http.StatusBadRequest)
		return
	}

	http.Error(w, "unauthorized", http.StatusUnauthorized)

}

func (h *PartHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	regCookie, err := r.Cookie("registration_session")
	if err != nil {
		http.Error(w, "registration session required", http.StatusUnauthorized)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	email, err := h.service.GetRegSession(r.Context(), regCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Username, email, req.Avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.service.CreateAuthSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.service.DeleteRegSession(r.Context(), regCookie.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clearCookie(w, "registration_session")

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_session",
		Value:    token,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"avatar":     user.Avatar,
		"created_at": user.CreatedAt,
	})
}
