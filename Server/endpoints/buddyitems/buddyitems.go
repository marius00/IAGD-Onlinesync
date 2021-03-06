package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	api "github.com/marmyr/iagdbackup/api/buddyitems"
	"github.com/marmyr/iagdbackup/endpoints/utils"
)

func main() {
	handler := utils.CreatePrivateLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
// https://github.com/awslabs/aws-lambda-go-api-proxy/blob/master/core/request.go#L20
// https://github.com/awslabs/aws-lambda-go-api-proxy/pull/21