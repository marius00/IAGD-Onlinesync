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
                        Delete = HoursToMilliseconds(6),
                        Download = HoursToMilliseconds(6),
                        Upload = HoursToMilliseconds(6)
                    },
                    MultiUsage = new AzureLimitEntry {
                        Delete = MinutesToMilliseconds(15),
                        Download = MinutesToMilliseconds(15),
                        Upload = MinutesToMilliseconds(15)

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