package utils

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/routing"
)

var ginLambda *ginadapter.GinLambda

func CreatePrivateLambdaEntrypoint(path string, method string, fn gin.HandlerFunc) interface{} {
	return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ginEngine := routing.MountProtectedRoute(path, method, fn)
		ginLambda = ginadapter.New(ginEngine)
		return ginLambda.Proxy(req)
	}
}

func CreatePublicLambdaEntrypoint(path string, method string, fn gin.HandlerFunc) interface{} {
	return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ginEngine := routing.MountPublicRoute(path, method, fn)
		ginLambda = ginadapter.New(ginEngine)
		return ginLambda.Proxy(req)
	}
}
