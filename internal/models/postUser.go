package models

type PostUser struct {
	Email string `json:"email"`
	// TODO: Может измениться
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
