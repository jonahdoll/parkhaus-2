package server

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"parkhaus-2/internal/parkhaus/handler"
	"parkhaus-2/internal/parkhaus/repository"
	"parkhaus-2/internal/parkhaus/service"
)

// New erzeugt den konfigurierten Gin-Router mit allen Middlewares und Routen.
func New(database *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(secureHeaders())
	router.Use(cors())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	repo := repository.NewParkhausRepository(database)
	svc := service.NewParkhausService(repo)
	parkhausHandler := handler.NewParkhausHandler(svc)
	parkhausHandler.RegisterRoutes(router)

	return router
}
