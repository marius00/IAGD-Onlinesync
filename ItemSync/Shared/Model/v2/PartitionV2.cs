using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Model
{
    public class PartitionV2 : TableEntity {
        public const string TableName = "partitionlistingv2";
        public bool IsActive { get; set; }
    }
}
