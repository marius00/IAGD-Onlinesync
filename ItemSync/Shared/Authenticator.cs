using ItemSync.Shared.Model;
using Microsoft.WindowsAzure.Storage.Table;
using System.Linq;
using System.Net.Http;

namespace ItemSync.Shared {
    public static class Authenticator {
        public static string Authenticate(CloudTableClient client, HttpRequestMessage request) {
            var key = request.Headers.GetValues("Simple-Auth")?.FirstOrDefault();            
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
