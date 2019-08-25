using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.AzureCloudTable;
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
    public static class DeleteV3 {
        /// <summary>
        /// Deletes an entire user account
        /// </summary>
        [FunctionName("v3_Delete")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post")]
            HttpRequest req,
            [StorageAccount("StorageConnectionString")]
            CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = await Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return new UnauthorizedResult();
            }

            var partitions = await PartionUtility.GetAllPartitions(log, client, partitionKey);
            log.Info($"Received a request from {partitionKey} to erase the account");

            // Foreach partition: Delete all items
            var itemTable = client.GetTableReference(ItemV2.TableName);
            foreach (var partition in partitions) {
                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partition.RowKey);
                var exQuery = new TableQuery<ItemV2>().Where(query);
                var items = await QueryHelper.ListAll(itemTable, exQuery);

                log.Info($"Deleting stored items in partition {partition.RowKey} for {partitionKey}");
                BatchDelete(log, itemTable, partition.RowKey, items);
            }

            // Delete the 'removed items' entries
            foreach (var partition in partitions) {
                log.Info($"Deleting removed items in partition {partition.RowKey} for {partitionKey}");
                DeleteDeletedItems(log, client, partition.RowKey);
            }

            // Delete the partition entries
            var partitionTable = client.GetTableReference(PartitionV2.TableName);
            foreach (var partition in partitions) {
                log.Info($"Deleting partition {partition.RowKey} for {partitionKey}");
                await partitionTable.ExecuteAsync(TableOperation.Delete(partition));
            }

            // Delete auth entry used for logins
            try {
                var table = client.GetTableReference(Authentication.TableName);
                await table.CreateIfNotExistsAsync();

                var query = TableQuery.CombineFilters(
                    TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, Authentication.PartitionName),
                    TableOperators.And,
                    TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, req.Headers["Simple-Auth"])
                );

                var exQuery = new TableQuery<Authentication>().Where(query);
                var authenticationKeys = await QueryHelper.ListAll(table, exQuery);
                BatchDelete(log, table, Authentication.PartitionName, authenticationKeys);
            }
            catch (Exception ex) {
                log.Warning("Error deleting auth entry", ex.StackTrace);
            }

            return new OkObjectResult("");
        }


        private static async void BatchDelete<T>(TraceWriter log, CloudTable table, string partitionKey, List<T> items) where T : ITableEntity {
            TableBatchOperation operation = new TableBatchOperation();
            foreach (var item in items) {
                operation.Delete(item);

                if (operation.Count >= 100) {
                    await table.ExecuteBatchAsync(operation);
                    operation = new TableBatchOperation();
                    log.Info($"Executing deletion batch on 100 items for {partitionKey}");
                }
            }

            // Remove any remaining items, outside of the 100s batches.
            if (operation.Count > 0) {
                log.Info($"Executing deletion batch on {operation.Count} items for {partitionKey}");
                await table.ExecuteBatchAsync(operation);
            }
        }

        /// <summary>
        /// Remove the entries for "deleted items"
        /// Items which are marked for deletion when IA syncs.
        /// </summary>
        private static async void DeleteDeletedItems(TraceWriter log, CloudTableClient client, string partition) {
            var table = client.GetTableReference(DeletedItemV2.TableName);
            await table.CreateIfNotExistsAsync();

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partition);
            var exQuery = new TableQuery<DeletedItemV2>().Where(query);

            var unfiltered = await QueryHelper.ListAll(table, exQuery);
            BatchDelete(log, table, partition, unfiltered);
        }
    }
}