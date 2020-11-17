### Compilation:
On _linux_/WSL simply run `make clean install`

### Backend deploy:
`sls deploy --verbose`

### Frontend deploy:
`sls client deploy`  
>This will deploy the contents under `client\build`


### Pre-deploy one-time RDS setup:
Prior to deploying for the first time, the following environmental variables must be set to permit RDS access:  
**TODO**

### Pre-deploy lambda setup: (Used for CSP header)
`aws2 ssm put-parameter --name lambda_host --type String --value lambdahostname.execute-api.us-east-1.amazonaws.com`


### Project structure
Shared code is primarily maintained under `internal`, the code behind each endpoint is found under `api`, and the actual endpoint mapping under `endpoints` (in addition to `serverless.yml` for AWS lambda deploys)


### Adding a new endpoint:
* Create the underlying logic under `api\myEndpointName`, exporting `Path`, `Method` and `ProcessRequest`
* Add a new file to `endpoints\myEndpointName\myEndpointName.go` which is the endpoint AWS will be using
* Add the new endpoint to `endpoints\monolith\monolith.go`, which allows it to run locally and anywhere aside from AWS.
* Add a reference to the endpoint in `serverless.yml` which adds a binary:endpoint mapping for the AWS deploy


# How does it work?
Data is partitioned per user.

### Authentication
The API expects the following headers to be set:
* `X-Api-User: user@example.com`  
* `Authorization: AccessTokenWithoutBearerPrefix`  
Endpoints:

### Upload
The upload endpoints accepts an array of items. Each item have the field `id` with a GUID value.  
The endpoint returns the partition the items were stored in, as well as any items which were not processed due to any errors. 

### Download
The download endpoint will return:
* The current server timestamp [TODO: Is this safe? what if there's concurrent uploads?]
* All items stored for a given partition, filter by a server timestamp.  
* All items that needs to be removed, which may be located in a different partition.

 
### Remove
The remove endpoint will accepts a list of items which should be deleted. The items can be located in any partition.  
The items will be removed from the corresponding partition, and added as a deletion entry to the current active partition. This ensures that all clients will be notified of a pending deletion.  
Entries in the array are expected to be sorted by partition key, _unsorted requests may get rejected_.

### Partitions
The partitions endpoint will return a list of all the partitions for a given user. Any unknown partitions should be fully synced down, as well as any partition which may have been closed since the last call to partitions.