package service

import (
	"regexp"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"

	"parkhaus-2/internal/parkhaus/apperr"
	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
)

// ifMatchRegex entspricht der Original-Regel ^"\d{1,3}".
var ifMatchRegex = regexp.MustCompile(`^"\d{1,3}"`)

// ParkhausService bündelt die Geschäftslogik.
type ParkhausService struct {
	repo      *repository.ParkhausRepository
	validator *validator.Validate
}

// NewParkhausService erstellt einen neuen Service.
func NewParkhausService(repo *repository.ParkhausRepository) *ParkhausService {
	return &ParkhausService{
		repo:      repo,
		validator: newValidator(),
	}
}

// FindByID liefert ein Parkhaus inkl. Adresse oder einen NotFoundError.
func (s *ParkhausService) FindByID(id uint) (*model.Parkhaus, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, &apperr.NotFoundError{}
	}
	return p, nil
}

// CountAll liefert die Gesamtanzahl aller Parkhäuser (für count-only).
func (s *ParkhausService) CountAll() (int64, error) {
	return s.repo.Count()
}

// Search führt die gefilterte, paginierte Suche aus und liefert ein Page-Objekt.
// Bei keinen Treffern wird ein NotFoundError zurückgegeben.
func (s *ParkhausService) Search(criteria repository.SearchCriteria, page, size int) (*model.Page, error) {
	number := normalizePageNumber(page)
	pageSize := normalizePageSize(size)

	parkhaeuser, total, err := s.repo.Search(criteria, number*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	if len(parkhaeuser) == 0 {
		return nil, &apperr.NotFoundError{}
	}

	result := model.NewPage(model.ToParkhausDTOs(parkhaeuser), pageSize, number, total)
	return &result, nil
}

// Create validiert und legt ein neues Parkhaus transaktional an. Liefert die neue ID.
func (s *ParkhausService) Create(dto model.CreateParkhausDTO) (uint, error) {
	if err := validateStruct(s.validator, dto); err != nil {
		return 0, err
	}

	// Kapazitätsprüfung: Anzahl Autos darf Kapazität nicht überschreiten.
	if len(dto.Autos) > dto.Kapazitaet {
		return 0, &apperr.KapazitaetUeberschrittenError{
			Message: "Die Anzahl der Autos ueberschreitet die Kapazitaet des Parkhauses.",
		}
	}

	// Name-Eindeutigkeit.
	exists, err := s.repo.ExistsByName(dto.Name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, &apperr.ParkhausExistsError{Name: dto.Name}
	}

	now := time.Now()
	parkhaus := &model.Parkhaus{
		Name:           dto.Name,
		Kapazitaet:     dto.Kapazitaet,
		TarifProStunde: dto.TarifProStunde,
		Erzeugt:        now,
		Aktualisiert:   now,
		Adresse: &model.Adresse{
			PLZ:        dto.Adresse.PLZ,
			Ort:        dto.Adresse.Ort,
			Strasse:    dto.Adresse.Strasse,
			Hausnummer: dto.Adresse.Hausnummer,
		},
	}
	for _, a := range dto.Autos {
		parkhaus.Autos = append(parkhaus.Autos, model.Auto{
			Kennzeichen:   a.Kennzeichen,
			Einfahrtszeit: a.Einfahrtszeit,
			Kundentyp:     a.Kundentyp,
		})
	}

	if err := s.repo.Create(parkhaus); err != nil {
		return 0, err
	}
	return parkhaus.ID, nil
}

// Update validiert, prüft die optimistische Sperre und ändert ein Parkhaus.
// Liefert die neue Versionsnummer.
func (s *ParkhausService) Update(id uint, ifMatch string, dto model.UpdateParkhausDTO) (int, error) {
	// If-Match-Header prüfen.
	if ifMatch == "" {
		return 0, &apperr.PreconditionRequiredError{
			Message: "Header If-Match fehlt.",
		}
	}
	if !ifMatchRegex.MatchString(ifMatch) {
		return 0, &apperr.VersionInvalidError{
			Message: "Die Versionsnummer " + ifMatch + " ist ungueltig.",
		}
	}

	if err := validateStruct(s.validator, dto); err != nil {
		return 0, err
	}

	p, err := s.repo.FindByID(id)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, &apperr.NotFoundError{}
	}

	// Versionsabgleich.
	clientVersion, perr := parseIfMatch(ifMatch)
	if perr != nil {
		return 0, &apperr.VersionInvalidError{
			Message: "Die Versionsnummer " + ifMatch + " ist ungueltig.",
		}
	}
	if clientVersion < p.Version {
		return 0, &apperr.VersionOutdatedError{
			Message: "Die Versionsnummer " + ifMatch + " ist nicht aktuell.",
		}
	}

	p.Name = dto.Name
	p.Kapazitaet = dto.Kapazitaet
	p.TarifProStunde = dto.TarifProStunde
	p.Version++
	p.Aktualisiert = time.Now()

	if err := s.repo.Update(p); err != nil {
		return 0, err
	}
	return p.Version, nil
}

// Delete löscht ein Parkhaus idempotent (kein Fehler, wenn nicht vorhanden).
func (s *ParkhausService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// AddAuto fügt ein Auto hinzu und prüft die Kapazität. Liefert die neue Auto-ID.
func (s *ParkhausService) AddAuto(parkhausID uint, dto model.CreateAutoDTO) (uint, error) {
	if err := validateStruct(s.validator, dto); err != nil {
		return 0, err
	}

	p, err := s.repo.FindByID(parkhausID)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, &apperr.NotFoundError{}
	}

	anzahl, err := s.repo.CountAutos(parkhausID)
	if err != nil {
		return 0, err
	}
	if anzahl >= int64(p.Kapazitaet) {
		return 0, &apperr.KapazitaetUeberschrittenError{
			Message: "Das Parkhaus hat keine freie Kapazitaet mehr.",
		}
	}

	auto := &model.Auto{
		Kennzeichen:   dto.Kennzeichen,
		Einfahrtszeit: dto.Einfahrtszeit,
		Kundentyp:     dto.Kundentyp,
		ParkhausID:    parkhausID,
	}
	if err := s.repo.AddAuto(auto); err != nil {
		return 0, err
	}
	return auto.ID, nil
}

// FindFile liefert die Datei eines Parkhauses oder einen NotFoundError.
func (s *ParkhausService) FindFile(parkhausID uint) (*model.ParkhausFile, error) {
	f, err := s.repo.FindFile(parkhausID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, &apperr.NotFoundError{}
	}
	return f, nil
}

// ReplaceFile ersetzt die Datei eines Parkhauses. Parkhaus muss existieren.
func (s *ParkhausService) ReplaceFile(parkhausID uint, filename, mimetype string, data []byte) error {
	p, err := s.repo.FindByID(parkhausID)
	if err != nil {
		return err
	}
	if p == nil {
		return &apperr.NotFoundError{}
	}

	file := &model.ParkhausFile{
		Data:       data,
		Filename:   filename,
		Mimetype:   mimetype,
		ParkhausID: parkhausID,
	}
	return s.repo.ReplaceFile(file)
}

// --- Pagination-Helfer ---

// normalizePageNumber wandelt die 1-basierte API-Seite in eine 0-basierte interne um.
func normalizePageNumber(page int) int {
	if page <= 0 {
		return model.DefaultPageNumber
	}
	return page - 1
}

// normalizePageSize erzwingt die Grenzen 1..MaxPageSize mit Default.
func normalizePageSize(size int) int {
	if size < 1 || size > model.MaxPageSize {
		return model.DefaultPageSize
	}
	return size
}

// parseIfMatch extrahiert die Versionsnummer aus einem If-Match-Header ("<n>").
func parseIfMatch(ifMatch string) (int, error) {
	trimmed := ifMatch
	trimmed = trimQuotes(trimmed)
	return strconv.Atoi(trimmed)
}

// trimQuotes entfernt führende/abschließende Anführungszeichen.
func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
