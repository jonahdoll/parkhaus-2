package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"parkhaus-2/internal/parkhaus/apperr"
	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
)

// --- FindByID ---

func TestParkhausService_FindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockParkhausRepository)
		want      *model.Parkhaus
		wantErr   bool
		errType   error
	}{
		{
			name: "Erfolg",
			id:   1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(&model.Parkhaus{ID: 1, Name: "Test"}, nil)
			},
			want:    &model.Parkhaus{ID: 1, Name: "Test"},
			wantErr: false,
		},
		{
			name: "NotFound",
			id:   999,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(999)).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name: "DBFehler",
			id:   1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindByID", uint(1)).Return(nil, errors.New("db error"))
			},
			want:    nil,
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

			got, err := svc.FindByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- CountAll ---

func TestParkhausService_CountAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mockSetup func(*MockParkhausRepository)
		want      int64
		wantErr   bool
	}{
		{
			name: "Erfolg",
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Count").Return(int64(6), nil)
			},
			want:    6,
			wantErr: false,
		},
		{
			name: "DBFehler",
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Count").Return(int64(0), errors.New("db error"))
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockParkhausRepository)
			tt.mockSetup(mockRepo)

			svc := NewParkhausService(mockRepo)

			got, err := svc.CountAll()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Search ---

func TestParkhausService_Search(t *testing.T) {
	t.Parallel()

	now := model.DefaultPageNumber
	pageSize := model.DefaultPageSize

	tests := []struct {
		name      string
		criteria  repository.SearchCriteria
		page      int
		size      int
		mockSetup func(*MockParkhausRepository)
		want      *model.Page
		wantErr   bool
		errType   error
	}{
		{
			name:     "Erfolg_ErsteSeite",
			criteria: repository.SearchCriteria{},
			page:     1,
			size:     5,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Search", mock.Anything, 0, 5).Return(
					[]model.Parkhaus{
						{ID: 1, Name: "Parkhaus Aachen"},
						{ID: 2, Name: "Parkhaus Berlin"},
					},
					int64(2),
					nil,
				)
			},
			want: func() *model.Page {
				p := model.NewPage(
					[]model.ParkhausDTO{
						{ID: 1, Name: "Parkhaus Aachen"},
						{ID: 2, Name: "Parkhaus Berlin"},
					},
					5, 0, 2,
				)
				return &p
			}(),
			wantErr: false,
		},
		{
			name:     "Erfolg_DefaultPagination",
			criteria: repository.SearchCriteria{},
			page:     0,
			size:     0,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Search", mock.Anything, now, pageSize).Return(
					[]model.Parkhaus{
						{ID: 1, Name: "Parkhaus Aachen"},
					},
					int64(1),
					nil,
				)
			},
			want: func() *model.Page {
				p := model.NewPage(
					[]model.ParkhausDTO{
						{ID: 1, Name: "Parkhaus Aachen"},
					},
					pageSize, now, 1,
				)
				return &p
			}(),
			wantErr: false,
		},
		{
			name:     "KeineTreffer_NotFound",
			criteria: repository.SearchCriteria{},
			page:     1,
			size:     5,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Search", mock.Anything, 0, 5).Return([]model.Parkhaus{}, int64(0), nil)
			},
			want:    nil,
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name:     "DBFehler",
			criteria: repository.SearchCriteria{},
			page:     1,
			size:     5,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("Search", mock.Anything, 0, 5).Return([]model.Parkhaus{}, int64(0), errors.New("db error"))
			},
			want:    nil,
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

			got, err := svc.Search(tt.criteria, tt.page, tt.size)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- FindFile ---

func TestParkhausService_FindFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parkhausID uint
		mockSetup  func(*MockParkhausRepository)
		want       *model.ParkhausFile
		wantErr    bool
		errType    error
	}{
		{
			name:       "Erfolg",
			parkhausID: 1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindFile", uint(1)).Return(&model.ParkhausFile{
					ParkhausID: 1,
					Filename:   "test.pdf",
					Mimetype:   "application/pdf",
					Data:       []byte("test data"),
				}, nil)
			},
			want: &model.ParkhausFile{
				ParkhausID: 1,
				Filename:   "test.pdf",
				Mimetype:   "application/pdf",
				Data:       []byte("test data"),
			},
			wantErr: false,
		},
		{
			name:       "NotFound",
			parkhausID: 999,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindFile", uint(999)).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
			errType: &apperr.NotFoundError{},
		},
		{
			name:       "DBFehler",
			parkhausID: 1,
			mockSetup: func(m *MockParkhausRepository) {
				m.On("FindFile", uint(1)).Return(nil, errors.New("db error"))
			},
			want:    nil,
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

			got, err := svc.FindFile(tt.parkhausID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
