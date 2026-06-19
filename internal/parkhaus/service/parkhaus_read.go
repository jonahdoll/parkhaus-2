package service

import (
	"parkhaus-2/internal/parkhaus/apperr"
	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
)

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
