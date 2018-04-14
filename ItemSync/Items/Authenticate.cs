using System;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Security.Claims;
using System.Threading;
using System.Threading.Tasks;
using ItemSync.Shared;
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
            [HttpTrigger(AuthorizationLevel.System, "get", "post", Route = null)]HttpRequestMessage req,
            [Table(Authentication.TableName)] ICollector<Authentication> collector,
            TraceWriter log
        ) {
#if DEBUG
            var user = User.PartitionKey;
#else
            var user = GetEmailClaim();
#endif
            log.Info($"Authentication token request received for {user}");

            if (string.IsNullOrWhiteSpace(user)) {
                log.Error("Got an empty authentication context");
                return req.CreateErrorResponse(HttpStatusCode.NoContent, "Authentication context is empty");
            }



            var auth = new Authentication {
                PartitionKey = Authentication.PartitionName,
                RowKey = Guid.NewGuid().ToString(),
                Identity = user
            };
            collector.Add(auth);
            
            var response = req.CreateResponse(HttpStatusCode.Redirect);
            response.Headers.Location = new Uri("https://auth.iagd.dreamcrash.org/token/" + auth.RowKey);

            return response;
        }
        
        private static string GetEmailClaim() {
            var p = Thread.CurrentPrincipal as ClaimsPrincipal;
            return p?.Claims.FirstOrDefault(m => m.Type == ClaimTypes.Email)?.Value ?? string.Empty;
        }
    }
}