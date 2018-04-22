using Microsoft.AspNetCore.Http;

namespace ItemSync.Shared
{
    public static class IpUtility {
        public static string GetClientIp(HttpRequest request) {
            return request.HttpContext.Connection.RemoteIpAddress.ToString();
        }
    }
}
