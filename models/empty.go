package models

import json "encoding/json"

// easyjson:skip
type EmptyStruct struct {
}

// easyjson:skip
type EmptyArray []struct{}

func (e EmptyArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{})
}
