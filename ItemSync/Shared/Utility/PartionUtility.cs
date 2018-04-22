using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using ItemSync.Shared.AzureCloudTable;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Utility
{
    static class PartionUtility {


        /// <summary>
        /// Get the current active upload partition, or create a new one
        /// </summary>
        /// <param name="client"></param>
        /// <param name="owner"></param>
        /// <returns></returns>
        public static async Task<Partition> GetUploadPartition(TraceWriter log, CloudTableClient client, string owner) {
            try {
                var table = client.GetTableReference(Partition.TableName);
                await table.CreateIfNotExistsAsync();


                var query = TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, owner);
                var exQuery = new TableQuery<Partition>().Where(query);
                var unfiltered = await QueryHelper.ListAll(table, exQuery);
                var filteredItems = unfiltered.Where(m => m.IsActive);

                var active = filteredItems.FirstOrDefault();
                if (active != null) {
                    return active;
                }
                else {
                    Partition p = new Partition {
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
    }
}
