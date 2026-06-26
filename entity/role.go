package entity

import "github.com/google/uuid"

type Role struct {
	RoleID   uuid.UUID `json:"role_id" gorm:"type:varchar(36);primaryKey"`
	RoleName string    `json:"role_name" gorm:"type:varchar(255);not null;unique"`

	Users []User `gorm:"foreignKey:RoleID;references:RoleID;constraint:onDelete:CASCADE"`
}
