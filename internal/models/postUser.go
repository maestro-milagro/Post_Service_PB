package models

type PostUser struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
