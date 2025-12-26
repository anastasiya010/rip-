package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/config"
	"metoda/internal/app/dsn"
	"metoda/internal/app/handler"
	"metoda/internal/app/repository"
	"metoda/internal/app/storage"
	"metoda/internal/pkg"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Инициализируем Minio клиент
	minioClient, errMinio := storage.NewMinioClient()
	if errMinio != nil {
		logrus.Fatalf("error initializing minio client: %v", errMinio)
	}

	hand := handler.NewHandler(rep, minioClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}

