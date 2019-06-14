package middleware

import (
	"github.com/gin-gonic/gin"
)

// ElrondHandler interface defines methods that can be used from `elrondFacade` context variable
type ElrondHandler interface {

}

// WithElrondFacade middleware will set up an ElrondFacade object in the gin context
func WithElrondFacade(elrondFacade ElrondHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", elrondFacade)
		c.Next()
	}
}
