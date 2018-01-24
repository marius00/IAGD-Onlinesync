using System;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using System.Linq;
using ItemSync.Shared;
using ItemSync.Shared.Dto;
using ItemSync.Shared.Model;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage.Table;
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items
{
    public static class Upload
    {
        [FunctionName("Upload")]
        public static async Task<HttpResponseMessage> Run(
            [HttpTrigger(AuthorizationLevel.Function, "post")]HttpRequestMessage req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            [Table(Item.TableName)] ICollector<Item> collector,
            TraceWriter log
        ) {
            var client = storageAccount.CreateCloudTableClient();
            string partitionKey = Authenticator.Authenticate(client, req);
            if (string.IsNullOrEmpty(partitionKey)) {
                return req.CreateResponse(HttpStatusCode.Unauthorized);
            }


            ItemUploadJson data = await req.Content.ReadAsAsync<ItemUploadJson>();
            if (data == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "Could not correctly parse the request parameters");
            }
            else if (data.Items == null && data.Deleted == null) {
                return req.CreateResponse(HttpStatusCode.BadRequest, "At least one of \"items\" or \"deleted\" must be non-null");
            }
            else if (data.Items?.Count == 0 && data.Deleted?.Count == 0) {
                return req.CreateResponse(HttpStatusCode.OK);
            }


            log.Info($"Received a request from {partitionKey} to upload {data.Items?.Count ?? 0} items and remove {data.Deleted?.Count ?? 0} items");
            if (data.Items?.Count > 0) {
                var newItems = data.Items.Select(m => ItemBuilder.Create(partitionKey, m.Data));
                foreach (var item in newItems) {
                    collector.Add(item);
                }
            }
            if (data.Deleted?.Count > 0) {
                var itemTable = client.GetTableReference(Item.TableName);

                foreach (var itemKey in data.Deleted) {
                    var entity = new DynamicTableEntity(partitionKey, itemKey);
                    entity.ETag = "*";
                    entity.Properties.Add("IsActive", new EntityProperty(false));
                    itemTable.Execute(TableOperation.Merge(entity));
                    
                }
            }
            
            return req.CreateResponse(HttpStatusCode.OK);
        }
    }
}
