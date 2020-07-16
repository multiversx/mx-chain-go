package middleware

import "github.com/gin-gonic/gin"

// WithTestingElrondFacade middleware will set up an ElrondFacade object in the gin context
// should only be used in testing and not in conjunction with other middlewares as c.Next() instruction
// is not safe to be used multiple times in the same context.
func WithTestingElrondFacade(elrondFacade ElrondHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", elrondFacade)
		c.Next()
	}
}
