package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

func main() {
	config.Init()
	auth.Init()
	config.InitCrypto()
	auth.InitOauth()

	dsn := os.Getenv("DATABASE_DSN")
	if err := config.Connect(context.Background(), dsn); err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	userRepo := user.NewRepository(config.DB)
	userService := user.NewService(userRepo)
	handler := user.NewHandler(userService).Handle

	lambda.Start(handler)
}
