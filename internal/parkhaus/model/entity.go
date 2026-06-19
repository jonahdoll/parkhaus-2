package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Kundentyp entspricht dem Postgres-Enum kundentyp.
type Kundentyp string

const (
	KundentypPremium  Kundentyp = "PREMIUM"
	KundentypBasis    Kundentyp = "BASIS"
	KundentypAnwohner Kundentyp = "ANWOHNER"
)

// Parkhaus ist die GORM-Entity für die Tabelle parkhaus (Schema parkhaus).
type Parkhaus struct {
	ID             uint            `gorm:"column:id;primaryKey"`
	Version        int             `gorm:"column:version;not null;default:0"`
	Name           string          `gorm:"column:name;not null;unique"`
	Kapazitaet     int             `gorm:"column:kapazitaet;not null"`
	TarifProStunde decimal.Decimal `gorm:"column:tarif_pro_stunde;type:numeric(8,2);not null"`
	Erzeugt        time.Time       `gorm:"column:erzeugt;not null;default:now()"`
	Aktualisiert   time.Time       `gorm:"column:aktualisiert;not null;default:now()"`

	Adresse *Adresse `gorm:"foreignKey:ParkhausID;constraint:OnDelete:CASCADE"`
	Autos   []Auto   `gorm:"foreignKey:ParkhausID;constraint:OnDelete:CASCADE"`
}

// TableName erzwingt den Tabellennamen ohne Pluralisierung.
func (Parkhaus) TableName() string { return "parkhaus" }

// Adresse ist die GORM-Entity für die Tabelle adresse (1:1 zu Parkhaus).
type Adresse struct {
	ID         uint   `gorm:"column:id;primaryKey"`
	PLZ        string `gorm:"column:plz;not null"`
	Ort        string `gorm:"column:ort;not null"`
	Strasse    string `gorm:"column:strasse;not null"`
	Hausnummer string `gorm:"column:hausnummer;not null"`
	ParkhausID uint   `gorm:"column:parkhaus_id;not null;unique"`
}

// TableName erzwingt den Tabellennamen.
func (Adresse) TableName() string { return "adresse" }

// Auto ist die GORM-Entity für die Tabelle auto (n:1 zu Parkhaus).
type Auto struct {
	ID            uint      `gorm:"column:id;primaryKey"`
	Kennzeichen   string    `gorm:"column:kennzeichen;not null"`
	Einfahrtszeit time.Time `gorm:"column:einfahrtszeit;not null"`
	Kundentyp     Kundentyp `gorm:"column:kundentyp;type:kundentyp;not null"`
	ParkhausID    uint      `gorm:"column:parkhaus_id;not null;index"`
}

// TableName erzwingt den Tabellennamen.
func (Auto) TableName() string { return "auto" }

// ParkhausFile ist die GORM-Entity für die Tabelle parkhaus_file (1:1 zu Parkhaus).
type ParkhausFile struct {
	ID         uint   `gorm:"column:id;primaryKey"`
	Data       []byte `gorm:"column:data;type:bytea;not null"`
	Filename   string `gorm:"column:filename;not null"`
	Mimetype   string `gorm:"column:mimetype"`
	ParkhausID uint   `gorm:"column:parkhaus_id;not null;unique"`
}

// TableName erzwingt den Tabellennamen.
func (ParkhausFile) TableName() string { return "parkhaus_file" }
