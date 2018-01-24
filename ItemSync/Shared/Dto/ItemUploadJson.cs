using ItemSync.Shared.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared.Dto {
    public class ItemUploadJson {
        public List<ItemJson> Items;
        public List<string> Deleted;
    }
}
