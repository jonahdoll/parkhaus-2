package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
	"parkhaus-2/internal/problemdetails"
)

// FindByID behandelt GET /rest/:id.
func (h *ParkhausHandler) FindByID(c *gin.Context) {
	if !acceptable(c) {
		c.Status(http.StatusNotAcceptable)
		return
	}

	id, ok := parseID(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	p, err := h.svc.FindByID(id)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	etag := fmt.Sprintf("\"%d\"", p.Version)
	if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", etag)
	c.JSON(http.StatusOK, model.ToParkhausDTO(*p))
}

// Search behandelt GET /rest (Suche/Liste/Count).
func (h *ParkhausHandler) Search(c *gin.Context) {
	if !acceptable(c) {
		c.Status(http.StatusNotAcceptable)
		return
	}

	query := c.Request.URL.Query()

	// count-only: Präsenz des Keys genügt.
	if query.Has("count-only") {
		count, err := h.svc.CountAll()
		if err != nil {
			problemdetails.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
		return
	}

	// Unbekannte Filter-Parameter -> 404.
	for key := range query {
		if !bekannteSuchParameter[key] {
			c.Status(http.StatusNotFound)
			return
		}
	}

	criteria, ok := buildCriteria(query)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	page := atoiDefault(query.Get("page"), 0)
	size := atoiDefault(query.Get("size"), model.DefaultPageSize)

	result, err := h.svc.Search(criteria, page, size)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// DownloadFile behandelt GET /rest/file/:id.
func (h *ParkhausHandler) DownloadFile(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	file, err := h.svc.FindFile(id)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	contentType := file.Mimetype
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	c.Data(http.StatusOK, contentType, file.Data)
}

// health liefert den Health-Status.
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

// --- Hilfsfunktionen ---

// parseID parst eine numerische ID; ok=false bei nicht-numerischem Wert.
func parseID(raw string) (uint, bool) {
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(n), true
}

// atoiDefault parst einen int oder liefert den Default.
func atoiDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

// buildCriteria baut die Suchkriterien; ok=false bei ungültigem numerischem Filter.
func buildCriteria(query map[string][]string) (repository.SearchCriteria, bool) {
	var criteria repository.SearchCriteria

	if v := first(query["name"]); v != "" {
		criteria.Name = &v
	}
	if v := first(query["kapazitaet"]); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return criteria, false
		}
		criteria.Kapazitaet = &n
	}
	if v := first(query["tarifProStunde"]); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return criteria, false
		}
		criteria.TarifProStunde = &f
	}
	return criteria, true
}

// first liefert den ersten Wert eines Slice oder "".
func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// acceptable prüft das Accept-Handling: wenn Accept gesetzt und weder */*
// noch json/html, dann nicht akzeptabel (-> 406).
func acceptable(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if accept == "" {
		return true
	}
	if strings.Contains(accept, "*/*") ||
		strings.Contains(accept, "json") ||
		strings.Contains(accept, "html") {
		return true
	}
	return false
}
