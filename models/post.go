package models

type Post struct {
	Id       int    `json:"id,omitempty"`
	Parent   int    `json:"parent,omitempty"`
	Author   string `json:"author"`
	Message  string `json:"message"`
	IsEdited bool   `json:"isEdited,omitempty"`
	Forum    string `json:"forum,omitempty"`
	Thread   int    `json:"thread"`
	Created  string `json:"created,omitempty"`
}

// easyjson:json
type Posts []Post

type PostUpdate struct {
	Message string `json:"message"`
}

type PostFull struct {
	Post   Post    `json:"post"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}
