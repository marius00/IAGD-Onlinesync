package main // Important

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	api "github.com/marmyr/iagdsendmail/api/sendmail"
	"github.com/marmyr/iagdsendmail/internal/eventbus"
)

var initialized = false
var ginLambda *ginadapter.GinLambda
func createPublicLambdaEntrypoint(path string, method string, fn gin.HandlerFunc) interface{} {
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

func main() {
	handler := createPublicLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
