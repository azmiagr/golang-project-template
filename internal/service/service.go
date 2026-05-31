package service

import (
	"github.com/azmiagr/golang-project-template/internal/repository"
	"github.com/azmiagr/golang-project-template/pkg/bcrypt"
	"github.com/azmiagr/golang-project-template/pkg/jwt"
)

type Service struct {
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{}
}
