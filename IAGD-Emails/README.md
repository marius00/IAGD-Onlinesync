### Compilation:
On _linux_/WSL simply run `make clean install`

### Backend deploy:
`sls deploy --verbose`

### Pre-deploy setup:
Prior to deploying for the first time, the following environmental variable must be set to permit access to the API:
`aws2 ssm put-parameter --name secret --type String --value password`

### Project structure
Shared code is primarily maintained under `internal`, the code behind each endpoint is found under `api`, and the actual endpoint mapping under `endpoints` (in addition to `serverless.yml` for AWS lambda deploys)

### Adding a new endpoint:
* Create the underlying logic under `api\myEndpointName`, exporting `Path`, `Method` and `ProcessRequest`
* Add a new file to `endpoints\myEndpointName\myEndpointName.go`, ensure that the package is named `main`
* Add the new endpoint to `endpoints\monolith\monolith.go`
* Add a reference to the endpoint in `serverless.yml`

### Using the API
* The API is exposed at the endpoint `/sendmail` and expects a header `Authorized` containing just the access token, without any bearer type,
and a json body matching `{"email": "email@example.com", "code": "123456789"}`. The access token must not exceed 9 characters.