package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		fmt.Printf("[INFO] %s | %3d %s | %12v | %15s | %-8s | %s\n",
			time.Now().Format(time.RFC3339),
			c.Writer.Status(),
			c.Request.Method,
			duration,
			c.ClientIP(),
			c.Errors.String(),
			path,
		)
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors.Last().Error(),
			})
		}
	}
}
