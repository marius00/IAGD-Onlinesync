using Microsoft.WindowsAzure.Storage.Table;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared.Model {
    public class Item : TableEntity {
        public const string TableName = "item";

        public string Data { get; set; }
        public bool IsActive { get; set; }
    }
}
