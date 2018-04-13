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
using ItemSync.Shared.Utility;

namespace ItemSync.Items {
    public static class Remove {
        [FunctionName("Remove")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post")]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }


            List<ItemToRemove> data = await req.Content.ReadAsAsync<List<ItemToRemove>>();
            if (data == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "Could not correctly parse the request parameters");
            }
            else if (data?.Count <= 0) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "No items to delete");
            }

            log.Info($"Received a request from {partitionKey} to remove {data.Count} items");
            var itemTable = client.GetTableReference(Item.TableName);

            var numBatchOperations = DeleteByPartition(partitionKey, itemTable, data);
            log.Info($"Marked {data.Count} items as deleted over {numBatchOperations} batches");


            var partition = PartionUtility.GetUploadPartition(log, client, partitionKey);
            var deletedItemsTable = client.GetTableReference(DeletedItem.TableName);
            DeleteInActivePartition(partition.RowKey, deletedItemsTable, data);
            log.Info($"Marked {data.Count} items as deleted in the active partition {partition.RowKey}");

            return req.CreateResponse(HttpStatusCode.OK);
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
        private static void DeleteInActivePartition(string activePartitionKey, CloudTable itemTable, List<ItemToRemove> itemKeys) {
            var batch = new TableBatchOperation();
            foreach (var itemKey in itemKeys) {
                batch.Add(TableOperation.InsertOrReplace(new DeletedItem {
                    PartitionKey = activePartitionKey,
                    RowKey = Guid.NewGuid().ToString(),
                    ItemPartitionKey = itemKey.Partition,
                    ItemRowKey = itemKey.Id
                }));


                if (batch.Count == 100) {
                    itemTable.ExecuteBatch(batch);
                    batch.Clear();
                }
            }


            if (batch.Count > 0) {
                itemTable.ExecuteBatch(batch);
            }
        }

        /// <summary>
        /// Mark the items as deleted in their residing partitions (this prevents download on fresh syncs)
        /// </summary>
        /// <param name="partitionKey"></param>
        /// <param name="itemTable"></param>
        /// <param name="itemKeys"></param>
        /// <returns></returns>
        private static int DeleteByPartition(string partitionKey, CloudTable itemTable, List<ItemToRemove> itemKeys) {
            int numOperations = 0;
            var batch = new TableBatchOperation();

            // Sorted by partition for batching
            itemKeys.Sort((a, b) => string.Compare(a.Partition, b.Partition));
            var previousPartition = itemKeys[0].Partition;

            foreach (var itemKey in itemKeys) {
                // No batches across partitions
                if (batch.Count == 100 || previousPartition != itemKey.Partition) {
                    itemTable.ExecuteBatch(batch);
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
                itemTable.ExecuteBatch(batch);
                numOperations++;
            }

            return numOperations;
        }
    }
}
