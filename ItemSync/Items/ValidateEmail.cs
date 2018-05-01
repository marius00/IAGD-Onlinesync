using System;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Runtime.Serialization;
using System.Security.Claims;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using System.Web.Http;
using ItemSync.Shared;
using ItemSync.Shared.Model;
using ItemSync.Shared.Service;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items {
    public static class ValidateEmail {
        [FunctionName("ValidateEmail")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post", "options", Route = null)]HttpRequest req,
            [Table(EmailAuthToken.TableName)] ICollector<EmailAuthToken> collector,
            [StorageAccount("StorageConnectionString")] CloudStorageAccount storageAccount,
            TraceWriter log
        ) {
            var email = req.Query["email"].FirstOrDefault()?.ToLower();

            if (string.IsNullOrWhiteSpace(email)) {
                return new BadRequestObjectResult("The query parameter \"email\" is empty or missing");
            }

            log.Info($"E-mail verification request received for {email}");
            var throttleService = new ThrottleService(storageAccount.CreateCloudTableClient(), log);
            var throttled = (await throttleService.ThrottleOrIncrement(email)) || (await throttleService.ThrottleOrIncrement(IpUtility.GetClientIp(req)));
            if (throttled) {
                return new StatusCodeResult(429);
            }

            var code = $"{new Random().Next(100, 999)}{new Random().Next(100, 999)}{new Random().Next(100, 999)}";
            var token = Guid.NewGuid().ToString().Replace("-", "");
            var auth = new EmailAuthToken(email, token, code) {
                Expiration = DateTimeOffset.UtcNow.AddHours(4),
            };
            
            var secret = Environment.GetEnvironmentVariable("email-secret", EnvironmentVariableTarget.Process);
            if (Upload(log, "http://grimdawn.dreamcrash.org/ia/backup/auth.php", $"token={secret}&target={email}&code={code}")) {
                collector.Add(auth);
                log.Info($"Successfully posted email for {email}");
            }

            return new OkObjectResult(new ResponseDto {
                Token = token
            });
        }


        private static bool Upload(TraceWriter log, string url, string postData) {
            HttpWebRequest request = WebRequest.Create(url) as HttpWebRequest;
            var encoding = new UTF8Encoding();
            byte[] data = encoding.GetBytes(postData);

            request.Method = "POST";
            request.Headers.Add(HttpRequestHeader.AcceptEncoding, "gzip, deflate");
            request.Headers.Add(HttpRequestHeader.ContentEncoding, "gzip");
            request.ContentType = "application/x-www-form-urlencoded";

            using (Stream stream = request.GetRequestStream()) {
                stream.Write(data, 0, data.Length);
            }

            using (HttpWebResponse response = (HttpWebResponse)request.GetResponse()) {
                if (response.StatusCode != HttpStatusCode.OK) {
                    log.Warning($"Post failed with response {response.StatusCode}");
                    return false;
                }

                return true;
            }

        }

    }

    class ResponseDto {
        public string Token { get; set; }
    }
}