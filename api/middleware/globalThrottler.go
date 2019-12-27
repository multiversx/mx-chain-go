package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// globalThrottler is a middleware global limiter used to limit total number of simultaneous requests
type globalThrottler struct {
	queue chan struct{}
}

// NewGlobalThrottler creates a new instance of a globalThrottler
func NewGlobalThrottler(maxConnections uint32) (*globalThrottler, error) {
	if maxConnections == 0 {
		return nil, ErrInvalidMaxNumRequests
	}

	return &globalThrottler{
		queue: make(chan struct{}, maxConnections),
	}, nil
}

// Limit returns the handler func used by the gin server to limit simultaneous requests
func (gt *globalThrottler) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		select {
		case gt.queue <- struct{}{}:
		default:
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
		<-gt.queue
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (gt *globalThrottler) IsInterfaceNil() bool {
	return gt == nil
}
