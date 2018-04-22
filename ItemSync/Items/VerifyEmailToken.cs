using System;
using System.Threading.Tasks;
using ItemSync.Shared;
using ItemSync.Shared.AzureCloudTable;
using ItemSync.Shared.Model;
using ItemSync.Shared.Service;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Items {
    public static class VerifyEmailToken {
        [FunctionName("VerifyEmailToken")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post", "options", Route = null)]HttpRequest req,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var token = req.Query["token"];
            var code = req.Query["code"];
            var client = storageAccount.CreateCloudTableClient();

            if (string.IsNullOrWhiteSpace(token)) {
                return new BadRequestObjectResult("The query parameter \"token\" is empty or missing");
            }
            else if (string.IsNullOrWhiteSpace(code)) {
                return new BadRequestObjectResult("The query parameter \"code\" is empty or missing");
            }

            log.Info($"E-mail verification request received for {token}");

            // Throttling required for verification as well, to prevent infinite attempts on a token.
            var throttleService = new ThrottleService(storageAccount.CreateCloudTableClient(), log);
            var throttled = (await throttleService.ThrottleOrIncrement("attempt-" + token)) || (await throttleService.ThrottleOrIncrement("attempt-" + IpUtility.GetClientIp(req)));
            if (throttled) {
                return new StatusCodeResult(429);
            }

            var table = client.GetTableReference(EmailAuthToken.TableName);
            await table.CreateIfNotExistsAsync();

            var query = TableQuery.CombineFilters(
                TableQuery.GenerateFilterCondition("PartitionKey", QueryComparisons.Equal, token),
                TableOperators.And,
                TableQuery.GenerateFilterCondition("RowKey", QueryComparisons.Equal, code)
            );
            var exQuery = new TableQuery<EmailAuthToken>().Where(query);

            var entry = await QueryHelper.Get(table, exQuery);
            if (entry != null) {
                if (DateTimeOffset.UtcNow < entry.Expiration) {
                    log.Info($"Authentication successful for {entry.Email}");

                    var auth = new Authentication {
                        PartitionKey = Authentication.PartitionName,
                        RowKey = (Guid.NewGuid().ToString() + Guid.NewGuid().ToString()).Replace("-", ""),
                        Identity = entry.Email
                    };

                    var authTable = client.GetTableReference(Authentication.TableName);
                    await authTable.CreateIfNotExistsAsync();
                    await authTable.ExecuteAsync(TableOperation.Insert(auth));

                    return new OkObjectResult(new VerifyEmailTokenResponseDto {
                        Success = true,
                        Token = auth.RowKey
                    });
                }
                else {
                    log.Info($"Authentication token {token} for {entry.Email} has already expired, expiration {entry.Expiration}");

                    return new OkObjectResult(new VerifyEmailTokenResponseDto {
                        Success = false
                    });
                }
            }
            else {
                log.Info("No entry found");
                return new OkObjectResult(new VerifyEmailTokenResponseDto {
                    Success = false
                });
            }
        }

    }

    class VerifyEmailTokenResponseDto {
        public bool Success { get; set; }
        public string Token { get; set; }
    }
}