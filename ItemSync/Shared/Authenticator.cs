using System;
using System.Collections.Generic;
using ItemSync.Shared.Model;
using Microsoft.WindowsAzure.Storage.Table;
using System.Linq;
using System.Net.Http;
using System.Web;

namespace ItemSync.Shared {
    public static class Authenticator {
        public static string Authenticate(CloudTableClient client, HttpRequestMessage request) {
            string key;

            if (request.Headers.TryGetValues("Simple-Auth", out var s)) {
                key = s.FirstOrDefault();
            }
            else {
                return string.Empty;
            }

            if (string.IsNullOrEmpty(key)) {
                return string.Empty;
            }

            var table = client.GetTableReference(Authentication.TableName);

            var query = TableQuery.CombineFilters(
                TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, Authentication.PartitionName),
                TableOperators.And,
                TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, key)
            );

            var exQuery = new TableQuery<Authentication>().Where(query);

            var authenticationKeys = table.ExecuteQuery(exQuery);
            return authenticationKeys.FirstOrDefault()?.Identity;
        }
    }
}
