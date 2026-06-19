package model

// Pagination-Defaults gemäß Spezifikation.
const (
	DefaultPageSize   = 5
	MaxPageSize       = 100
	DefaultPageNumber = 0
)

// PageInfo beschreibt die Metadaten einer Seite.
type PageInfo struct {
	Size          int   `json:"size"`
	Number        int   `json:"number"` // 0-basiert
	TotalElements int64 `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
}

// Page ist das generische Seiten-Antwortobjekt für GET /rest.
type Page struct {
	Content []ParkhausDTO `json:"content"`
	Page    PageInfo      `json:"page"`
}

// NewPage erzeugt ein Page-Objekt und berechnet totalPages = ceil(total/size).
func NewPage(content []ParkhausDTO, size, number int, totalElements int64) Page {
	totalPages := 0
	if size > 0 {
		totalPages = int((totalElements + int64(size) - 1) / int64(size))
	}
	return Page{
		Content: content,
		Page: PageInfo{
			Size:          size,
			Number:        number,
			TotalElements: totalElements,
			TotalPages:    totalPages,
		},
	}
}
