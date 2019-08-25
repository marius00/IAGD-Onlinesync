using System.Threading.Tasks;
using System.Linq;
using ItemSync.Shared;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;
using Microsoft.WindowsAzure.Storage;
using ItemSync.Shared.AzureCloudTable;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace ItemSync.Items.v3 {
    public static class PartitionsV3 {
        [FunctionName("v3_Partitions")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get")]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = await Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return new UnauthorizedResult();
            }


            var itemTable = client.GetTableReference(PartitionV2.TableName);
            await itemTable.CreateIfNotExistsAsync();

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partitionKey);
            var exQuery = new TableQuery<PartitionV2>().Where(query);

            var domainPartitions = await QueryHelper.ListAll(itemTable, exQuery);
            var partitions = domainPartitions
                .Select(p => new PartitionResponse { Partition = p.RowKey.Replace($"{partitionKey}-", ""), IsActive = p.IsActive} )
                .ToList();


            log.Info($"A total of {partitions.Count} partitions were returned");
            return new OkObjectResult(partitions);
        }

        public class PartitionResponse {
            public string Partition { get; set; }
            public bool IsActive { get; set; }
        }

    }
}
