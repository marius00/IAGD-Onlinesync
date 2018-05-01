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
    public static class Download {

        [FunctionName("Download")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = null)]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            try {
                var client = storageAccount.CreateCloudTableClient();
                string userKey = await Authenticator.Authenticate(client, req);
                if (string.IsNullOrEmpty(userKey)) {
                    return new UnauthorizedResult();
                }
                
                var subPartition = req.Query["partition"];

                if (string.IsNullOrWhiteSpace(subPartition)) {
                    return new BadRequestObjectResult("The query parameter \"partition\" empty or missing");
                }

                log.Info($"User {userKey} has requested an item download for sub partition {subPartition}");
                var itemTable = client.GetTableReference(ItemV1.TableName);

                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal,
                    userKey + subPartition);
                var exQuery = new TableQuery<ItemV1>().Where(query);

                var unfilteredItems = await QueryHelper.ListAll(itemTable, exQuery);
                var filteredItems = unfilteredItems.Where(m => m.IsActive).ToList();
                log.Info($"A total of {filteredItems.Count} items were returned");


                var deleted = await GetDeletedItems(userKey + subPartition, client);
                log.Info($"A total of {deleted.Count} items were marked for deletion");

                if (filteredItems.Count > 80) {
                    log.Info($"Disabling partition {subPartition} for {userKey} (may already be disabled)");
                    var table = client.GetTableReference(PartitionV1.TableName);
                    await table.CreateIfNotExistsAsync();
                    DisablePartition(userKey, subPartition, table);
                }

                var result = new DownloadResponse {
                    Items = filteredItems.Select(m => Map(userKey, m)).ToList(),
                    Removed = deleted
                };

                return new OkObjectResult(result);
            }
            catch (Exception ex) {
                log.Error(ex.Message, ex);
                return new ExceptionResult(ex, false);
            }
        }

        private static DownloadItemJson Map(string owner, ItemV1 itemV1) {
            return new DownloadItemJson {
                Partition = itemV1.PartitionKey.Replace(owner, ""),
                Id = itemV1.RowKey,
                BaseRecord = itemV1.BaseRecord,
                EnchantmentRecord = itemV1.EnchantmentRecord,
                EnchantmentSeed = itemV1.EnchantmentSeed,
                IsHardcore = itemV1.IsHardcore,
                MateriaCombines = itemV1.MateriaCombines,
                MateriaRecord = itemV1.MateriaRecord,
                Mod = itemV1.Mod,
                ModifierRecord = itemV1.ModifierRecord,
                PrefixRecord = itemV1.PrefixRecord,
                RelicCompletionBonusRecord = itemV1.RelicCompletionBonusRecord,
                RelicSeed = itemV1.RelicSeed,
                Seed = itemV1.Seed,
                StackCount = itemV1.StackCount,
                SuffixRecord = itemV1.SuffixRecord,
                TransmuteRecord = itemV1.TransmuteRecord
            };
        }

        private static async Task<List<DeletedItemDto>> GetDeletedItems(string partition, CloudTableClient client) {
            var table = client.GetTableReference(DeletedItemV1.TableName);
            await table.CreateIfNotExistsAsync();

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partition);
            var exQuery = new TableQuery<DeletedItemV1>().Where(query);


            var unfiltered = await QueryHelper.ListAll(table, exQuery);
            return unfiltered
                .Select(m => new DeletedItemDto { Partition = m.ItemPartitionKey , Id = m.ItemRowKey })
                .ToList();
        }

        public class DeletedItemDto {
            public string Id { get; set; }
            public string Partition { get; set; }
        }
        public class DownloadResponse {
            public List<DownloadItemJson> Items { get; set; }
            public List<DeletedItemDto> Removed { get; set; }
        }



        private static async void DisablePartition(string owner, string rowkey, CloudTable table) {
            var entity = new DynamicTableEntity(owner, owner + rowkey);
            entity.ETag = "*";
            entity.Properties.Add("IsActive", new EntityProperty(false));
            await table.ExecuteAsync(TableOperation.Merge(entity));
        }
    }
}
