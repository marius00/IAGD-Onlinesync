using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Threading.Tasks;
using ItemSync.Shared.Dto;
using Newtonsoft.Json;

namespace ItemSync.Shared {
    public static class User {
#if DEBUG
        public static string PartitionKey => "dummy@example.com".ToLower();
#endif

        public static async Task<ClaimDto> GetClaims(string token) {
            var httpClient = new HttpClient();
            httpClient.DefaultRequestHeaders.TryAddWithoutValidation("cookie", token);
            var json = await httpClient.GetStringAsync(GetEasyAuthEndpoint());
            return JsonConvert.DeserializeObject<List<ClaimDto>>(json).FirstOrDefault();
        }


        private static string GetEasyAuthEndpoint() {
            var hostname = Environment.GetEnvironmentVariable("WEBSITE_HOSTNAME");
#if DEBUG
            if (hostname.StartsWith("localhost")) {
                return "https://iagd.azurewebsites.net/.auth/me";
            }
#endif
            string requestUri = $"https://{hostname}/.auth/me";
            return requestUri;
        }
    }
}
