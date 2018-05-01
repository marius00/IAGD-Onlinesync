using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Model
{
    class DeletedItemV2 : TableEntity {
        public const string TableName = "deleteditemv2";

        public string ItemPartitionKey { get; set; }
        public string ItemRowKey { get; set; }
    }
}
