package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// InitDecimalMarshal sorgt dafür, dass decimal.Decimal in JSON als Zahl (ohne Anführungszeichen)
// serialisiert wird – gemäß Spezifikation muss tarifProStunde eine number sein.
func InitDecimalMarshal() {
	decimal.MarshalJSONWithoutQuotes = true
}

// --- Ausgabe-DTOs (JSON camelCase) ---

// ParkhausDTO ist die JSON-Repräsentation eines Parkhauses.
// tarifProStunde wird als Zahl serialisiert (decimal.Decimal marshalt ohne Anführungszeichen).
type ParkhausDTO struct {
	ID             uint            `json:"id"`
	Version        int             `json:"version"`
	Name           string          `json:"name"`
	Kapazitaet     int             `json:"kapazitaet"`
	TarifProStunde decimal.Decimal `json:"tarifProStunde"`
	Erzeugt        time.Time       `json:"erzeugt"`
	Aktualisiert   time.Time       `json:"aktualisiert"`
	Adresse        *AdresseDTO     `json:"adresse,omitempty"`
	Autos          []AutoDTO       `json:"autos,omitempty"`
}

// AdresseDTO ist die JSON-Repräsentation einer Adresse.
type AdresseDTO struct {
	ID         uint   `json:"id"`
	PLZ        string `json:"plz"`
	Ort        string `json:"ort"`
	Strasse    string `json:"strasse"`
	Hausnummer string `json:"hausnummer"`
	ParkhausID uint   `json:"parkhausId"`
}

// AutoDTO ist die JSON-Repräsentation eines Autos.
type AutoDTO struct {
	ID            uint      `json:"id"`
	Kennzeichen   string    `json:"kennzeichen"`
	Einfahrtszeit time.Time `json:"einfahrtszeit"`
	Kundentyp     Kundentyp `json:"kundentyp"`
	ParkhausID    uint      `json:"parkhausId"`
}

// --- Eingabe-DTOs (mit Validierung via go-playground/validator) ---

// CreateParkhausDTO ist das Eingabe-DTO für POST /rest.
type CreateParkhausDTO struct {
	Name           string           `json:"name" validate:"required,max=100,startswith_word"`
	Kapazitaet     int              `json:"kapazitaet" validate:"required,gt=0"`
	TarifProStunde decimal.Decimal  `json:"tarifProStunde" validate:"dgte0"`
	Adresse        CreateAdresseDTO `json:"adresse" validate:"required"`
	Autos          []CreateAutoDTO  `json:"autos" validate:"omitempty,dive"`
}

// CreateAdresseDTO ist das Eingabe-DTO für eine Adresse beim Anlegen.
type CreateAdresseDTO struct {
	PLZ        string `json:"plz" validate:"required,min=1,max=10"`
	Ort        string `json:"ort" validate:"required,min=1,max=100"`
	Strasse    string `json:"strasse" validate:"required,min=1,max=100"`
	Hausnummer string `json:"hausnummer" validate:"required,min=1,max=10"`
}

// CreateAutoDTO ist das Eingabe-DTO für ein Auto.
type CreateAutoDTO struct {
	Kennzeichen   string    `json:"kennzeichen" validate:"required,min=1,max=20"`
	Einfahrtszeit time.Time `json:"einfahrtszeit" validate:"required"`
	Kundentyp     Kundentyp `json:"kundentyp" validate:"required,oneof=PREMIUM BASIS ANWOHNER"`
}

// UpdateParkhausDTO ist das Eingabe-DTO für PUT /rest/:id (nur drei Felder).
type UpdateParkhausDTO struct {
	Name           string          `json:"name" validate:"required,max=100,startswith_word"`
	Kapazitaet     int             `json:"kapazitaet" validate:"required,gt=0"`
	TarifProStunde decimal.Decimal `json:"tarifProStunde" validate:"dgte0"`
}

// --- Mapping Entity -> DTO ---

// ToParkhausDTO wandelt eine Parkhaus-Entity in ihr DTO um.
func ToParkhausDTO(p Parkhaus) ParkhausDTO {
	dto := ParkhausDTO{
		ID:             p.ID,
		Version:        p.Version,
		Name:           p.Name,
		Kapazitaet:     p.Kapazitaet,
		TarifProStunde: p.TarifProStunde,
		Erzeugt:        p.Erzeugt,
		Aktualisiert:   p.Aktualisiert,
	}
	if p.Adresse != nil {
		a := ToAdresseDTO(*p.Adresse)
		dto.Adresse = &a
	}
	for _, auto := range p.Autos {
		dto.Autos = append(dto.Autos, ToAutoDTO(auto))
	}
	return dto
}

// ToAdresseDTO wandelt eine Adresse-Entity in ihr DTO um.
func ToAdresseDTO(a Adresse) AdresseDTO {
	return AdresseDTO(a)
}

// ToAutoDTO wandelt eine Auto-Entity in ihr DTO um.
func ToAutoDTO(a Auto) AutoDTO {
	return AutoDTO(a)
}

// ToParkhausDTOs wandelt eine Liste von Entities in DTOs um.
func ToParkhausDTOs(parkhaeuser []Parkhaus) []ParkhausDTO {
	dtos := make([]ParkhausDTO, 0, len(parkhaeuser))
	for _, p := range parkhaeuser {
		dtos = append(dtos, ToParkhausDTO(p))
	}
	return dtos
}
