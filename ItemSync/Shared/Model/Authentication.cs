using Microsoft.WindowsAzure.Storage.Table;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared.Model {
    public class Authentication : TableEntity {
        public const string TableName = "auth";
        public const string PartitionName = "auth";


        public string Identity { get; set; }
    }
}
