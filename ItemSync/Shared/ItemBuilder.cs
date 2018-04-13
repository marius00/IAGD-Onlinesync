using ItemSync.Shared.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using ItemSync.Shared.Dto;

namespace ItemSync.Shared {
    static class ItemBuilder {
        public static Item Create(string email, ItemJson pi) {            
            return new Item {
                PartitionKey = email,
                RowKey = Guid.NewGuid().ToString().Replace("-", ""),

                BaseRecord = pi.BaseRecord,
                EnchantmentRecord = pi.EnchantmentRecord,
                EnchantmentSeed = pi.EnchantmentSeed,
                IsHardcore = pi.IsHardcore,
                MateriaCombines = pi.MateriaCombines,
                MateriaRecord = pi.MateriaRecord,
                Mod = pi.Mod,
                ModifierRecord = pi.ModifierRecord,
                PrefixRecord = pi.PrefixRecord,
                RelicCompletionBonusRecord = pi.RelicCompletionBonusRecord,
                RelicSeed = pi.RelicSeed,
                Seed = pi.Seed,
                StackCount = pi.StackCount,
                SuffixRecord = pi.SuffixRecord,
                TransmuteRecord = pi.TransmuteRecord,
                IsActive = true
            };
        }
    }
}
