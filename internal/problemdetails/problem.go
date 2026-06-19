package problemdetails

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"parkhaus-2/internal/parkhaus/apperr"
)

// ProblemDetail ist das RFC-9457-Antwortobjekt.
type ProblemDetail struct {
	Title      string `json:"title"`
	StatusCode int    `json:"statusCode"`
	Detail     any    `json:"detail"`
}

// titleForStatus liefert den Titel gemäß Statuscode-Mapping der Spezifikation.
func titleForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusPreconditionFailed:
		return "Precondition Failed"
	case http.StatusUnprocessableEntity:
		return "Unprocessable Content"
	case http.StatusPreconditionRequired:
		return "Precondition Required"
	default:
		return "Client Error"
	}
}

// Write schreibt eine Problem-Details-Antwort mit application/problem+json.
func Write(c *gin.Context, status int, detail any) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, ProblemDetail{
		Title:      titleForStatus(status),
		StatusCode: status,
		Detail:     detail,
	})
}

// HandleError mappt einen Domänenfehler auf die passende HTTP-Antwort.
func HandleError(c *gin.Context, err error) {
	var validationErr *apperr.ValidationError
	var notFoundErr *apperr.NotFoundError
	var existsErr *apperr.ParkhausExistsError
	var kapazitaetErr *apperr.KapazitaetUeberschrittenError
	var versionInvalidErr *apperr.VersionInvalidError
	var versionOutdatedErr *apperr.VersionOutdatedError
	var preconditionErr *apperr.PreconditionRequiredError

	switch {
	case errors.As(err, &validationErr):
		Write(c, http.StatusUnprocessableEntity, validationErr.Issues)
	case errors.As(err, &existsErr):
		Write(c, http.StatusUnprocessableEntity, existsErr.Error())
	case errors.As(err, &kapazitaetErr):
		Write(c, http.StatusUnprocessableEntity, kapazitaetErr.Error())
	case errors.As(err, &versionInvalidErr):
		Write(c, http.StatusPreconditionFailed, versionInvalidErr.Error())
	case errors.As(err, &versionOutdatedErr):
		Write(c, http.StatusPreconditionFailed, versionOutdatedErr.Error())
	case errors.As(err, &preconditionErr):
		Write(c, http.StatusPreconditionRequired, preconditionErr.Error())
	case errors.As(err, &notFoundErr):
		// NotFound: Standard-404 ohne Problem Details.
		c.Status(http.StatusNotFound)
	default:
		// Sonstiger interner Fehler: Plain-Text.
		c.String(http.StatusInternalServerError, "Interner Fehler")
	}
}
