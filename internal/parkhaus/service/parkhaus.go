package service

import (
	"regexp"

	"github.com/go-playground/validator/v10"

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
