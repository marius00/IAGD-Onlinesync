using ItemSync.Shared.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared {
    static class ItemBuilder {

        public static Item Create(string email, string key, string data) {
            return new Item {
                PartitionKey = email,
                RowKey = $"item-{key}",
                Data = data,
                IsActive = true
            };
        }


        public static Item Create(string email, string data) {            
            return new Item {
                PartitionKey = email,
                RowKey = $"item-{Guid.NewGuid().ToString()}",
                Data = data,
                IsActive = true
            };
        }
    }
}
