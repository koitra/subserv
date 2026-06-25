package uuidext

import "github.com/google/uuid"

func NewV7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
