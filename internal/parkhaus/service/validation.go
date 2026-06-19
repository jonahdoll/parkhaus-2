package service

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"parkhaus-2/internal/parkhaus/apperr"
)

// wordStartRegex entspricht der Zod-Regel ^\w.* (beginnt mit Wortzeichen).
var wordStartRegex = regexp.MustCompile(`^\w.*`)

// newValidator erzeugt einen validator mit den benötigten Custom-Regeln.
func newValidator() *validator.Validate {
	v := validator.New()

	// JSON-Feldnamen statt Go-Feldnamen für die Fehler-Pfade verwenden.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// startswith_word: String muss mit einem Wortzeichen beginnen (^\w.*).
	_ = v.RegisterValidation("startswith_word", func(fl validator.FieldLevel) bool {
		return wordStartRegex.MatchString(fl.Field().String())
	})

	// dgte0: decimal.Decimal >= 0.
	_ = v.RegisterValidation("dgte0", func(fl validator.FieldLevel) bool {
		if d, ok := fl.Field().Interface().(decimal.Decimal); ok {
			return d.GreaterThanOrEqual(decimal.Zero)
		}
		return false
	})

	return v
}

// validateStruct validiert ein Struct und liefert bei Fehlern einen *apperr.ValidationError.
func validateStruct(v *validator.Validate, s any) error {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !asValidationErrors(err, &ve) {
		return err
	}

	issues := make([]apperr.ValidationIssue, 0, len(ve))
	for _, fe := range ve {
		issues = append(issues, apperr.ValidationIssue{
			Path:    fieldPath(fe.Namespace()),
			Message: messageFor(fe),
		})
	}
	return &apperr.ValidationError{Issues: issues}
}

// asValidationErrors prüft, ob err vom Typ validator.ValidationErrors ist.
func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

// fieldPath wandelt einen validator-Namespace (z.B. "CreateParkhausDTO.adresse.plz")
// in einen Pfad ohne Root-Struct um (["adresse","plz"]).
func fieldPath(namespace string) []string {
	parts := strings.Split(namespace, ".")
	if len(parts) > 1 {
		parts = parts[1:] // Root-Struct-Namen entfernen
	}
	return parts
}

// messageFor erzeugt eine lesbare Fehlermeldung für ein Feld.
func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Feld ist erforderlich"
	case "max":
		return "Wert ist zu lang (max. " + fe.Param() + ")"
	case "min":
		return "Wert ist zu kurz (min. " + fe.Param() + ")"
	case "gt":
		return "Wert muss groesser als " + fe.Param() + " sein"
	case "dgte0":
		return "Wert muss groesser oder gleich 0 sein"
	case "oneof":
		return "Wert muss einer von [" + fe.Param() + "] sein"
	case "startswith_word":
		return "Wert muss mit einem Buchstaben oder einer Ziffer beginnen"
	default:
		return "ungueltiger Wert"
	}
}
