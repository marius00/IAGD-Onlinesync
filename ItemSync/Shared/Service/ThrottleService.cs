using System;
using System.Collections.Generic;
using System.Text;
using System.Threading.Tasks;
using ItemSync.Shared.AzureCloudTable;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Service
{
    class ThrottleService {
        private const int ThrottleLimit = 5;
        private readonly TraceWriter _log;
        private readonly CloudTable _table;

        public ThrottleService(CloudTableClient client, TraceWriter log) {
            _table = client.GetTableReference(ThrottleEntry.TableName);
            _log = log;
        }

        public async void Clear() {
            await _table.DeleteIfExistsAsync();
        }

        public async Task<bool> ThrottleOrIncrement(string key) {
            await _table.CreateIfNotExistsAsync();

            var query = TableQuery.CombineFilters(
                TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, "any"),
                TableOperators.And,
                TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, key)
            );

            var exQuery = new TableQuery<ThrottleEntry>().Where(query);
            var entry = await QueryHelper.Get(_table, exQuery);
            if (entry == null) {
                // Create
                entry = new ThrottleEntry {
                    PartitionKey = "any",
                    RowKey = key,
                    Count = 1
                };

                await _table.ExecuteAsync(TableOperation.Insert(entry));
                _log.Info($"Throttle count for ${key} initialized to {entry.Count}");
                return false;
            }
            else {
                if (entry.Count >= ThrottleLimit) {
                    _log.Info($"Throttle count for ${key} has been reached, throttled");
                    return true;
                }
                else {
                    entry.Count++;
                    await _table.ExecuteAsync(TableOperation.Replace(entry));
                    _log.Info($"Throttle count for ${key} increased to {entry.Count}");
                    return false;
                }
            }
        }
    }
}
