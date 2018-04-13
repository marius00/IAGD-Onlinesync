using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using Microsoft.Azure;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Items {
    public static class Download {

        [FunctionName("Download")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = null)]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            try {
                var client = storageAccount.CreateCloudTableClient();
                string userKey = Authenticator.Authenticate(client, req);
                if (string.IsNullOrEmpty(userKey)) {
                    return req.CreateResponse(HttpStatusCode.Unauthorized);
                }

                var subPartition = req.GetQueryNameValuePairs()
                    .FirstOrDefault(q => string.Compare(q.Key, "partition", true) == 0)
                    .Value;

                if (string.IsNullOrWhiteSpace(subPartition)) {
                    return req.CreateResponse(
                        HttpStatusCode.BadRequest,
                        "The query parameter \"partition\" empty or missing"
                    );
                }

                log.Info($"User {userKey} has requested an item download for sub partition {subPartition}");
                var itemTable = client.GetTableReference(Item.TableName);

                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal,
                    userKey + subPartition);
                var exQuery = new TableQuery<Item>().Where(query);
                var filteredItems = itemTable.ExecuteQuery(exQuery).Where(m => m.IsActive).ToList();
                log.Info($"A total of {filteredItems.Count} items were returned");


                var deleted = GetDeletedItems(userKey + subPartition, client);
                log.Info($"A total of {deleted.Count} items were marked for deletion");

                //if (filteredItems.Count > 80 && subPartition) {
                //} // TODO: Disable the partition if active

                var result = new DownloadResponse {
                    Items = filteredItems.Select(m => Map(userKey, m)).ToList(),
                    Removed = deleted
                };
                return req.CreateResponse(HttpStatusCode.OK, result);
            }
            catch (Exception ex) {
                log.Error(ex.Message, ex);
                return req.CreateResponse(HttpStatusCode.InternalServerError);
            }
        }

        private static DownloadItemJson Map(string owner, Item item) {
            return new DownloadItemJson {
                Partition = item.PartitionKey.Replace(owner, ""),
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

        private static List<DeletedItemDto> GetDeletedItems(string partition, CloudTableClient client) {
            var table = client.GetTableReference(DeletedItem.TableName);
            table.CreateIfNotExists();

            var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partition);
            var exQuery = new TableQuery<DeletedItem>().Where(query);
            return table.ExecuteQuery(exQuery)
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
    }
}
