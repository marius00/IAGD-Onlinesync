package storage

import (
	"database/sql"
	"github.com/marmyr/iagdbackup/internal/config"
)

// TODO: Move somewhere more appropriate
type JsonItem struct {
	Id string `json:"id"`
	Ts int64  `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore"`

	BaseRecord                 string `json:"baseRecord"`
	PrefixRecord               string `json:"prefixRecord"`
	SuffixRecord               string `json:"suffixRecord" `
	ModifierRecord             string `json:"modifierRecord"`
	TransmuteRecord            string `json:"transmuteRecord"`
	MateriaRecord              string `json:"materiaRecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord"`
	EnchantmentRecord          string `json:"enchantmentRecord"`

	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed"`
	EnchantmentSeed int64 `json:"enchantmentSeed"`
	MateriaCombines int64 `json:"materiaCombines"`
	StackCount      int64 `json:"stackCount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt"`

	// Metadata
	Name             string  `json:"name"`
	NameLowercase    string  `json:"nameLowercase"`
	Rarity           string  `json:"rarity"`
	LevelRequirement float64 `json:"levelRequirement"`
	PrefixRarity     int64   `json:"prefixRarity"`
}

type InputItem struct {
	UserId config.UserId `json:"-" db:"userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" db:"ishardcore"`

	BaseRecord                 sql.NullInt64 `json:"baseRecord" db:"id_baserecord"`
	PrefixRecord               sql.NullInt64 `json:"prefixRecord" db:"id_prefixrecord"`
	SuffixRecord               sql.NullInt64 `json:"suffixRecord" db:"id_suffixrecord"`
	ModifierRecord             sql.NullInt64 `json:"modifierRecord" db:"id_modifierrecord"`
	TransmuteRecord            sql.NullInt64 `json:"transmuteRecord" db:"id_transmuterecord"`
	MateriaRecord              sql.NullInt64 `json:"materiaRecord" db:"id_materiarecord"`
	RelicCompletionBonusRecord sql.NullInt64 `json:"relicCompletionBonusRecord" db:"id_reliccompletionbonusrecord"`
	EnchantmentRecord          sql.NullInt64 `json:"enchantmentRecord" db:"id_enchantmentrecord"`

	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" db:"relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" db:"enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" db:"materiacombines"`
	StackCount      int64 `json:"stackCount" db:"stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" db:"created_at"`

	// Metadata
	Name             string  `json:"name" db:"name"`
	NameLowercase    string  `json:"nameLowercase" db:"namelowercase"`
	Rarity           string  `json:"rarity" db:"rarity"`
	LevelRequirement float64 `json:"levelRequirement" db:"levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" db:"prefixrarity"`
}

func (InputItem) Table() string {
	return "item"
}

// We don't need to return all the stats, only a subset of the fields.
// Fields such as cached stats and searchable text are only used for the webview of backups
type OutputItem struct {
	UserId config.UserId `json:"-" db:"userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" db:"ishardcore"`

	BaseRecord                 string `json:"baseRecord" db:"baserecord"`
	PrefixRecord               string `json:"prefixRecord" db:"prefixrecord"`
	SuffixRecord               string `json:"suffixRecord" db:"suffixrecord"`
	ModifierRecord             string `json:"modifierRecord" db:"modifierrecord"`
	TransmuteRecord            string `json:"transmuteRecord" db:"transmuterecord"`
	MateriaRecord              string `json:"materiaRecord" db:"materiarecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord" db:"reliccompletionbonusrecord"`
	EnchantmentRecord          string `json:"enchantmentRecord" db:"enchantmentrecord"`

	// TODO: Buddy items does not need seed, but is it worth a new struct just to exclude it?
	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" db:"relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" db:"enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" db:"materiacombines"`
	StackCount      int64 `json:"stackCount" db:"stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" db:"created_at"`

	// Metadata
	Name             string  `json:"name" db:"name"`
	NameLowercase    string  `json:"nameLowercase" db:"namelowercase"`
	Rarity           string  `json:"rarity" db:"rarity"`
	LevelRequirement float64 `json:"levelRequirement" db:"levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" db:"prefixrarity"`
}

func (OutputItem) Table() string {
	return "item"
}

// Reference to items which have been deleted. These needs to be stored in DB to ensure that it's deleted from other clients. May have multiple consumers.
type DeletedItem struct {
	UserId config.UserId `json:"-" db:"userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`
}

func (DeletedItem) Table() string {
	return "deleteditem"
}

// Mapping for record foreign keys, used on item insert.
type RecordReference struct {
	Id     uint64 `json:"-" db:"id_record"`
	Record string `json:"record"`
}

func (RecordReference) Table() string {
	return "records"
}
