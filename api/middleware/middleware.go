package middleware

import (
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `facade` context variable
type Handler interface {
}

// WithFacade middleware will set up a facade object in the gin context
func WithFacade(facade Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("facade", facade)
		c.Next()
	}
}
