package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
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

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (st *sourceThrottler) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		remoteAddr, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: err.Error(),
					Code:  shared.ReturnCodeInternalError,
				},
			)
			return
		}

		st.mutRequests.Lock()
		requests := st.sourceRequests[remoteAddr]
		isQuotaReached := requests >= st.maxNumRequests
		st.sourceRequests[remoteAddr]++
		st.mutRequests.Unlock()

		if isQuotaReached {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Sprintf("%s for address %s", ErrTooManyRequests.Error(), remoteAddr),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
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
