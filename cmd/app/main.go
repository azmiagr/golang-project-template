package main

import (
	"golang-project-template/internal/handler/rest"
	"golang-project-template/internal/repository"
	"golang-project-template/internal/service"
	"golang-project-template/pkg/bcrypt"
	"golang-project-template/pkg/config"
	"golang-project-template/pkg/database/mariadb"
	"golang-project-template/pkg/jwt"
	"golang-project-template/pkg/middleware"
	"log"
)

func main() {
	config.LoadEnvironment()

	db, err := mariadb.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	bcrypt := bcrypt.Init()
	jwt := jwt.Init()
	svc := service.NewService(repo, bcrypt, jwt)

	middleware := middleware.Init(svc, jwt)
	r := rest.NewRest(svc, middleware)
	r.MountEndpoint()

	r.Run()
}
