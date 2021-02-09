package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/marmyr/iagdbackup/api/character"
	"github.com/marmyr/iagdbackup/internal/routing"
)

func main() {
	handler := func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		engine := routing.MountProtectedRoute(character.UploadPath, character.UploadMethod, character.UploadProcessRequest)
		routing.AddProtectedRoute(engine, character.DownloadPath, character.DownloadMethod, character.DownloadProcessRequest)
		routing.AddProtectedRoute(engine, character.ListPath, character.ListMethod, character.ListProcessRequest)
		ginLambda := ginadapter.New(engine)
		return ginLambda.Proxy(req)
	}

	lambda.Start(handler)
}
