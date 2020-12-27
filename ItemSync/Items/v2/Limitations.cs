using System;
using System.Threading.Tasks;
using System.Web.Http;
using ItemSync.Shared;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;
using Microsoft.WindowsAzure.Storage;

namespace ItemSync.Items.v2 {
    public static class Limitations {
        [FunctionName("v2_Limits")]
        public static async Task<IActionResult> Run([HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = null)]HttpRequest req) {
            return new OkObjectResult(
                new AzureLimitsDto {
                    Regular = new AzureLimitEntry {
                        Delete = HoursToMilliseconds(3),
                        Download = HoursToMilliseconds(3),
                        Upload = HoursToMilliseconds(3)
                    },
                    MultiUsage = new AzureLimitEntry {
                        Delete = MinutesToMilliseconds(1),
                        Download = MinutesToMilliseconds(2),
                        Upload = MinutesToMilliseconds(2)
                    }
                }
            );
        }

        static long HoursToMilliseconds(int hours) {
            return hours * 60 * 60 * 1000;
        }
        static long MinutesToMilliseconds(int minutes) {
            return minutes * 60 * 1000;
        }

        public class AzureLimitsDto {
            public AzureLimitEntry Regular { get; set; }
            public AzureLimitEntry MultiUsage { get; set; }
        }
        public class AzureLimitEntry {
            public long Download { get; set; }
            public long Upload { get; set; }
            public long Delete { get; set; }
        }
    }

}