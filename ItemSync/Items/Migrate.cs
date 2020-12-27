using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using System.Web.Http;
using ItemSync.Shared.AzureCloudTable;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.AspNetCore.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Items {
    public static class Migrate {

        [FunctionName("Migrate")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post", Route = null)]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            try {
                var client = storageAccount.CreateCloudTableClient();
                string userKey = await Authenticator.Authenticate(client, req);
                if (string.IsNullOrEmpty(userKey)) {
                    return new UnauthorizedResult();
                }
                
                var result = new MigrateResponse {
                    User = userKey
                };

                return new OkObjectResult(result);
            }
            catch (Exception ex) {
                log.Error(ex.Message, ex);
                return new ExceptionResult(ex, false);
            }
        }

        public class MigrateResponse {
            public string User { get; set; }
        }
    }
}
