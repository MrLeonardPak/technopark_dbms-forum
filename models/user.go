package models

type User struct {
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
	Nickname string `json:"nickname,omitempty"`
}

// easyjson:json
type Users []User
