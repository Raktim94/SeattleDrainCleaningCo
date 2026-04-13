package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nodedr/submify/apps/api/internal/config"
)

// SubmifyCORS sets Access-Control-* using OriginAllowed (same-host tunnel-safe, env lists, LAN relax).
func SubmifyCORS(cfg config.Config) gin.HandlerFunc {
	maxAge := int((12 * time.Hour).Seconds())
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if !OriginAllowed(origin, c.Request, cfg) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, x-api-key")
			h.Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
