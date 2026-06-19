package service

import (
	"regexp"

	"github.com/go-playground/validator/v10"

	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
)

// ifMatchRegex entspricht der Original-Regel ^"\d{1,3}".
var ifMatchRegex = regexp.MustCompile(`^"\d{1,3}"`)

// ParkhausRepository definiert das Interface für den Datenbankzugriff,
// das der Service benötigt.
type ParkhausRepository interface {
	FindByID(id uint) (*model.Parkhaus, error)
	ExistsByName(name string) (bool, error)
	Count() (int64, error)
	Search(criteria repository.SearchCriteria, offset, limit int) ([]model.Parkhaus, int64, error)
	Create(p *model.Parkhaus) error
	Update(p *model.Parkhaus) error
	Delete(id uint) error
	CountAutos(parkhausID uint) (int64, error)
	AddAuto(auto *model.Auto) error
	FindFile(parkhausID uint) (*model.ParkhausFile, error)
	ReplaceFile(file *model.ParkhausFile) error
}

// ParkhausService bündelt die Geschäftslogik.
type ParkhausService struct {
	repo      ParkhausRepository
	validator *validator.Validate
}

// NewParkhausService erstellt einen neuen Service.
func NewParkhausService(repo ParkhausRepository) *ParkhausService {
	return &ParkhausService{
		repo:      repo,
		validator: newValidator(),
	}
}
