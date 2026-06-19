package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"parkhaus-2/internal/parkhaus/model"
)

// SearchCriteria bündelt die optionalen Suchfilter für die Liste.
type SearchCriteria struct {
	Name           *string
	Kapazitaet     *int
	TarifProStunde *float64
}

// ParkhausRepository kapselt den DB-Zugriff via GORM.
type ParkhausRepository struct {
	db *gorm.DB
}

// NewParkhausRepository erstellt ein neues Repository.
func NewParkhausRepository(db *gorm.DB) *ParkhausRepository {
	return &ParkhausRepository{db: db}
}

// DB liefert das zugrunde liegende GORM-Handle (für Transaktionen im Service).
func (r *ParkhausRepository) DB() *gorm.DB { return r.db }

// FindByID lädt ein Parkhaus inkl. Adresse anhand der ID.
func (r *ParkhausRepository) FindByID(id uint) (*model.Parkhaus, error) {
	var p model.Parkhaus
	err := r.db.Preload("Adresse").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ExistsByName prüft, ob ein Parkhaus mit dem Namen existiert.
func (r *ParkhausRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Parkhaus{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// Count liefert die Gesamtanzahl aller Parkhäuser.
func (r *ParkhausRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Parkhaus{}).Count(&count).Error
	return count, err
}

// Search lädt Parkhäuser inkl. Adresse anhand der Kriterien mit Pagination.
// Gibt die Treffer der Seite und die Gesamtanzahl der gefilterten Treffer zurück.
func (r *ParkhausRepository) Search(criteria SearchCriteria, offset, limit int) ([]model.Parkhaus, int64, error) {
	query := r.db.Model(&model.Parkhaus{})

	if criteria.Name != nil {
		// Teilstring-Suche (case-insensitive): "Berlin" findet auch "Parkhaus Berlin".
		// LIKE-Sonderzeichen im Suchbegriff werden escaped, damit sie woertlich gelten.
		query = query.Where("name ILIKE ?", "%"+escapeLike(*criteria.Name)+"%")
	}
	if criteria.Kapazitaet != nil {
		query = query.Where("kapazitaet <= ?", *criteria.Kapazitaet)
	}
	if criteria.TarifProStunde != nil {
		query = query.Where("tarif_pro_stunde <= ?", *criteria.TarifProStunde)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var parkhaeuser []model.Parkhaus
	err := query.Preload("Adresse").
		Order("id").
		Offset(offset).
		Limit(limit).
		Find(&parkhaeuser).Error
	if err != nil {
		return nil, 0, err
	}
	return parkhaeuser, total, nil
}

// Create legt ein Parkhaus (inkl. Adresse und Autos via Assoziationen) transaktional an.
func (r *ParkhausRepository) Create(p *model.Parkhaus) error {
	return r.db.Create(p).Error
}

// Update speichert geänderte Felder eines Parkhauses.
func (r *ParkhausRepository) Update(p *model.Parkhaus) error {
	return r.db.Save(p).Error
}

// Delete löscht ein Parkhaus anhand der ID (Cascade entfernt Adresse/Autos/File).
func (r *ParkhausRepository) Delete(id uint) error {
	return r.db.Delete(&model.Parkhaus{}, id).Error
}

// CountAutos liefert die Anzahl der Autos eines Parkhauses.
func (r *ParkhausRepository) CountAutos(parkhausID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Auto{}).Where("parkhaus_id = ?", parkhausID).Count(&count).Error
	return count, err
}

// AddAuto fügt ein Auto zu einem Parkhaus hinzu.
func (r *ParkhausRepository) AddAuto(auto *model.Auto) error {
	return r.db.Create(auto).Error
}

// FindFile lädt die Datei zu einem Parkhaus.
func (r *ParkhausRepository) FindFile(parkhausID uint) (*model.ParkhausFile, error) {
	var f model.ParkhausFile
	err := r.db.Where("parkhaus_id = ?", parkhausID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// ReplaceFile ersetzt die Datei eines Parkhauses transaktional (alte löschen, neue anlegen).
func (r *ParkhausRepository) ReplaceFile(file *model.ParkhausFile) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("parkhaus_id = ?", file.ParkhausID).Delete(&model.ParkhausFile{}).Error; err != nil {
			return err
		}
		return tx.Create(file).Error
	})
}

// escapeLike maskiert die LIKE-Sonderzeichen (\, %, _), damit ein Suchbegriff
// woertlich (als Teilstring) und nicht als Muster interpretiert wird.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}
