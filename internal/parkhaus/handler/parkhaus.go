package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/parkhaus/repository"
	"parkhaus-2/internal/parkhaus/service"
	"parkhaus-2/internal/problemdetails"
)

// bekannteSuchParameter sind die für GET /rest erlaubten Query-Parameter.
var bekannteSuchParameter = map[string]bool{
	"name":           true,
	"kapazitaet":     true,
	"tarifProStunde": true,
	"page":           true,
	"size":           true,
	"count-only":     true,
}

// ParkhausHandler bündelt die HTTP-Handler für Parkhäuser.
type ParkhausHandler struct {
	svc *service.ParkhausService
}

// NewParkhausHandler erstellt einen neuen Handler.
func NewParkhausHandler(svc *service.ParkhausService) *ParkhausHandler {
	return &ParkhausHandler{svc: svc}
}

// RegisterRoutes registriert alle Routen (Health + /rest).
func (h *ParkhausHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health/liveness", health)
	r.GET("/health/readiness", health)

	rest := r.Group("/rest")
	{
		rest.GET("", h.Search)
		rest.GET("/:id", h.FindByID)
		rest.GET("/file/:id", h.DownloadFile)
		rest.POST("", h.Create)
		rest.PUT("/:id", h.Update)
		rest.POST("/:id", h.UploadFile)
		rest.POST("/:id/autos", h.AddAuto)
		rest.DELETE("/:id", h.Delete)
	}
}

// health liefert den Health-Status.
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

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

// Create behandelt POST /rest.
func (h *ParkhausHandler) Create(c *gin.Context) {
	var dto model.CreateParkhausDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		problemdetails.Write(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.svc.Create(dto)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("%s/%d", baseURL(c), id))
	c.Status(http.StatusCreated)
}

// Update behandelt PUT /rest/:id.
func (h *ParkhausHandler) Update(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	var dto model.UpdateParkhausDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		problemdetails.Write(c, http.StatusBadRequest, err.Error())
		return
	}

	ifMatch := c.GetHeader("If-Match")
	newVersion, err := h.svc.Update(id, ifMatch, dto)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	c.Header("ETag", fmt.Sprintf("\"%d\"", newVersion))
	c.Status(http.StatusNoContent)
}

// AddAuto behandelt POST /rest/:id/autos.
func (h *ParkhausHandler) AddAuto(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	var dto model.CreateAutoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		problemdetails.Write(c, http.StatusBadRequest, err.Error())
		return
	}

	autoID, err := h.svc.AddAuto(id, dto)
	if err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("%s/%d/autos/%d", baseURL(c), id, autoID))
	c.Status(http.StatusCreated)
}

// Delete behandelt DELETE /rest/:id (idempotent, immer 204).
func (h *ParkhausHandler) Delete(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		// Nicht-numerische ID: trotzdem 204 (idempotent).
		c.Status(http.StatusNoContent)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		problemdetails.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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

// UploadFile behandelt POST /rest/:id (multipart/form-data, Feld "file").
func (h *ParkhausHandler) UploadFile(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		problemdetails.Write(c, http.StatusBadRequest, "Ungueltiges Formular.")
		return
	}
	files := form.File["file"]
	if len(files) != 1 {
		problemdetails.Write(c, http.StatusBadRequest, "Es muss genau eine Datei im Feld 'file' uebergeben werden.")
		return
	}

	fileHeader := files[0]
	opened, err := fileHeader.Open()
	if err != nil {
		problemdetails.Write(c, http.StatusBadRequest, "Datei konnte nicht gelesen werden.")
		return
	}
	defer opened.Close()

	data := make([]byte, fileHeader.Size)
	if _, err := opened.Read(data); err != nil && fileHeader.Size > 0 {
		problemdetails.Write(c, http.StatusBadRequest, "Datei konnte nicht gelesen werden.")
		return
	}

	mimetype := fileHeader.Header.Get("Content-Type")
	if err := h.svc.ReplaceFile(id, fileHeader.Filename, mimetype, data); err != nil {
		problemdetails.HandleError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("%s/file/%d", baseURL(c), id))
	c.Status(http.StatusNoContent)
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

// baseURL liefert die Basis-URL des /rest-Pfads für Location-Header.
func baseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/rest", scheme, c.Request.Host)
}
