package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	api "github.com/marmyr/myservice/api/partitions"
	"github.com/marmyr/myservice/endpoints/utils"
)

func main() {
	handler := utils.CreatePrivateLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
