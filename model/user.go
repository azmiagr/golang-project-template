package model

import "github.com/google/uuid"

type GetUserParam struct {
	UserID uuid.UUID `json:"-"`
}
