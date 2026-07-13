
package storage

import "github.com/marmyr/iagdbackup/internal/util"

// MaxStringLength bounds every user-supplied string field on an item.
//
const MaxStringLength = 255

// recordFields returns every DBR record path on the item. These are stored,
// deduplicated, in core.db's records table.
func (item JsonItem) recordFields() []string {
	return []string{
		item.BaseRecord,
		item.PrefixRecord,
		item.SuffixRecord,
		item.ModifierRecord,
		item.TransmuteRecord,
		item.MateriaRecord,
		item.RelicCompletionBonusRecord,
		item.EnchantmentRecord,
		item.AscendantAffixNameRecord,
		item.AscendantAffix2hNameRecord,
	}
}

// metadataStrings returns the free-text strings stored directly on the item row.
// These may be localized item names, so they are length-capped but not
// restricted to ASCII.
func (item JsonItem) metadataStrings() []string {
	return []string{
		item.Mod,
		item.Name,
		item.NameLowercase,
		item.Rarity,
	}
}

// HasValidRecords reports whether every record path is printable ASCII and
// within the length cap.
func (item JsonItem) HasValidRecords() bool {
	for _, record := range item.recordFields() {
		if record == "" {
			continue
		}
		if !util.IsASCII(record) || len(record) > MaxStringLength {
			return false
		}
	}

	return true
}

// HasOversizedString reports whether the id or any metadata string exceeds the
// length cap. Record paths are covered separately by HasValidRecords.
func (item JsonItem) HasOversizedString() bool {
	if len(item.Id) > MaxStringLength {
		return true
	}

	for _, s := range item.metadataStrings() {
		if len(s) > MaxStringLength {
			return true
		}
	}

	return false
}
