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
            try {
                var ip = IpUtility.GetClientIp(req);
                log.Info($"User {ip} is attempting to authenticate a token");
                var client = storageAccount.CreateCloudTableClient();
                string partitionKey = Authenticator.Authenticate(client, req);

                if (string.IsNullOrEmpty(partitionKey)) {
                    log.Warning($"{ip}: Authentication failure");
                    return req.CreateResponse(HttpStatusCode.Unauthorized);
                }
                else {
                    log.Info($"{ip}: Authentication success");
                    return req.CreateResponse(HttpStatusCode.OK);
                }
            }
            catch (Exception ex) {
                log.Error("Unhandlex exception while processing request");
                log.Error(ex.Message, ex);
                return req.CreateErrorResponse(HttpStatusCode.InternalServerError, "Internal server error");
            }
        }
    }
}