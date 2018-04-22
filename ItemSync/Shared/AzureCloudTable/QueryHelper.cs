using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.AzureCloudTable {
    static class QueryHelper {
        public static async Task<T> Get<T>(CloudTable table, TableQuery<T> query) where T : ITableEntity, new() {
            var re = new T();
            TableContinuationToken continuationToken = null;
            do {
                var entries = await table.ExecuteQuerySegmentedAsync(query, continuationToken);
                re = entries.FirstOrDefault();
                continuationToken = entries.ContinuationToken;
            } while (continuationToken != null);

            return re;
        }

        public static async Task<List<T>> ListAll<T>(CloudTable table, TableQuery<T> query) where T : ITableEntity, new() {
            var re = new T();
            List<T> result = new List<T>();
            TableContinuationToken continuationToken = null;
            do {
                var entries = await table.ExecuteQuerySegmentedAsync(query, continuationToken);
                result.AddRange(entries);
                
                continuationToken = entries.ContinuationToken;
            } while (continuationToken != null);

            return result;
        }
    }
}
