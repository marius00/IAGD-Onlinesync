SELECT id,
       userid,
       base.record                       AS baserecord,
       IFNULL(prefix.record, '')         as prefixrecord,
       IFNULL(suffix.record, '')         as suffixrecord,
       IFNULL(modifier.record, '')       as modifierrecord,
       IFNULL(relic.record, '')          as reliccompletionbonusrecord,
       IFNULL(transmute.record, '')      as transmuterecord,
       IFNULL(materia.record, '')        as materiarecord,
       IFNULL(enchantment.record, '')    as enchantmentrecord,
       IFNULL(ascAffixName.record, '')   as ascendantAffixRecord,
       IFNULL(ascAffix2hName.record, '') as ascendantAffix2hRecord,
       seed,
       relicseed,
       prefixrarity,
       IFNULL(unknown, 0)                as unknown,
       enchantmentseed,
       materiacombines,
       stackcount,
       IFNULL(`rerollsused`, 0) as rerollsused,
       name,
       namelowercase,
       rarity,
       levelrequirement,
       `mod`,
       ishardcore,
       created_at,
       ts
FROM item i
         LEFT JOIN records as base ON i.id_baserecord = base.id_record
         LEFT JOIN records as prefix ON i.id_prefixrecord = prefix.id_record
         LEFT JOIN records AS suffix ON i.id_suffixrecord = suffix.id_record
         LEFT JOIN records AS modifier ON i.id_modifierrecord = modifier.id_record
         LEFT JOIN records AS transmute ON i.id_transmuterecord = transmute.id_record
         LEFT JOIN records AS materia ON i.id_materiarecord = materia.id_record
         LEFT JOIN records AS relic ON i.id_reliccompletionbonusrecord = relic.id_record
         LEFT JOIN records AS enchantment ON i.id_enchantmentrecord = enchantment.id_record
         LEFT JOIN records AS ascAffixName ON i.id_ascendantaffixname = ascAffixName.id_record
         LEFT JOIN records AS ascAffix2hName ON i.id_ascendantaffix2hname = ascAffix2hName.id_record
WHERE userid = ?
  AND ts > ?
ORDER BY ts ASC
LIMIT ?