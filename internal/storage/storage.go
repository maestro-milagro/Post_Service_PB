package storage

import "errors"

var (
	ErrNoFollowers  = errors.New("no followers found")
	ErrUserNotFound = errors.New("user not found")
	ErrSubExist     = errors.New("subscription already exist")
)
