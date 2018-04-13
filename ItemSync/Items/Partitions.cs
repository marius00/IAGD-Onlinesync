using System;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using System.Linq;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;
using Microsoft.WindowsAzure.Storage;
using System.Collections.Generic;
using System.Diagnostics;
using System.Web.Http;
using ItemSync.Shared.Utility;
using Newtonsoft.Json.Serialization;

namespace ItemSync.Items {
    public static class Partitions {
        [FunctionName("Partitions")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get")]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }


            var itemTable = client.GetTableReference(Partition.TableName);

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partitionKey);
            var exQuery = new TableQuery<Partition>().Where(query);
            var partitions = itemTable.ExecuteQuery(exQuery)
                .Select(p => new PartitionResponse { Partition = p.RowKey.Replace(partitionKey, ""), IsActive = p.IsActive} )
                .ToList();


            log.Info($"A total of {partitions.Count} partitions were returned");
            return req.CreateResponse(HttpStatusCode.OK, partitions, Json.Config);
        }

        public class PartitionResponse {
            public string Partition { get; set; }
            public bool IsActive { get; set; }
        }

    }
}
