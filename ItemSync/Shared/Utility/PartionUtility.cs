using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using ItemSync.Shared.AzureCloudTable;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Utility {
    static class PartionUtility {
        /// <summary>
        /// Get the current active upload partition, or create a new one
        /// </summary>
        /// <param name="client"></param>
        /// <param name="owner"></param>
        /// <returns></returns>
        public static async Task<PartitionV1> GetUploadPartition(TraceWriter log, CloudTableClient client, string owner) {
            try {
                var table = client.GetTableReference(PartitionV1.TableName);
                await table.CreateIfNotExistsAsync();


                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, owner);
                var exQuery = new TableQuery<PartitionV1>().Where(query);
                var unfiltered = await QueryHelper.ListAll(table, exQuery);
                var filteredItems = unfiltered.Where(m => m.IsActive);

                var active = filteredItems.FirstOrDefault();
                if (active != null) {
                    return active;
                }
                else {
                    PartitionV1 p = new PartitionV1 {
                        PartitionKey = owner,
                        RowKey = $"{owner}-{DateTimeOffset.UtcNow.ToString("yyyy-MM")}-{Guid.NewGuid().ToString().Replace("-", "")}",
                        IsActive = true
                    };

                    await table.ExecuteAsync(TableOperation.Insert(p));
                    return p;
                }
            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.ToString());
                throw ex;
            }
        }

        public static async Task<PartitionV2> GetUploadPartitionV2(TraceWriter log, CloudTableClient client, string owner) {
            try {
                var table = client.GetTableReference(PartitionV2.TableName);
                await table.CreateIfNotExistsAsync();


                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, owner);
                var exQuery = new TableQuery<PartitionV2>().Where(query);
                var unfiltered = await QueryHelper.ListAll(table, exQuery);
                var filteredItems = unfiltered.Where(m => m.IsActive);

                var active = filteredItems.FirstOrDefault();
                if (active != null) {
                    return active;
                }
                else {
                    PartitionV2 p = new PartitionV2 {
                        PartitionKey = owner,
                        RowKey = $"{owner}-{DateTimeOffset.UtcNow.ToString("yyyy-MM")}-{Guid.NewGuid().ToString().Replace("-", "")}",
                        IsActive = true
                    };

                    await table.ExecuteAsync(TableOperation.Insert(p));
                    return p;
                }
            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.ToString());
                throw;
            }
        }


        public static async Task<PartitionV2> GetUploadPartitionV3(TraceWriter log, CloudTableClient client, string owner, string partition) {
            try {
                var table = client.GetTableReference(PartitionV2.TableName);
                await table.CreateIfNotExistsAsync();

                var query = TableQuery.CombineFilters(
                    TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, owner),
                    TableOperators.And,
                    TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, $"{owner}-{partition}")
                );

                var exQuery = new TableQuery<PartitionV2>().Where(query);
                var partitions = await QueryHelper.ListAll(table, exQuery);

                if (partitions.Count > 0) {
                    return partitions.FirstOrDefault();
                }

                PartitionV2 p = new PartitionV2 {
                    PartitionKey = owner,
                    RowKey = $"{owner}-{partition}",
                    IsActive = true
                };

                await table.ExecuteAsync(TableOperation.Insert(p));
                return p;
            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.ToString());
                throw;
            }
        }

        public static async Task<List<PartitionV2>> GetAllPartitions(TraceWriter log, CloudTableClient client, string owner) {
            try {
                var table = client.GetTableReference(PartitionV2.TableName);
                await table.CreateIfNotExistsAsync();

                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, owner);
                var exQuery = new TableQuery<PartitionV2>().Where(query);
                var partitions = await QueryHelper.ListAll(table, exQuery);

                return partitions;
            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.ToString());
                throw;
            }
        }
    }
}