package service

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"parkhaus-2/internal/parkhaus/apperr"
	"parkhaus-2/internal/parkhaus/model"
)

// --- Create ---

func TestParkhausService_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		dto       model.CreateParkhausDTO
		mockSetup func(*MockParkhausRepository)
		wantID    uint
		wantErr   bool
		errType   error
	}{
		{
			name: "Erfolg",
			dto: model.CreateParkhausDTO{
				Name:           "Neues Parkhaus",
				Kapazitaet:     10,
				TarifProStunde: decimal.NewFromFloat(2.50),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
				Autos: []model.CreateAutoDTO{
					{Kennzeichen: "KA-AB-123", Einfahrtszeit: now, Kundentyp: model.KundentypBasis},
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("ExistsByName", "Neues Parkhaus").Return(false, nil)
				m.On("Create", mock.MatchedBy(func(p *model.Parkhaus) bool {
					return p.Name == "Neues Parkhaus" && p.Kapazitaet == 10 &&
						len(p.Autos) == 1 && p.Adresse != nil
				})).Return(nil)
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name: "Validierungsfehler",
			dto: model.CreateParkhausDTO{
				Name:           "",
				Kapazitaet:     10,
				TarifProStunde: decimal.NewFromFloat(2.50),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				// Keine Mock-Erwartungen, da die Validierung vor dem Repo-Zugriff fehlschlägt.
			},
			wantID:  0,
			wantErr: true,
			errType: &apperr.ValidationError{},
		},
		{
			name: "KapazitaetUeberschritten",
			dto: model.CreateParkhausDTO{
				Name:           "Test",
				Kapazitaet:     2,
				TarifProStunde: decimal.NewFromFloat(1.00),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
				Autos: []model.CreateAutoDTO{
					{Kennzeichen: "KA-1", Einfahrtszeit: now, Kundentyp: model.KundentypBasis},
					{Kennzeichen: "KA-2", Einfahrtszeit: now, Kundentyp: model.KundentypBasis},
					{Kennzeichen: "KA-3", Einfahrtszeit: now, Kundentyp: model.KundentypBasis},
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				// Keine Mock-Erwartungen, da die Kapazitätsprüfung vor dem Repo-Zugriff fehlschlägt.
			},
			wantID:  0,
			wantErr: true,
			errType: &apperr.KapazitaetUeberschrittenError{},
		},
		{
			name: "NameExistiertBereits",
			dto: model.CreateParkhausDTO{
				Name:           "Existierendes",
				Kapazitaet:     5,
				TarifProStunde: decimal.NewFromFloat(1.00),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("ExistsByName", "Existierendes").Return(true, nil)
			},
			wantID:  0,
			wantErr: true,
			errType: &apperr.ParkhausExistsError{},
		},
		{
			name: "DBFehler_ExistsByName",
			dto: model.CreateParkhausDTO{
				Name:           "Test",
				Kapazitaet:     5,
				TarifProStunde: decimal.NewFromFloat(1.00),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("ExistsByName", "Test").Return(false, errors.New("db error"))
			},
			wantID:  0,
			wantErr: true,
			errType: errors.New("db error"),
		},
		{
			name: "DBFehler_Create",
			dto: model.CreateParkhausDTO{
				Name:           "Neues Parkhaus",
				Kapazitaet:     10,
				TarifProStunde: decimal.NewFromFloat(2.50),
				Adresse: model.CreateAdresseDTO{
					PLZ: "76131", Ort: "Karlsruhe", Strasse: "Hauptstr.", Hausnummer: "1",
				},
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("ExistsByName", "Neues Parkhaus").Return(false, nil)
				m.On("Create", mock.Anything).Return(errors.New("db error"))
			},
			wantID:  0,
			wantErr: true,
			errType: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			gotID, err := svc.Create(tt.dto)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Zero(t, gotID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Update ---

func TestParkhausService_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		ifMatch   string
		dto       model.UpdateParkhausDTO
		mockSetup func(*MockParkhausRepository)
		want      int
		wantErr   bool
		errType   error
	}{
		{
			name:    "Erfolg",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name:           "Geändert",
				Kapazitaet:     20,
				TarifProStunde: decimal.NewFromFloat(3.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Alt", Version: 0, Kapazitaet: 10,
				}, nil)
				m.On("ExistsByName", "Geändert").Return(false, nil)
				m.On("Update", mock.MatchedBy(func(p *model.Parkhaus) bool {
					return p.Name == "Geändert" && p.Kapazitaet == 20 && p.Version == 1
				})).Return(nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name:    "IfMatchFehlt",
			id:      1,
			ifMatch: "",
			dto: model.UpdateParkhausDTO{
				Name: "Test", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {},
			want:      0,
			wantErr:   true,
			errType:   &apperr.PreconditionRequiredError{},
		},
		{
			name:    "IfMatchUngueltig",
			id:      1,
			ifMatch: "abc",
			dto: model.UpdateParkhausDTO{
				Name: "Test", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {},
			want:      0,
			wantErr:   true,
			errType:   &apperr.VersionInvalidError{},
		},
		{
			name:    "Validierungsfehler",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {},
			want:      0,
			wantErr:   true,
			errType:   &apperr.ValidationError{},
		},
		{
			name:    "NotFound",
			id:      999,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "Test", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(999)).Return(nil, nil)
			},
			want:    0,
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name:    "VersionOutdated",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "Test", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Alt", Version: 2, Kapazitaet: 10,
				}, nil)
			},
			want:    0,
			wantErr: true,
			errType: &apperr.VersionOutdatedError{},
		},
		{
			name:    "NameExistiertBereits",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "AndererName", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Alt", Version: 0, Kapazitaet: 10,
				}, nil)
				m.On("ExistsByName", "AndererName").Return(true, nil)
			},
			want:    0,
			wantErr: true,
			errType: &apperr.ParkhausExistsError{},
		},
		{
			name:    "DBFehler_FindByID",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "Test", Kapazitaet: 5, TarifProStunde: decimal.NewFromFloat(1.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(nil, errors.New("db error"))
			},
			want:    0,
			wantErr: true,
			errType: errors.New("db error"),
		},
		{
			name:    "DBFehler_Update",
			id:      1,
			ifMatch: `"0"`,
			dto: model.UpdateParkhausDTO{
				Name: "Geändert", Kapazitaet: 20, TarifProStunde: decimal.NewFromFloat(3.00),
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Alt", Version: 0, Kapazitaet: 10,
				}, nil)
				m.On("ExistsByName", "Geändert").Return(false, nil)
				m.On("Update", mock.Anything).Return(errors.New("db error"))
			},
			want:    0,
			wantErr: true,
			errType: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			got, err := svc.Update(tt.id, tt.ifMatch, tt.dto)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Zero(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Delete ---

func TestParkhausService_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockParkhausRepository)
		wantErr   bool
	}{
		{
			name: "Erfolg",
			id:   1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Delete", uint(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Idempotent_NichtVorhanden",
			id:   999,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Delete", uint(999)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "DBFehler",
			id:   1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Delete", uint(1)).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			err := svc.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- AddAuto ---

func TestParkhausService_AddAuto(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		parkhausID uint
		dto        model.CreateAutoDTO
		mockSetup  func(*MockParkhausRepository)
		wantID     uint
		wantErr    bool
		errType    error
	}{
		{
			name:       "Erfolg",
			parkhausID: 1,
			dto: model.CreateAutoDTO{
				Kennzeichen: "KA-XY-123", Einfahrtszeit: now, Kundentyp: model.KundentypPremium,
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Test", Kapazitaet: 10,
				}, nil)
				m.On("CountAutos", uint(1)).Return(int64(3), nil)
				m.On("AddAuto", mock.MatchedBy(func(a *model.Auto) bool {
					return a.Kennzeichen == "KA-XY-123" && a.ParkhausID == 1
				})).Return(nil)
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name:       "Validierungsfehler",
			parkhausID: 1,
			dto: model.CreateAutoDTO{
				Kennzeichen: "", Einfahrtszeit: now, Kundentyp: model.KundentypBasis,
			},
			mockSetup: func(m *MockParkhausRepository) {},
			wantID:    0,
			wantErr:   true,
			errType:   &apperr.ValidationError{},
		},
		{
			name:       "NotFound",
			parkhausID: 999,
			dto: model.CreateAutoDTO{
				Kennzeichen: "KA-XY-123", Einfahrtszeit: now, Kundentyp: model.KundentypBasis,
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(999)).Return(nil, nil)
			},
			wantID:  0,
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name:       "KapazitaetVoll",
			parkhausID: 1,
			dto: model.CreateAutoDTO{
				Kennzeichen: "KA-XY-123", Einfahrtszeit: now, Kundentyp: model.KundentypBasis,
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Test", Kapazitaet: 5,
				}, nil)
				m.On("CountAutos", uint(1)).Return(int64(5), nil)
			},
			wantID:  0,
			wantErr: true,
			errType: &apperr.KapazitaetUeberschrittenError{},
		},
		{
			name:       "DBFehler_FindByID",
			parkhausID: 1,
			dto: model.CreateAutoDTO{
				Kennzeichen: "KA-XY-123", Einfahrtszeit: now, Kundentyp: model.KundentypBasis,
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(nil, errors.New("db error"))
			},
			wantID:  0,
			wantErr: true,
			errType: errors.New("db error"),
		},
		{
			name:       "DBFehler_AddAuto",
			parkhausID: 1,
			dto: model.CreateAutoDTO{
				Kennzeichen: "KA-XY-123", Einfahrtszeit: now, Kundentyp: model.KundentypBasis,
			},
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{
					ID: 1, Name: "Test", Kapazitaet: 10,
				}, nil)
				m.On("CountAutos", uint(1)).Return(int64(3), nil)
				m.On("AddAuto", mock.Anything).Return(errors.New("db error"))
			},
			wantID:  0,
			wantErr: true,
			errType: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			gotID, err := svc.AddAuto(tt.parkhausID, tt.dto)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Zero(t, gotID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- ReplaceFile ---

func TestParkhausService_ReplaceFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parkhausID uint
		filename   string
		mimetype   string
		data       []byte
		mockSetup  func(*MockParkhausRepository)
		wantErr    bool
		errType    error
	}{
		{
			name:       "Erfolg",
			parkhausID: 1,
			filename:   "dokument.pdf",
			mimetype:   "application/pdf",
			data:       []byte("pdf content"),
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{ID: 1, Name: "Test"}, nil)
				m.On("ReplaceFile", mock.MatchedBy(func(f *model.ParkhausFile) bool {
					return f.ParkhausID == 1 && f.Filename == "dokument.pdf" &&
						f.Mimetype == "application/pdf" && string(f.Data) == "pdf content"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "NotFound",
			parkhausID: 999,
			filename:   "test.txt",
			mimetype:   "text/plain",
			data:       []byte("hello"),
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(999)).Return(nil, nil)
			},
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name:       "DBFehler_FindByID",
			parkhausID: 1,
			filename:   "test.txt",
			mimetype:   "text/plain",
			data:       []byte("hello"),
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errType: errors.New("db error"),
		},
		{
			name:       "DBFehler_ReplaceFile",
			parkhausID: 1,
			filename:   "test.txt",
			mimetype:   "text/plain",
			data:       []byte("hello"),
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{ID: 1, Name: "Test"}, nil)
				m.On("ReplaceFile", mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
			errType: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			err := svc.ReplaceFile(tt.parkhausID, tt.filename, tt.mimetype, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
