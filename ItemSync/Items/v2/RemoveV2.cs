using System;
using System.Threading.Tasks;
using ItemSync.Shared;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;
using Microsoft.WindowsAzure.Storage;
using System.Collections.Generic;
using ItemSync.Shared.Model;
using ItemSync.Shared.Utility;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Newtonsoft.Json;

namespace ItemSync.Items.v2 {
    public static class RemoveV2 {
        [FunctionName("v2_Remove")]
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
            var data = JsonConvert.DeserializeObject<List<ItemToRemove>>(json);
            
            if (data == null) {
                return new BadRequestObjectResult("Could not correctly parse the request parameters");
            }
            else if (data?.Count <= 0) {
                return new BadRequestObjectResult("No items to delete");
            }

            log.Info($"Received a request from {partitionKey} to remove {data.Count} items");
            var itemTable = client.GetTableReference(ItemV2.TableName);

            DeleteByPartition(partitionKey, itemTable, data, log);


            var partition = await PartionUtility.GetUploadPartition(log, client, partitionKey);
            var deletedItemsTable = client.GetTableReference(DeletedItemV2.TableName);
            DeleteInActivePartition(partition.RowKey, deletedItemsTable, data);
            log.Info($"Marked {data.Count} items as deleted in the active partition {partition.RowKey}");

            return new OkResult();
        }

        public class ItemToRemove {
            public string Id { get; set; }
            public string Partition { get; set; }
        }

        /// <summary>
        /// Mark the items as deleted in the active partition (this updates day-to-day users)
        /// </summary>
        /// <param name="activePartitionKey"></param>
        /// <param name="itemTable"></param>
        /// <param name="itemKeys"></param>
        private static async void DeleteInActivePartition(string activePartitionKey, CloudTable itemTable, List<ItemToRemove> itemKeys) {
            var batch = new TableBatchOperation();
            foreach (var itemKey in itemKeys) {
                batch.Add(TableOperation.InsertOrReplace(new DeletedItemV2 {
                    PartitionKey = activePartitionKey,
                    RowKey = Guid.NewGuid().ToString(),
                    ItemPartitionKey = itemKey.Partition,
                    ItemRowKey = itemKey.Id
                }));


                if (batch.Count == 100) {
                    await itemTable.ExecuteBatchAsync(batch);
                    batch.Clear();
                }
            }


            if (batch.Count > 0) {
                await itemTable.ExecuteBatchAsync(batch);
            }
        }

        /// <summary>
        /// Mark the items as deleted in their residing partitions (this prevents download on fresh syncs)
        /// </summary>
        /// <param name="partitionKey"></param>
        /// <param name="itemTable"></param>
        /// <param name="itemKeys"></param>
        /// <returns></returns>
        private static async void DeleteByPartition(string partitionKey, CloudTable itemTable, List<ItemToRemove> itemKeys, TraceWriter log) {
            int numOperations = 0;
            var batch = new TableBatchOperation();

            // Sorted by partition for batching
            itemKeys.Sort((a, b) => string.Compare(a.Partition, b.Partition));
            var previousPartition = itemKeys[0].Partition;

            foreach (var itemKey in itemKeys) {
                // No batches across partitions
                if (batch.Count == 100 || previousPartition != itemKey.Partition) {
                    await itemTable.ExecuteBatchAsync(batch);
                    batch.Clear();
                    numOperations++;
                }

                var entity = new DynamicTableEntity(partitionKey + itemKey.Partition, itemKey.Id);
                entity.ETag = "*";
                entity.Properties.Add("IsActive", new EntityProperty(false));
                batch.Add(TableOperation.Merge(entity));
                previousPartition = itemKey.Partition;

            }

            if (batch.Count > 0) {
                await itemTable.ExecuteBatchAsync(batch);
                numOperations++;
            }


            log.Info($"Marked {itemKeys.Count} items as deleted over {numOperations} batches");
        }
    }
}
