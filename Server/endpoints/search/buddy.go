package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	api "github.com/marmyr/myservice/api/search"
	"github.com/marmyr/myservice/endpoints/utils"
)

func main() {
	handler := utils.CreatePublicLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
