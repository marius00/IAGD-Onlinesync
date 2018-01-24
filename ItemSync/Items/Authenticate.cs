using System;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Security.Claims;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;

namespace ItemSync.Items
{
    public static class Authenticate
    {
        [FunctionName("Authenticate")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Function, "get", "post", Route = null)]HttpRequestMessage req,
            [Table(Authentication.TableName)] ICollector<Authentication> collector,
            TraceWriter log
        ) {
            var user = User.PartitionKey;
            log.Info($"Authentication token request received for {user}");

            var auth = new Authentication {
                PartitionKey = Authentication.PartitionName,
                RowKey = Guid.NewGuid().ToString(),
                Identity = user
            };
            collector.Add(auth);

            var result = new SuccessfulAuthenticationJson {
                Token = auth.RowKey
            };
            return req.CreateResponse(HttpStatusCode.OK, result);
        }
    }
}