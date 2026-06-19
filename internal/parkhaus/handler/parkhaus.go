package handler

import (
	"github.com/gin-gonic/gin"

	"parkhaus-2/internal/parkhaus/service"
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
