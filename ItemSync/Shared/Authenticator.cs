using System;
using System.Collections.Generic;
using ItemSync.Shared.Model;
using Microsoft.WindowsAzure.Storage.Table;
using System.Linq;
using System.Net.Http;
using System.Threading.Tasks;
using System.Web;
using ItemSync.Shared.AzureCloudTable;
using Microsoft.AspNetCore.Http;

namespace ItemSync.Shared {
    public static class Authenticator {
        public static async Task<string> Authenticate(CloudTableClient client, HttpRequest request) {
#if DEBUG
            return "localdev@example.com";
#endif

            string key;
            if (request.Headers.ContainsKey("Simple-Auth")) {
                key = request.Headers["Simple-Auth"];
            }
            else {
                return string.Empty;
            }

            if (string.IsNullOrEmpty(key)) {
                return string.Empty;
            }

            var table = client.GetTableReference(Authentication.TableName);
            await table.CreateIfNotExistsAsync();

            var query = TableQuery.CombineFilters(
                TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, Authentication.PartitionName),
                TableOperators.And,
                TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, key)
            );

            var exQuery = new TableQuery<Authentication>().Where(query);

            
            var authenticationKeys = await QueryHelper.ListAll(table, exQuery);
            return authenticationKeys.FirstOrDefault()?.Identity;
        }
    }
}
