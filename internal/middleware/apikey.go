package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const HeaderName = "X-API-Key"

func APIKey(expected string) gin.HandlerFunc {
	expectedBytes := []byte(expected)
	return func(c *gin.Context) {
		got := c.GetHeader(HeaderName)
		if got == "" || subtle.ConstantTimeCompare([]byte(got), expectedBytes) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
