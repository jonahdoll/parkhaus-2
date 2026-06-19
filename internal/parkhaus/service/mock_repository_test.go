package service

import (
	"github.com/stretchr/testify/mock"

	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
)

// MockParkhausRepository ist ein testify/mock-basierter Mock des Repositories.
type MockParkhausRepository struct {
	mock.Mock
}

func (m *MockParkhausRepository) FindByID(id uint) (*model.Parkhaus, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Parkhaus), args.Error(1)
}

func (m *MockParkhausRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *MockParkhausRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockParkhausRepository) Search(
	criteria repository.SearchCriteria, offset, limit int,
) ([]model.Parkhaus, int64, error) {
	args := m.Called(criteria, offset, limit)
	return args.Get(0).([]model.Parkhaus), args.Get(1).(int64), args.Error(2)
}

func (m *MockParkhausRepository) Create(p *model.Parkhaus) error {
	args := m.Called(p)
	// Simuliere das Setzen der ID durch die DB (wie GORM es tut).
	if args.Error(0) == nil {
		p.ID = 1
	}
	return args.Error(0)
}

func (m *MockParkhausRepository) Update(p *model.Parkhaus) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockParkhausRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockParkhausRepository) CountAutos(parkhausID uint) (int64, error) {
	args := m.Called(parkhausID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockParkhausRepository) AddAuto(auto *model.Auto) error {
	args := m.Called(auto)
	// Simuliere das Setzen der ID durch die DB (wie GORM es tut).
	if args.Error(0) == nil {
		auto.ID = 1
	}
	return args.Error(0)
}

func (m *MockParkhausRepository) FindFile(parkhausID uint) (*model.ParkhausFile, error) {
	args := m.Called(parkhausID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ParkhausFile), args.Error(1)
}

func (m *MockParkhausRepository) ReplaceFile(file *model.ParkhausFile) error {
	args := m.Called(file)
	return args.Error(0)
}
