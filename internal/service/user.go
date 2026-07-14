package service

import (
	"github.com/azmiagr/golang-project-template/entity"
	"github.com/azmiagr/golang-project-template/internal/repository"
	"github.com/azmiagr/golang-project-template/model"
	"github.com/azmiagr/golang-project-template/pkg/database/mariadb"
	"gorm.io/gorm"
)

type IUserService interface {
	GetUser(param model.GetUserParam) (*entity.User, error)
}

type UserService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
}

func NewUserService(userRepo repository.IUserRepository) IUserService {
	return &UserService{
		db:       mariadb.Connection,
		userRepo: userRepo,
	}
}

func (s *UserService) GetUser(param model.GetUserParam) (*entity.User, error) {
	return s.userRepo.GetUser(s.db, param)
}
