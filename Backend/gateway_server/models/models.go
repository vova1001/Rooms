package models

import (
	"encoding/json"
	"time"
)

type User struct {
	Id        string    `json:"id"`
	UserName  string    `json:"user_name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"room_id"`
	UserID    string     `json:"user_id"`
	UserName  string     `json:"user_name"`
	Avatar    string     `json:"avatar"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type IncomingEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type CreateMessagePayload struct {
	Text string `json:"text"`
}

type EditMessagePayload struct {
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
}

type DeleteMessagePayload struct {
	MessageID string `json:"message_id"`
}
