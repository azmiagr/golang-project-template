package entity

import "github.com/google/uuid"

type User struct {
	UserID uuid.UUID `json:"id" gorm:"type:varchar(36);primaryKey"`
	RoleID uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	Status string    `json:"status" gorm:"type:enum('active','inactive');default:'inactive'"`
}
