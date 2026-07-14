package repository

import (
	"github.com/azmiagr/golang-project-template/entity"
	"github.com/azmiagr/golang-project-template/model"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error) {
	var user entity.User
	if err := tx.Where(&param).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
