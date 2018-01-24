using System;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.Model;
using Microsoft.Azure;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Items {
    public static class Download {
        static readonly long MinimumEpoch = new DateTimeOffset(2018, 1, 1, 0, 0, 0, TimeSpan.Zero).ToUnixTimeMilliseconds();

        [FunctionName("Download")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Function, "get", Route = null)]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }


            long timestamp;
            var timestampString = req.GetQueryNameValuePairs()
                .FirstOrDefault(q => string.Compare(q.Key, "timestamp", true) == 0)
                .Value;

            if (!long.TryParse(timestampString, out timestamp)) {
                return req.CreateResponse(
                    HttpStatusCode.BadRequest,
                    "The query parameter \"timestamp\" is mandatory for item queries"
                );
            }
            else if (timestamp < MinimumEpoch) {
                return req.CreateResponse(
                    HttpStatusCode.BadRequest,
                    "The query parameter \"timestamp\" is not a valid epoch timestamp"
                );
            }

            DateTimeOffset laterThan = DateTimeOffset.FromUnixTimeMilliseconds(timestamp);
            log.Info($"User {partitionKey} has requested an item download for items newer than {laterThan}");
            
            var itemTable = client.GetTableReference(Item.TableName);

            var query = TableQuery.CombineFilters(
                TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, partitionKey),
                TableOperators.And,
                TableQuery.GenerateFilterConditionForDate("Timestamp", QueryComparisons.GreaterThan, laterThan)
            );
            var exQuery = new TableQuery<Item>().Where(query);
            var filteredItems = itemTable.ExecuteQuery(exQuery).Where(m => m.IsActive);

            log.Info($"A total of {filteredItems.Count()} items were returned");
            return req.CreateResponse(HttpStatusCode.OK, filteredItems);
        }
    }
}
