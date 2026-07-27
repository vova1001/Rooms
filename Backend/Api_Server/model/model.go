package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	Avatar    string    `json:"avatar"`
	Email     string    `json:"email"`
}

type Room struct {
	ID        uuid.UUID `json:"id"`
	RoomName  string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	RoomUsers []RoomUser
}

type RoomUser struct {
	RoomID   uuid.UUID `json:"room_id"`
	UserInfo User
}

// res\req

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type RoomResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type RoomUsersResponse struct {
	RoomID uuid.UUID      `json:"room_id"`
	Users  []UserResponse `json:"users"`
}

type SendEmailCodeRequest struct {
	Email     string `json:"email"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
}
