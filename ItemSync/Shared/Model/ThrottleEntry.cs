using System;
using System.Collections.Generic;
using System.Text;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Model
{
    class ThrottleEntry : TableEntity {
        public const string TableName = "throttleentry";
        public int Count { get; set; }
    }
}
