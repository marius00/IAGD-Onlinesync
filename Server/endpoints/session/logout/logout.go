package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	api "github.com/marmyr/iagdbackup/api/session/logout"
	"github.com/marmyr/iagdbackup/endpoints/utils"
)

func main() {
	handler := utils.CreatePrivateLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
