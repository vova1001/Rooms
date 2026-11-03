package model

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateProfileReq struct {
	Username *string
	Bio      *string
}

type UpdateRoomPrivacyReq struct {
	RoomID    uuid.UUID `json:"-"`
	OwnerID   uuid.UUID `json:"-"`
	IsPrivate bool      `json:"-"`
	PassHash  *string   `json:"-"`
}

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

type VerifyCodeResult struct {
	Token            string
	RequiresRegister bool
}

type CreateEventReq struct {
	RoomID      uuid.UUID `json:"-"`
	CreatorID   uuid.UUID `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	IsFree      bool      `json:"is_free"`
	PriceMinor  int64     `json:"price_minor"`
	Capacity    int       `json:"capacity"`
}

type Event struct {
	ID          uuid.UUID `json:"id"`
	RoomID      uuid.UUID `json:"room_id"`
	CreatorID   uuid.UUID `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	IsFree      bool      `json:"is_free"`
	PriceMinor  int64     `json:"price_minor"`
	Currency    string    `json:"currency"`
	Capacity    int       `json:"capacity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type EventRoom struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Visibility string    `json:"visibility"`
}

type EventItem struct {
	ID                uuid.UUID `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	StartsAt          time.Time `json:"starts_at"`
	EndsAt            time.Time `json:"ends_at"`
	IsFree            bool      `json:"is_free"`
	PriceMinor        int64     `json:"price_minor"`
	Currency          string    `json:"currency"`
	Capacity          int       `json:"capacity"`
	ParticipantsCount int       `json:"participants_count"`
	IsFull            bool      `json:"is_full"`
	Room              EventRoom `json:"room"`
}

type Booking struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MyBooking — заявка пользователя вместе с карточкой события (для профиля)
type MyBooking struct {
	Booking
	Event EventItem `json:"event"`
}
