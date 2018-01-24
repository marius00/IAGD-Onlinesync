using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared {
    public static class User {
#if DEBUG
        public static string PartitionKey => "dummy@example.com";
#else
        public static string PartitionKey => ClaimsPrincipal.Current.Identity.Name;
#endif
    }
}
