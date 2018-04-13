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
using ItemSync.Shared.Utility;

namespace ItemSync.Items {
    public static class Upload {
        [FunctionName("Upload")]
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


            List<ItemJson> items = await req.Content.ReadAsAsync<List<ItemJson>>();
            if (items == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "Could not correctly parse the request parameters");
            }
            else if (items?.Count > 100) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "Too many items to store, maximum 100 per call");
            }
            else if (items?.Count <= 0) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "No items to store");
            }
            
            var partition = PartionUtility.GetUploadPartition(log, client, partitionKey);
            log.Info($"Received a request from {partitionKey} to upload {items.Count} items to {partition.RowKey}");

            var itemTable = client.GetTableReference(Item.TableName);
            await itemTable.CreateIfNotExistsAsync();
            var itemMapping = Insert(partitionKey, partition.RowKey, itemTable, items);
            log.Info($"Inserted {items} items over {itemMapping} batches");


            // Update the partition reference
            // TODO TODO TODO: This is
            // TODO TODO TODO: This is
            // TODO TODO TODO: This is
            // TODO TODO TODO: This is best handled on download, if partition starstwith today, and items > x, then untag it as active.
            // That way partition can grow to any size, but stops growing once "done" with large syncups
            /*
            partition.NumItems += items.Count;
            partition.IsActive = partition.NumItems < 200;
            var table = client.GetTableReference(Partition.TableName);
            table.Execute(TableOperation.Replace(partition));
            */

            return req.CreateResponse(HttpStatusCode.OK, itemMapping);
        }


        class UploadResultItem {
            public long LocalId { get; set; }
            public string Id { get; set; }
            public string Partition { get; set; }
        }

        private static List<UploadResultItem> Insert(string owner, string partition, CloudTable itemTable, List<ItemJson> items) {
            Debug.Assert(items.Count <= 100);
            var ouputKey = partition.Remove(0, owner.Length);

            var batch = new TableBatchOperation();
            foreach (var item in items) {
                batch.Add(TableOperation.Insert(ItemBuilder.Create(partition, item)));
            }

            var mapping = new List<UploadResultItem>();
            var results = itemTable.ExecuteBatch(batch);
            for (int i = 0; i < items.Count; i++) {
                var saved = results[i].Result as Item;
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
