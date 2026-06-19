package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/problemdetails"
)

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

// baseURL liefert die Basis-URL des /rest-Pfads für Location-Header.
func baseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/rest", scheme, c.Request.Host)
}
