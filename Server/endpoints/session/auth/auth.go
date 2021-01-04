package main


import (
	"github.com/aws/aws-lambda-go/lambda"
	api "github.com/marmyr/iagdbackup/api/session/auth"
	"github.com/marmyr/iagdbackup/endpoints/utils"
)

func main() {
	handler := utils.CreatePublicLambdaEntrypoint(api.Path, api.Method, api.ProcessRequest)
	lambda.Start(handler)
}
