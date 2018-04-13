using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Claims; // NOT UNUSED IN RELEASE MODE
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared {
    public static class User {
#if DEBUG
        //public static string PartitionKey => ClaimsPrincipal.Current.Identity.Name;
        public static string PartitionKey => "dummy@example.com".ToLower();
#else
        public static string PartitionKey => ClaimsPrincipal.Current.Identity.Name?.ToLower();
#endif
    }
}
