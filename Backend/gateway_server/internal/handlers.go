package internal

import (
	m "backend/gateway_server/models"
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

type Handler struct {
	s     *Service
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		s:     s,
		conns: make(map[string]*websocket.Conn),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("WS handler entered")

	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")
	userName := r.URL.Query().Get("user_name")
	avatar := r.URL.Query().Get("avatar")

	if roomID == "" || userID == "" {
		http.Error(w, "error roomId or userId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed %v", err)
		return
	}
	defer conn.Close()

	var user = &m.User{
		Id:       userID,
		UserName: userName,
		Avatar:   avatar,
	}

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

	if oldConn, ok := h.conns[userID]; ok {
		_ = oldConn.Close()
	}
	h.conns[userID] = conn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		if currentConn, ok := h.conns[userID]; ok && currentConn == conn {
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
