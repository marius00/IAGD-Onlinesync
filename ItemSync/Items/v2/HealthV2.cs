using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.Azure.WebJobs.Host;

namespace ItemSync.Items.v2
{
    public static class HealthV2
    {
        [FunctionName("v2_Health")]
        public static IActionResult Run ([HttpTrigger(AuthorizationLevel.Anonymous, "get", "post", Route = null)]HttpRequest req, TraceWriter log) {
            return new OkObjectResult($"Health: OK");
        }
    }
}
