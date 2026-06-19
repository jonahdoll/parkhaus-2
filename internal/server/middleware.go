package server

import (
	"log/slog"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

// corsOrigins sind die erlaubten Origins gemäß Spezifikation.
var corsOrigins = []string{
	"https://localhost:4200",
	"http://localhost:4200",
	"https://localhost:5173",
	"http://localhost:5173",
	"http://localhost:8843",
	"http://localhost:8880",
}

// secureHeaders setzt sicherheitsrelevante Header.
func secureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Next()
	}
}

// cors implementiert die CORS-Behandlung für die erlaubten Origins.
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && slices.Contains(corsOrigins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Content-Type, If-Match, If-None-Match, Accept")
			c.Header("Access-Control-Expose-Headers", "ETag, Location, Content-Type")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// requestLogger protokolliert Requests nur im Debug-Level.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Debug("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"dauer", time.Since(start).String(),
		)
	}
}
