### Compilation:
On _linux_/WSL simply run `make clean install`

### Backend deploy:
`sls deploy --verbose`

### Frontend deploy:
`sls client deploy`  
>This will deploy the contents under `client\build`


### Pre-deploy one-time RDS setup:
Prior to deploying for the first time, the following environmental variables must be set to permit RDS access:  
`aws2 ssm put-parameter --name database_user --type String --value postgres`  
`aws2 ssm put-parameter --name database_password --type String --value password`  
`aws2 ssm put-parameter --name database_host --type String --value dbhost`  
`aws2 ssm put-parameter --name database_name --type String --value dbname`
`aws2 ssm put-parameter --name allowed_origin --type String --value http://my-s3-bucket-here.s3-website-us-east-1.amazonaws.com`


**Obs**: Trouble connecting to the RDS? The security group may require you to permit inbound from `0.0.0.0/0`  
(**Research security implications before doing this in production**)

### Pre-deploy lambda setup: (Used for CSP header)
`aws2 ssm put-parameter --name lambda_host --type String --value lambdahostname.execute-api.us-east-1.amazonaws.com`


### Project structure
Shared code is primarily maintained under `internal`, the code behind each endpoint is found under `api`, and the actual endpoint mapping under `endpoints` (in addition to `serverless.yml` for AWS lambda deploys)


Google:
*) Create a new project under `Google cloud functions`.  
Set up credentials: https://serverless.com/framework/docs/providers/google/guide/credentials/



### Adding a new endpoint:
* Create the underlying logic under `api\myEndpointName`, exporting `Path`, `Method` and `ProcessRequest`
* Add a new file to `endpoints\myEndpointName\myEndpointName.go`
* Add the new endpoint to `endpoints\monolith\monolith.go`
* Add a reference to the endpoint in `serverless.yml`