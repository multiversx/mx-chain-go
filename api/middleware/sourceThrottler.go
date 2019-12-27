package middleware

import (
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// globalThrottler is a middleware limiter used to limit total number of requests originating from the same source
type sourceThrottler struct {
	mutRequests    sync.Mutex
	sourceRequests map[string]uint32
	maxNumRequests uint32
}

// NewSourceThrottler creates a new instance of a sourceThrottler
func NewSourceThrottler(maxNumRequests uint32) (*sourceThrottler, error) {
	if maxNumRequests == 0 {
		return nil, ErrInvalidMaxNumRequests
	}

	return &sourceThrottler{
		mutRequests:    sync.Mutex{},
		sourceRequests: make(map[string]uint32),
		maxNumRequests: maxNumRequests,
	}, nil
}

// Limit returns the handler func used by the gin server to limit requests originating from the same source
func (st *sourceThrottler) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		remoteAddr, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		st.mutRequests.Lock()
		requests := st.sourceRequests[remoteAddr]
		isQuotaReached := requests >= st.maxNumRequests
		st.sourceRequests[remoteAddr]++
		st.mutRequests.Unlock()

		if isQuotaReached {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

// Reset resets all accumulated counters
func (st *sourceThrottler) Reset() {
	st.mutRequests.Lock()
	st.sourceRequests = make(map[string]uint32)
	st.mutRequests.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (st *sourceThrottler) IsInterfaceNil() bool {
	return st == nil
}
