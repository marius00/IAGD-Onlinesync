using System;
using System.Threading.Tasks;
using System.Web.Http;
using ItemSync.Shared;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items
{
    public static class VerifyToken {
        [FunctionName("VerifyToken")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = null)]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            try {
                var ip = IpUtility.GetClientIp(req);
                log.Info($"User {ip} is attempting to authenticate a token");
                var client = storageAccount.CreateCloudTableClient();
                string partitionKey = await Authenticator.Authenticate(client, req);

                if (string.IsNullOrEmpty(partitionKey)) {
                    log.Warning($"{ip}: Authentication failure");
                    return new UnauthorizedResult();
                }
                else {
                    log.Info($"{ip}: Authentication success");
                    return new OkResult();
                }
            }
            catch (Exception ex) {
                log.Error("Unhandlex exception while processing request");
                log.Error(ex.Message, ex);
                return new InternalServerErrorResult();
            }
        }
    }
}