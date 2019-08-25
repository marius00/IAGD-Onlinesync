using ItemSync.Shared.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using ItemSync.Shared.Dto;

namespace ItemSync.Shared {
    static class ItemBuilder {
        public static ItemV1 CreateV1(string email, ItemJson pi) {
            return new ItemV1 {
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

        public static ItemV2 CreateV2(string email, ItemJson pi) {
            return CreateV3(email, Guid.NewGuid().ToString().Replace("-", ""), pi);
        }

        public static ItemV2 CreateV3(string email, string id, ItemJson item) {
            return new ItemV2 {
                PartitionKey = email,
                RowKey = id,
                BaseRecord = item.BaseRecord,
                EnchantmentRecord = item.EnchantmentRecord,
                EnchantmentSeed = item.EnchantmentSeed,
                IsHardcore = item.IsHardcore,
                MateriaCombines = item.MateriaCombines,
                MateriaRecord = item.MateriaRecord,
                Mod = item.Mod,
                ModifierRecord = item.ModifierRecord,
                PrefixRecord = item.PrefixRecord,
                RelicCompletionBonusRecord = item.RelicCompletionBonusRecord,
                RelicSeed = item.RelicSeed,
                Seed = item.Seed,
                StackCount = item.StackCount,
                SuffixRecord = item.SuffixRecord,
                TransmuteRecord = item.TransmuteRecord,
                IsActive = true
            };
        }
    }
}
