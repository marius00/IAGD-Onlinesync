using System;
using System.Linq;
using System.Security.Claims;
using System.Threading;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Model;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;

namespace ItemSync.Items
{
    public static class Authenticate
    {
        [FunctionName("Authenticate")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.System, "get", "post", Route = null)]HttpRequest req,
            [Table(Authentication.TableName)] ICollector<Authentication> collector,
            TraceWriter log
        ) {
#if DEBUG
            var user = User.PartitionKey;
#else
            var user = await GetEmailClaim(req);
#endif
            log.Info($"Authentication token request received for {user}");

            if (string.IsNullOrWhiteSpace(user)) {
                log.Error("Got an empty authentication context");
                return new BadRequestObjectResult("Authentication context is empty");
            }


            var auth = new Authentication {
                PartitionKey = Authentication.PartitionName,
                RowKey = (Guid.NewGuid().ToString() + Guid.NewGuid().ToString()).Replace("-", ""),
                Identity = user
            };
            collector.Add(auth);

            return new RedirectResult("https://auth.iagd.dreamcrash.org/token/" + auth.RowKey);
        }
        
        private static async Task<string> GetEmailClaim(HttpRequest req) {
            var c = req.Headers["cookie"];
            var claims = await User.GetClaims(c);
            var email = claims.user_claims.FirstOrDefault(m => m.typ == ClaimTypes.Email)?.val;
            return email;
        }
    }
}