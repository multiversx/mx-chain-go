package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
)

// WithElrondFacade middleware will set up an ElrondFacade object in the gin context
func WithElrondFacade(elrondFacade node.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", elrondFacade)
		c.Next()
	}
}