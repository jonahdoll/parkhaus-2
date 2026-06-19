package service

import (
	"strconv"
	"time"

	"parkhaus-2/internal/parkhaus/apperr"
	"parkhaus-2/internal/parkhaus/model"
)

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

	// Name-Eindeutigkeit prüfen (außer es ist der eigene Name).
	if p.Name != dto.Name {
		exists, err := s.repo.ExistsByName(dto.Name)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, &apperr.ParkhausExistsError{Name: dto.Name}
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

// --- Hilfsfunktionen ---

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
