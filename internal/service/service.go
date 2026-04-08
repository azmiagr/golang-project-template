package service

import (
	"golang-project-template/internal/repository"
	"golang-project-template/pkg/bcrypt"
	"golang-project-template/pkg/jwt"
)

type Service struct {
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{}
}
