package utils

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
)

var initialized = false
var ginLambda *ginadapter.GinLambda

func CreatePrivateLambdaEntrypoint(path string, method string, fn gin.HandlerFunc) interface{} {
	Handler := func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if !initialized {
			ginEngine := eventbus.MountProtectedRoute(path, method, fn)
			ginLambda = ginadapter.New(ginEngine)
			initialized = true
		}
		return ginLambda.Proxy(req)
	}

	return Handler
}

func CreatePublicLambdaEntrypoint(path string, method string, fn gin.HandlerFunc) interface{} {
	Handler := func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if !initialized {
			ginEngine := eventbus.MountPublicRoute(path, method, fn)
			ginLambda = ginadapter.New(ginEngine)
			initialized = true
		}
		return ginLambda.Proxy(req)
	}

	return Handler
}
