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

namespace ItemSync.Items.v3 {
    public static class DownloadV3 {
        public static string Combine(string userKey, string partitionKey) {
            if (userKey.EndsWith("-") || partitionKey.StartsWith("-"))
                return userKey + partitionKey;
            else {
                return userKey + "-" + partitionKey;
            }
        }

        [FunctionName("v3_Download")]
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
                var itemTable = client.GetTableReference(ItemV2.TableName);

                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, Combine(userKey, subPartition));
                var exQuery = new TableQuery<ItemV2>().Where(query);

                var unfilteredItems = await QueryHelper.ListAll(itemTable, exQuery);
                var filteredItems = unfilteredItems.Where(m => m.IsActive).ToList();
                log.Info($"A total of {filteredItems.Count} items were returned");

                // Where timestamp is older than say a week.. and there are less than say 30 items..
                var old = DateTimeOffset.Now.AddDays(-7);
                var isOutdated = unfilteredItems.Any(item => item.Timestamp < old);

                var deleted = await GetDeletedItems(Combine(userKey, subPartition), client);
                log.Info($"A total of {deleted.Count} items were marked for deletion");

                var disableNow = filteredItems.Count > 80 || isOutdated;
                if (disableNow) {
                    log.Info($"Disabling partition {subPartition} for {userKey} (may already be disabled)");
                    var table = client.GetTableReference(PartitionV2.TableName);
                    await table.CreateIfNotExistsAsync();
                    DisablePartition(userKey, subPartition, table);
                }

                var result = new DownloadResponse {
                    Items = filteredItems.Select(m => Map(userKey, m)).ToList(),
                    Removed = deleted,
                    DisableNow = disableNow
                };

                return new OkObjectResult(result);
            }
            catch (Exception ex) {
                log.Error(ex.Message, ex);
                return new ExceptionResult(ex, false);
            }
        }

        private static DownloadItemJson Map(string owner, ItemV2 item) {
            
            return new DownloadItemJson {
                Partition = item.PartitionKey.Substring(owner.Length + 1), // Remove owner@email and the "-" suffix
                Id = item.RowKey,
                BaseRecord = item.BaseRecord,
                EnchantmentRecord = item.EnchantmentRecord,
                EnchantmentSeed = item.EnchantmentSeed,
                IsHardcore = item.IsHardcore,
                MateriaCombines = item.MateriaCombines,
                MateriaRecord = item.MateriaRecord,
                Mod = item.Mod,
                ModifierRecord = item.ModifierRecord,
                PrefixRecord = item.PrefixRecord,
                RelicCompletionBonusRecord = item.RelicCompletionBonusRecord,
                RelicSeed = item.RelicSeed,
                Seed = item.Seed,
                StackCount = item.StackCount,
                SuffixRecord = item.SuffixRecord,
                TransmuteRecord = item.TransmuteRecord
            };
        }

        private static async Task<List<DeletedItemDto>> GetDeletedItems(string partition, CloudTableClient client) {
            var table = client.GetTableReference(DeletedItemV2.TableName);
            await table.CreateIfNotExistsAsync();

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partition);
            var exQuery = new TableQuery<DeletedItemV2>().Where(query);


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
            public bool DisableNow { get; set; }
        }

        private static async void DisablePartition(string owner, string rowkey, CloudTable table) {
            var entity = new DynamicTableEntity(owner, Combine(owner, rowkey));
            entity.ETag = "*";
            entity.Properties.Add("IsActive", new EntityProperty(false));
            await table.ExecuteAsync(TableOperation.Merge(entity));
        }
    }
}
