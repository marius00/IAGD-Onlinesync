using System;
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

namespace ItemSync.Items.v3 {
    public static class UploadV3 {
        [FunctionName("v3_Upload")]
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
            var items = JsonConvert.DeserializeObject<List<ItemJsonV3>>(json);
            
            if (items == null) {
                return new BadRequestObjectResult("Could not correctly parse the request parameters");
            }
            else if (items.Count > 100) {
                return new BadRequestObjectResult("Too many items to store, maximum 100 per call");
            }
            else if (items.Count <= 0) {
                return new BadRequestObjectResult("No items to store");
            } else if (items.Select(item => item.RemotePartition).Distinct().Count() > 1) {
                return new BadRequestObjectResult("Cannot store items from multiple partitions in a single call");
            }
            
            // Just ensuring that a partition entry exists, so that we can find it again later.
            var partition = await PartionUtility.GetUploadPartitionV3(log, client, partitionKey, items.First().RemotePartition);
            log.Info($"Received a request from {partitionKey} to upload {items.Count} items to {partition.RowKey}");



            var itemTable = client.GetTableReference(ItemV2.TableName);
            await itemTable.CreateIfNotExistsAsync();
            var itemMapping = await Insert(log, partition.RowKey, itemTable, items);
            log.Info($"Inserted {items.Count} items");

            bool shouldClosePartition = items.Count >= 90;
            if (shouldClosePartition) {
                log.Info($"Closing partition {partition.PartitionKey}, due to received count at {items.Count}");
                partition.IsActive = false;
                await client.GetTableReference(PartitionV2.TableName).ExecuteAsync(TableOperation.Replace(partition));
            }

            var result = new UploadResultDto {
                Partition = partition.RowKey.Remove(0, partitionKey.Length), // Strip the owner email from the partition
                IsClosed = shouldClosePartition,
            };
            return new OkObjectResult(result);
        }

        class UploadResultDto {
            public string Partition { get; set; }
            public bool IsClosed { get; set; }
        }

        private static async Task<bool> Insert(TraceWriter log, string partition, CloudTable itemTable, List<ItemJsonV3> items) {
            Debug.Assert(items.Count <= 100);

            var batch = new TableBatchOperation();
            foreach (var item in items) {
                batch.Add(TableOperation.Insert(ItemBuilder.CreateV3(partition, item.RemoteId, item)));
            }

            
            try {
                await itemTable.ExecuteBatchAsync(batch);
            }
            catch (StorageException ex) {
                var ri = ex.RequestInformation;
                if (ri.HttpStatusMessage == "0:The specified entity already exists." && ri.HttpStatusCode == 0x199) {
                    log.Info("Items already stored, skipping insert and returning 200 OK");
                    return true;
                }
                return false;

            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.StackTrace);
                return false;
            }

            return true;
        }
        
    }
}
