using System;
using System.Threading.Tasks;
using ItemSync.Shared.AzureCloudTable;
using ItemSync.Shared.Model;
using ItemSync.Shared.Service;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Items
{
    public static class Cleanup
    {
        [FunctionName("Cleanup")]
        public static async Task Run(
            [TimerTrigger("0 0 3 * * *")]TimerInfo myTimer, 
            TraceWriter log,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount
        ) {
            log.Info($"Performing cleanup - {DateTime.Now}");

            var client = storageAccount.CreateCloudTableClient();
            new ThrottleService(client, log).Clear();

            CloudTable table = client.GetTableReference(EmailAuthToken.TableName);
            TableQuery<EmailAuthToken> rangeQuery = new TableQuery<EmailAuthToken>().Where(
                TableQuery.GenerateFilterConditionForDate("Timestamp", 
                    QueryComparisons.LessThan, 
                    DateTime.UtcNow.AddHours(-5)
                )
            );

            try {
                // TODO: Group by PartitionKey and delete in batches.
                await table.CreateIfNotExistsAsync();
                var entities = await QueryHelper.ListAll(table, rangeQuery);

                log.Info($"Got {entities.Count} entities for deletion");
                foreach (var entity in entities) {
                    await table.ExecuteAsync(TableOperation.Delete(entity));
                }

            }
            catch (Exception ex) {
                log.Warning(ex.Message, ex.Source);
                throw;
            }

            log.Info("Cleanup completed");
        }
    }
}
