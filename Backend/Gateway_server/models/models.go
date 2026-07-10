package models

import "time"

type User struct {
	Id        string    `json:"id"`
	UserName  string    `json:"user_name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}
