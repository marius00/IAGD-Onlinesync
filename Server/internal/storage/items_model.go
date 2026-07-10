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
	AscendantAffixNameRecord   string `json:"ascendantAffixNameRecord"`
	AscendantAffix2hNameRecord string `json:"ascendantAffix2hNameRecord"`

	Seed             int64 `json:"seed"`
	RelicSeed        int64 `json:"relicSeed"`
	EnchantmentSeed  int64 `json:"enchantmentSeed"`
	MateriaCombines  int64 `json:"materiaCombines"`
	StackCount       int64 `json:"stackCount"`
	RerollsUsed      int64 `json:"rerollsUsed" db:"rerollsused"`
	AffixRerollsUsed int64 `json:"affixRerollsUsed" db:"affixrerollsused"`

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
	UserId config.UserId `json:"-" db:"userid" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" db:"ishardcore" gorm:"column:ishardcore"`

	BaseRecord                 sql.NullInt64 `json:"baseRecord" db:"id_baserecord" gorm:"column:id_baserecord"`
	PrefixRecord               sql.NullInt64 `json:"prefixRecord" db:"id_prefixrecord" gorm:"column:id_prefixrecord"`
	SuffixRecord               sql.NullInt64 `json:"suffixRecord" db:"id_suffixrecord" gorm:"column:id_suffixrecord"`
	ModifierRecord             sql.NullInt64 `json:"modifierRecord" db:"id_modifierrecord" gorm:"column:id_modifierrecord"`
	TransmuteRecord            sql.NullInt64 `json:"transmuteRecord" db:"id_transmuterecord" gorm:"column:id_transmuterecord"`
	MateriaRecord              sql.NullInt64 `json:"materiaRecord" db:"id_materiarecord" gorm:"column:id_materiarecord"`
	RelicCompletionBonusRecord sql.NullInt64 `json:"relicCompletionBonusRecord" db:"id_reliccompletionbonusrecord" gorm:"column:id_reliccompletionbonusrecord"`
	EnchantmentRecord          sql.NullInt64 `json:"enchantmentRecord" db:"id_enchantmentrecord" gorm:"column:id_enchantmentrecord"`
	AscendantAffixName         sql.NullInt64 `json:"ascendantAffixNameRecord" db:"id_ascendantaffixname" gorm:"column:id_ascendantaffixname"`
	AscendantAffix2hName       sql.NullInt64 `json:"ascendantAffix2hNameRecord" db:"id_ascendantaffix2hname" gorm:"column:id_ascendantaffix2hname"`

	Seed             int64 `json:"seed"`
	RelicSeed        int64 `json:"relicSeed" db:"relicseed" gorm:"column:relicseed"`
	EnchantmentSeed  int64 `json:"enchantmentSeed" db:"enchantmentseed" gorm:"column:enchantmentseed"`
	MateriaCombines  int64 `json:"materiaCombines" db:"materiacombines" gorm:"column:materiacombines"`
	StackCount       int64 `json:"stackCount" db:"stackcount" gorm:"column:stackcount"`
	RerollsUsed      int64 `json:"rerollsUsed" db:"rerollsused" gorm:"column:rerollsused"`
	AffixRerollsUsed int64 `json:"affixRerollsUsed" db:"affixrerollsused" gorm:"column:affixrerollsused"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" db:"created_at"`

	// Metadata
	Name             string  `json:"name" db:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" db:"namelowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" db:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" db:"levelrequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" db:"prefixrarity" gorm:"column:prefixrarity"`
}

func (InputItem) Table() string {
	return "item"
}
func (InputItem) TableName() string {
	return "item"
}

// TODO!! Error performing maintenance, ctx expired
// [GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
//Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.

// We don't need to return all the stats, only a subset of the fields.
// Fields such as cached stats and searchable text are only used for the webview of backups
type OutputItem struct {
	UserId config.UserId `json:"-" db:"userid" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" db:"ishardcore" gorm:"column:ishardcore"`

	BaseRecord                 string `json:"baseRecord" db:"baserecord" gorm:"column:baserecord"`
	PrefixRecord               string `json:"prefixRecord" db:"prefixrecord" gorm:"column:prefixrecord"`
	SuffixRecord               string `json:"suffixRecord" db:"suffixrecord" gorm:"column:suffixrecord"`
	ModifierRecord             string `json:"modifierRecord" db:"modifierrecord" gorm:"column:modifierrecord"`
	TransmuteRecord            string `json:"transmuteRecord" db:"transmuterecord" gorm:"column:transmuterecord"`
	MateriaRecord              string `json:"materiaRecord" db:"materiarecord" gorm:"column:materiarecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord" db:"reliccompletionbonusrecord" gorm:"column:reliccompletionbonusrecord"`
	EnchantmentRecord          string `json:"enchantmentRecord" db:"enchantmentrecord" gorm:"column:enchantmentrecord"`
	AscendantAffixNameRecord   string `json:"ascendantAffixNameRecord" db:"ascendantAffixRecord" gorm:"column:ascendantAffixRecord"`
	AscendantAffix2hNameRecord string `json:"ascendantAffix2hNameRecord" db:"ascendantAffix2hRecord" gorm:"column:ascendantAffix2hRecord"`

	// TODO: Buddy items does not need seed, but is it worth a new struct just to exclude it?
	Seed             int64 `json:"seed"`
	RelicSeed        int64 `json:"relicSeed" db:"relicseed" gorm:"column:relicseed"`
	EnchantmentSeed  int64 `json:"enchantmentSeed" db:"enchantmentseed" gorm:"column:enchantmentseed"`
	MateriaCombines  int64 `json:"materiaCombines" db:"materiacombines" gorm:"column:materiacombines"`
	StackCount       int64 `json:"stackCount" db:"stackcount" gorm:"column:stackcount"`
	RerollsUsed      int64 `json:"rerollsUsed" db:"rerollsused" gorm:"column:rerollsused"`
	AffixRerollsUsed int64 `json:"affixRerollsUsed" db:"affixrerollsused" gorm:"column:affixrerollsused"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" db:"created_at" gorm:"column:created_at"`

	// Metadata
	Name             string  `json:"name" db:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" db:"namelowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" db:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" db:"levelrequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" db:"prefixrarity" gorm:"column:prefixrarity"`
	Unknown          int64   `json:"unknown" db:"unknown" gorm:"column:unknown"`
}

func (OutputItem) Table() string {
	return "item"
}
func (OutputItem) TableName() string {
	return "item"
}

// Reference to items which have been deleted. These needs to be stored in DB to ensure that it's deleted from other clients. May have multiple consumers.
type DeletedItem struct {
	UserId config.UserId `json:"-" db:"userid" gorm:"column:userid"`
	Id     string        `json:"id"`
	Ts     int64         `json:"ts"`
}

func (DeletedItem) Table() string {
	return "deleteditem"
}
func (DeletedItem) TableName() string {
	return "deleteditem"
}

// Mapping for record foreign keys, used on item insert.
type RecordReference struct {
	Id     uint64 `json:"-" db:"id_record" gorm:"column:id_record"`
	Record string `json:"record"`
}

func (RecordReference) Table() string {
	return "records"
}
func (RecordReference) TableName() string {
	return "records"
}
