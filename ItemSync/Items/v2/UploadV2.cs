using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using ItemSync.Shared.Utility;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;
using Newtonsoft.Json;

namespace ItemSync.Items.v2 {
    public static class UploadV2 {
        [FunctionName("v2_Upload")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post")]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = await Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return new UnauthorizedResult();
            }

            string json = await req.ReadAsStringAsync();
            var items = JsonConvert.DeserializeObject<List<ItemJson>>(json);
            
            if (items == null) {
                return new BadRequestObjectResult("Could not correctly parse the request parameters");
            }
            else if (items.Count > 100) {
                return new BadRequestObjectResult("Too many items to store, maximum 100 per call");
            }
            else if (items.Count <= 0) {
                return new BadRequestObjectResult("No items to store");
            }
            
            var partition = await PartionUtility.GetUploadPartitionV2(log, client, partitionKey);
            log.Info($"Received a request from {partitionKey} to upload {items.Count} items to {partition.RowKey}");

            var itemTable = client.GetTableReference(ItemV2.TableName);
            await itemTable.CreateIfNotExistsAsync();
            var itemMapping = await Insert(partitionKey, partition.RowKey, itemTable, items);
            log.Info($"Inserted {items.Count} items over {itemMapping.Count} batches");

            bool shouldClosePartition = items.Count >= 90;
            if (shouldClosePartition) {
                log.Info($"Closing partition {partition.PartitionKey}, due to received count at {items.Count}");
                partition.IsActive = false;
                await client.GetTableReference(PartitionV2.TableName).ExecuteAsync(TableOperation.Replace(partition));
            }

            var result = new UploadResultDto {
                Partition = itemMapping.First().Partition,
                IsClosed = shouldClosePartition,
                Items = itemMapping
            };
            return new OkObjectResult(result);
        }

        class UploadResultDto {
            public string Partition { get; set; }
            public bool IsClosed { get; set; }
            public List<UploadResultItem> Items { get; set; }
        }


        class UploadResultItem {
            public long LocalId { get; set; }
            public string Id { get; set; }
            public string Partition { get; set; }
        }

        private static async Task<List<UploadResultItem>> Insert(string owner, string partition, CloudTable itemTable, List<ItemJson> items) {
            Debug.Assert(items.Count <= 100);
            var ouputKey = partition.Remove(0, owner.Length);

            var batch = new TableBatchOperation();
            foreach (var item in items) {
                batch.Add(TableOperation.Insert(ItemBuilder.CreateV2(partition, item)));
            }

            var mapping = new List<UploadResultItem>();
            var results = await itemTable.ExecuteBatchAsync(batch);
            for (int i = 0; i < items.Count; i++) {
                var saved = results[i].Result as ItemV2;
                mapping.Add(new UploadResultItem {
                    LocalId = items[i].LocalId,
                    Id = saved.RowKey,
                    Partition = ouputKey // The email/whatever is only used to represent the item internally
                });
            }
            

            return mapping;
        }
        
    }
}
