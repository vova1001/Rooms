package handlers

import (
	hm "backend/gateway_server/internal/handlers/message"
	s "backend/gateway_server/internal/service"
	smsg "backend/gateway_server/internal/service/message"
	m "backend/gateway_server/models"
	sharedGetAuthS "backend/shared/getAuthSession"
	sharedRepoUsers "backend/shared/users"

	"context"
	"encoding/json"
	"errors"
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
	writeMu  sync.Mutex
}

type Handler struct {
	s                *s.Service
	mu               sync.RWMutex
	conns            map[string]*Client
	rooms            map[string]map[string]*Client
	sharedAuthRedear *sharedGetAuthS.Reader
	sharedRepoUsers  *sharedRepoUsers.RepositoryUser
	handlerMsg       *hm.Handler
}

func NewHandler(s *s.Service, authReader *sharedGetAuthS.Reader, sharedRepoUsers *sharedRepoUsers.RepositoryUser) *Handler {
	return &Handler{
		s:                s,
		mu:               sync.RWMutex{},
		conns:            make(map[string]*Client),
		rooms:            make(map[string]map[string]*Client),
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
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[string]*Client)
	}
	h.rooms[roomID][userID] = client
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		if currentClient, ok := h.conns[userID]; ok && currentClient.Conn == conn {
			delete(h.conns, userID)
			delete(h.rooms[roomID], userID)
			if len(h.rooms[roomID]) == 0 {
				delete(h.rooms, roomID)
			}
		}
		h.mu.Unlock()
	}()
	history, err := h.handlerMsg.History(ctx, roomID, nil)
	if err != nil {
		log.Printf("history room=%s: %v", roomID, err)
	} else {
		client.Send(map[string]any{
			"type": "message.history",
			"data": history,
		})
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s disconnected", userID)
			break
		}
		var event m.IncomingEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			client.Send(map[string]string{"type": "message.error", "error": "bad json"})
			continue
		}

		switch event.Type {
		case "message.create":
			message, err := h.handlerMsg.Create(ctx, roomID, userID, user.UserName, user.Avatar, event.Data)
			if errors.Is(err, smsg.ErrRateLimited) {
				client.Send(map[string]string{"type": "message.error", "error": "Слишком много сообщений. Подождите пару секунд."})
				continue
			}
			if err != nil {
				log.Printf("create msg user=%s: %v", userID, err)
				client.Send(map[string]string{"type": "message.error", "error": err.Error()})
				continue
			}
			h.broadcast(roomID, map[string]any{
				"type": "message.created",
				"data": message,
			})

		case "message.history":
			history, err := h.handlerMsg.History(ctx, roomID, event.Data)
			if err != nil {
				log.Printf("history user=%s: %v", userID, err)
				client.Send(map[string]string{"type": "message.error", "error": "failed to load history"})
				continue
			}
			client.Send(map[string]any{
				"type": "message.history",
				"data": history,
			})

		default:
			client.Send(map[string]string{"type": "message.error", "error": "unknown event type"})
		}
	}
}
func (h *Handler) broadcast(roomID string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.rooms[roomID] {
		if err := client.Send(payload); err != nil {
			log.Printf("broadcast to %s: %v", client.UserId, err)
		}
	}
}

func (c *Client) Send(payload any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.Conn.WriteJSON(payload)
}
