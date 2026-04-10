package repository

import (
	"errors"
)

var (
	ErrVersionMismatch = errors.New("Version mismatch")
	ErrNotFound        = errors.New("Not found")
	ErrAlreadyExists   = errors.New("Already exists")
)
