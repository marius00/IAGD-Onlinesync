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

namespace ItemSync.Items {
    public static class Upload {
        [FunctionName("Upload")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Function, "post")]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }


            ItemUploadJson data = await req.Content.ReadAsAsync<ItemUploadJson>();
            if (data == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "Could not correctly parse the request parameters");
            }
            else if (data.Items == null && data.Deleted == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "At least one of \"items\" or \"deleted\" must be non-null");
            }
            else if (data.Items?.Count == 0 && data.Deleted?.Count == 0) {
                return req.CreateResponse(HttpStatusCode.OK);
            }


            log.Info($"Received a request from {partitionKey} to upload {data.Items?.Count ?? 0} items and remove {data.Deleted?.Count ?? 0} items");
            var itemTable = client.GetTableReference(Item.TableName);

            if (data.Items?.Count > 0) {
                var newItems = data.Items.Select(m => ItemBuilder.Create(partitionKey, m.Data));
                var numBatchOperations = Insert(partitionKey, itemTable, newItems);
                log.Info($"Inserted {data.Items} items over {numBatchOperations} batches");
            }

            if (data.Deleted?.Count > 0) {
                var numBatchOperations = Delete(partitionKey, itemTable, data.Deleted);
                log.Info($"Marked {data.Deleted.Count} items as deleted over {numBatchOperations} batches");
            }

            return req.CreateResponse(HttpStatusCode.OK);
        }

        private static int Insert(string partitionKey, CloudTable itemTable, IEnumerable<Item> items) {
            int numOperations = 0;
            var batch = new TableBatchOperation();
            foreach (var item in items) {
                batch.Add(TableOperation.Insert(item));

                if (batch.Count == 100) {
                    itemTable.ExecuteBatch(batch);
                    batch.Clear();
                    numOperations++;
                }
            }

            if (batch.Count > 0) {
                itemTable.ExecuteBatch(batch);
                numOperations++;
            }

            return numOperations;
        }

        private static int Delete(string partitionKey, CloudTable itemTable, List<string> itemKeys) {
            int numOperations = 0;
            var batch = new TableBatchOperation();
            foreach (var itemKey in itemKeys) {
                var entity = new DynamicTableEntity(partitionKey, itemKey);
                entity.ETag = "*";
                entity.Properties.Add("IsActive", new EntityProperty(false));
                batch.Add(TableOperation.Merge(entity));

                if (batch.Count == 100) {
                    itemTable.ExecuteBatch(batch);
                    batch.Clear();
                    numOperations++;
                }
            }

            if (batch.Count > 0) {
                itemTable.ExecuteBatch(batch);
                numOperations++;
            }

            return numOperations;
        }
    }
}
