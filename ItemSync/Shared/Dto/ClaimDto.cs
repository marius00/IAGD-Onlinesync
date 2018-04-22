using System;
using System.Collections.Generic;
using System.Text;

namespace ItemSync.Shared.Dto
{
    public class ClaimDto {
        public string provider_name { get; set; }
        public List<UserClaim> user_claims { get; set; }
    }
}
