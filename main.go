package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/router"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

var chiLambda *chiadapter.ChiLambdaV2

func init() {
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
	userHandler := user.NewHandler(userService)

	r := router.New(router.RouterConfig{
		UserHandler: userHandler,
	})

	chiLambda = chiadapter.NewV2(r.(*chi.Mux))
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return chiLambda.ProxyWithContextV2(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
