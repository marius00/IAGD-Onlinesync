using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Model
{
    class DeletedItem : TableEntity {
        public const string TableName = "deleteditem";

        public string ItemPartitionKey { get; set; }
        public string ItemRowKey { get; set; }
    }
}
