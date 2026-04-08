package entity

import "github.com/google/uuid"

type YourEntity struct {
	ID uuid.UUID `json:"id" gorm:"type:varchar(36);primaryKey"`
}
