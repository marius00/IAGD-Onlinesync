using System;
using ItemSync.Shared.Model;
using ItemSync.Shared.Service;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items
{
    public static class Cleanup
    {
        [FunctionName("Cleanup")]
        public static async void Run(
            [TimerTrigger("0 0 3 * * *")]TimerInfo myTimer, 
            TraceWriter log,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount
        ) {
            log.Info($"Performing cleanup - {DateTime.Now}");

            var client = storageAccount.CreateCloudTableClient();
            new ThrottleService(client, log).Clear();
            
            // Not the perfect solution, but unlikely to ever affect anyone.
            await client.GetTableReference(EmailAuthToken.TableName).DeleteIfExistsAsync();
        }
    }
}
