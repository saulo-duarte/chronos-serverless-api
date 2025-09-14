package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"

	"github.com/saulo-duarte/chronos-lambda/internal/container"
	"github.com/saulo-duarte/chronos-lambda/internal/router"
)

var chiLambda *chiadapter.ChiLambdaV2

func init() {
	c := container.New()

	r := router.New(router.RouterConfig{
		UserHandler:         c.UserContainer.Handler,
		ProjectHandler:      c.ProjectContainer.Handler,
		TaskHandler:         c.TaskContainer.Handler,
		StudySubjectHandler: c.StudySubjectContainer.Handler,
	})

	chiLambda = chiadapter.NewV2(r.(*chi.Mux))
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("DEBUG: Incoming request: %+v\n", req)

	resp, err := chiLambda.ProxyWithContextV2(ctx, req)
	if err != nil {
		log.Printf("ERROR: ProxyWithContextV2 returned an error: %v\n", err)
	}
	return resp, err
}

func main() {
	log.Println("Lambda starting...")
	lambda.Start(Handler)
}
