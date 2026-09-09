package handlers

import (
	hm "backend/gateway_server/internal/handlers/message"
	s "backend/gateway_server/internal/service"
	m "backend/gateway_server/models"
	sharedGetAuthS "backend/shared/getAuthSession"
	sharedRepoUsers "backend/shared/users"

	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	UserId   string
	RoomId   string
	UserName string
	Avatar   string
	Conn     *websocket.Conn
}

type Handler struct {
	s                *s.Service
	mu               sync.RWMutex
	conns            map[string]*Client
	sharedAuthRedear *sharedGetAuthS.Reader
	sharedRepoUsers  *sharedRepoUsers.RepositoryUser
	handlerMsg       *hm.Handler
}

func NewHandler(s *s.Service, authReader *sharedGetAuthS.Reader, sharedRepoUsers *sharedRepoUsers.RepositoryUser) *Handler {
	return &Handler{
		s:                s,
		mu:               sync.RWMutex{},
		conns:            make(map[string]*Client),
		sharedAuthRedear: authReader,
		sharedRepoUsers:  sharedRepoUsers,
		handlerMsg:       hm.NewHandler(s.ServiceMsg),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("WS handler entered")

	roomID := r.URL.Query().Get("room_id")

	if roomID == "" {
		http.Error(w, "error roomId or userId", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("auth_session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	authSession, err := h.sharedAuthRedear.GetAuth(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("get auth session: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sharedUser, err := h.sharedRepoUsers.GetUserByID(r.Context(), authSession.UserID)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	var user = &m.User{
		Id:       sharedUser.ID.String(),
		UserName: sharedUser.Username,
		Avatar:   sharedUser.Avatar,
	}

	var userID = user.Id

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed %v", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()

	joinRes, err := h.s.Join(ctx, roomID, user)
	if err != nil {
		log.Printf(
			"join room=%s user=%s: %v",
			roomID,
			userID,
			err,
		)

		conn.WriteJSON(map[string]string{
			"err": err.Error(),
		})
		return
	}

	if err := conn.WriteJSON(map[string]any{
		"type": "joined",
		"data": joinRes,
	}); err != nil {
		log.Printf(
			"write joined response user=%s: %v",
			userID,
			err,
		)
		return

	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := h.s.Leave(ctx, roomID, userID); err != nil {
			log.Printf("err leave %v", err)
		}

	}()

	h.mu.Lock()

	if oldClient, ok := h.conns[userID]; ok {
		_ = oldClient.Conn.Close()
	}

	client := &Client{
		UserId:   userID,
		RoomId:   roomID,
		UserName: user.UserName,
		Avatar:   user.Avatar,
		Conn:     conn,
	}

	h.conns[userID] = client
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		if currentClient, ok := h.conns[userID]; ok && currentClient.Conn == conn {
			delete(h.conns, userID)
		}
		h.mu.Unlock()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s disconnected", userID)
			break
		}
		log.Printf("msg from %s: %s", userID, string(msg))
	}
}
