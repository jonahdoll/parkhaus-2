package apperr

import "fmt"

// ValidationIssue beschreibt einen einzelnen Validierungsfehler mit Feld-Pfad.
type ValidationIssue struct {
	Path    []string `json:"path"`
	Message string   `json:"message"`
}

// ValidationError bündelt mehrere Validierungsfehler (-> HTTP 422).
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validierung fehlgeschlagen: %d fehler", len(e.Issues))
}

// NotFoundError signalisiert eine nicht gefundene Ressource (-> HTTP 404).
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	if e.Message == "" {
		return "nicht gefunden"
	}
	return e.Message
}

// ParkhausExistsError signalisiert einen bereits vergebenen Namen (-> HTTP 422).
type ParkhausExistsError struct {
	Name string
}

func (e *ParkhausExistsError) Error() string {
	return fmt.Sprintf("Ein Parkhaus mit dem Namen %s existiert bereits.", e.Name)
}

// KapazitaetUeberschrittenError signalisiert überschrittene Kapazität (-> HTTP 422).
type KapazitaetUeberschrittenError struct {
	Message string
}

func (e *KapazitaetUeberschrittenError) Error() string {
	if e.Message == "" {
		return "Kapazitaet ueberschritten"
	}
	return e.Message
}

// VersionInvalidError signalisiert eine ungültige Versionsangabe (-> HTTP 412).
type VersionInvalidError struct {
	Message string
}

func (e *VersionInvalidError) Error() string {
	return e.Message
}

// VersionOutdatedError signalisiert eine veraltete Versionsangabe (-> HTTP 412).
type VersionOutdatedError struct {
	Message string
}

func (e *VersionOutdatedError) Error() string {
	return e.Message
}

// PreconditionRequiredError signalisiert einen fehlenden If-Match-Header (-> HTTP 428).
type PreconditionRequiredError struct {
	Message string
}

func (e *PreconditionRequiredError) Error() string {
	return e.Message
}
