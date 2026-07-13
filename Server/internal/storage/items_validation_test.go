package storage

import (
	"strings"
	"testing"
)

func TestHasValidRecords(t *testing.T) {
	cap := strings.Repeat("a", MaxStringLength)
	over := strings.Repeat("a", MaxStringLength+1)

	cases := []struct {
		name string
		item JsonItem
		want bool
	}{
		{"empty records", JsonItem{}, true},
		{"valid ascii record", JsonItem{BaseRecord: "records/items/x.dbr"}, true},
		{"record at cap", JsonItem{BaseRecord: cap}, true},
		{"record over cap", JsonItem{BaseRecord: over}, false},
		{"non-ascii record", JsonItem{PrefixRecord: "récords/x.dbr"}, false},
		{"oversized ascendant record", JsonItem{AscendantAffixNameRecord: over}, false},
		{"oversized relic bonus record", JsonItem{RelicCompletionBonusRecord: over}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.item.HasValidRecords(); got != tc.want {
				t.Fatalf("HasValidRecords() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasOversizedString(t *testing.T) {
	over := strings.Repeat("x", MaxStringLength+1)
	cap := strings.Repeat("x", MaxStringLength)

	cases := []struct {
		name string
		item JsonItem
		want bool
	}{
		{"empty", JsonItem{}, false},
		{"metadata at cap", JsonItem{Name: cap, Rarity: cap, Mod: cap, NameLowercase: cap}, false},
		{"oversized name", JsonItem{Name: over}, true},
		{"oversized rarity", JsonItem{Rarity: over}, true},
		{"oversized mod", JsonItem{Mod: over}, true},
		{"oversized namelowercase", JsonItem{NameLowercase: over}, true},
		{"oversized id", JsonItem{Id: over}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.item.HasOversizedString(); got != tc.want {
				t.Fatalf("HasOversizedString() = %v, want %v", got, tc.want)
			}
		})
	}
}
