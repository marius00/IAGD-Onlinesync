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
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items
{
    public static class VerifyToken {
        [FunctionName("VerifyToken")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = null)]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);

            var ip = IpUtility.GetClientIp(req);
            if (string.IsNullOrEmpty(partitionKey)) {
                log.Warning($"{ip}: Authentication failure");
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }
            else {
                log.Info($"{ip}: Authentication success");
                return req.CreateResponse(HttpStatusCode.OK);
            }
        }
    }
}