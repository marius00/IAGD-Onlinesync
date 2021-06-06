package storage

import (
	"database/sql"
	"github.com/marmyr/iagdbackup/internal/config"
)

// TODO: Move somewhere more appropriate
// TODO: Remove GORM references (not yet, still used to read from postgres)
type JsonItem struct {
	UserId config.UserId `json:"-" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" gorm:"column:ishardcore"`

	BaseRecord                 string `json:"baseRecord" gorm:"column:baserecord"`
	PrefixRecord               string `json:"prefixRecord" gorm:"column:prefixrecord"`
	SuffixRecord               string `json:"suffixRecord" gorm:"column:suffixrecord"`
	ModifierRecord             string `json:"modifierRecord" gorm:"column:modifierrecord"`
	TransmuteRecord            string `json:"transmuteRecord" gorm:"column:transmuterecord"`
	MateriaRecord              string `json:"materiaRecord" gorm:"column:materiarecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord" gorm:"column:reliccompletionbonusrecord"`
	EnchantmentRecord          string `json:"enchantmentRecord" gorm:"column:enchantmentrecord"`

	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" gorm:"column:relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" gorm:"column:enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" gorm:"column:materiacombines"`
	StackCount      int64 `json:"stackCount" gorm:"column:stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`

	// Metadata
	Name             string  `json:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" gorm:"column:prefixrarity"`
}

type InputItem struct {
	UserId config.UserId `json:"-" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" gorm:"column:ishardcore"`

	BaseRecord                 sql.NullInt64 `json:"baseRecord" gorm:"column:id_baserecord"`
	PrefixRecord               sql.NullInt64 `json:"prefixRecord" gorm:"column:id_prefixrecord"`
	SuffixRecord               sql.NullInt64 `json:"suffixRecord" gorm:"column:id_suffixrecord"`
	ModifierRecord             sql.NullInt64 `json:"modifierRecord" gorm:"column:id_modifierrecord"`
	TransmuteRecord            sql.NullInt64 `json:"transmuteRecord" gorm:"column:id_transmuterecord"`
	MateriaRecord              sql.NullInt64 `json:"materiaRecord" gorm:"column:id_materiarecord"`
	RelicCompletionBonusRecord sql.NullInt64 `json:"relicCompletionBonusRecord" gorm:"column:id_reliccompletionbonusrecord"`
	EnchantmentRecord          sql.NullInt64 `json:"enchantmentRecord" gorm:"column:id_enchantmentrecord"`

	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" gorm:"column:relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" gorm:"column:enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" gorm:"column:materiacombines"`
	StackCount      int64 `json:"stackCount" gorm:"column:stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`

	// Metadata
	Name             string  `json:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" gorm:"column:prefixrarity"`
}

func (InputItem) TableName() string {
	return "item"
}

// We don't need to return all the stats, only a subset of the fields.
// Fields such as cached stats and searchable text are only used for the webview of backups
type OutputItem struct {
	UserId config.UserId `json:"-" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" gorm:"column:ishardcore"`

	BaseRecord                 string `json:"baseRecord" gorm:"column:baserecord"`
	PrefixRecord               string `json:"prefixRecord" gorm:"column:prefixrecord"`
	SuffixRecord               string `json:"suffixRecord" gorm:"column:suffixrecord"`
	ModifierRecord             string `json:"modifierRecord" gorm:"column:modifierrecord"`
	TransmuteRecord            string `json:"transmuteRecord" gorm:"column:transmuterecord"`
	MateriaRecord              string `json:"materiaRecord" gorm:"column:materiarecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord" gorm:"column:reliccompletionbonusrecord"`
	EnchantmentRecord          string `json:"enchantmentRecord" gorm:"column:enchantmentrecord"`

	// TODO: Buddy items does not need seed, but is it worth a new struct just to exclude it?
	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" gorm:"column:relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" gorm:"column:enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" gorm:"column:materiacombines"`
	StackCount      int64 `json:"stackCount" gorm:"column:stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`

	// Metadata
	Name             string  `json:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" gorm:"column:prefixrarity"`
}

func (OutputItem) TableName() string {
	return "item"
}

// Reference to items which have been deleted. These needs to be stored in DB to ensure that it's deleted from other clients. May have multiple consumers.
type DeletedItem struct {
	UserId config.UserId `json:"-" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`
}

func (DeletedItem) TableName() string {
	return "deleteditem"
}

// Mapping for record foreign keys, used on item insert.
type RecordReference struct {
	Id     uint64 `json:"-" gorm:"column:id_record"`
	Record string `json:"record"`
}

func (RecordReference) TableName() string {
	return "records"
}
