using Microsoft.WindowsAzure.Storage.Table;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace ItemSync.Shared.Model {
    public class ItemV1 : TableEntity {
        public const string TableName = "item";
        public bool IsActive { get; set; }

        public string Mod { get; set; }
        public virtual bool IsHardcore { get; set; }

        public string BaseRecord { get; set; }
        public string PrefixRecord { get; set; }
        public string SuffixRecord { get; set; }
        public string ModifierRecord { get; set; }
        public string TransmuteRecord { get; set; }
        public string MateriaRecord { get; set; }
        public string RelicCompletionBonusRecord { get; set; }
        public string EnchantmentRecord { get; set; }

        public long Seed { get; set; }
        public long RelicSeed { get; set; }
        public long EnchantmentSeed { get; set; }
        public long MateriaCombines { get; set; }
        public long StackCount { get; set; }
    }
}
