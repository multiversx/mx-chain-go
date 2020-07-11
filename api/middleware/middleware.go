package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/gin-gonic/gin"
)

// ElrondHandler interface defines methods that can be used from `elrondFacade` context variable
//TODO rename ElrondHandler (take out Elrond part)
type ElrondHandler interface {
	IsInterfaceNil() bool
}

// middleWare can decide if a message can be processed or not.
// It is very important that we have a maximum of one instance of the middleware implementations as the
// call c.Next() should not be done multiple times
type middleWare struct {
	facade         ElrondHandler
	queue          chan struct{}
	mutRequests    sync.Mutex
	sourceRequests map[string]uint32
	maxNumRequests uint32
}

// NewMiddleware creates a new instance of a gin middleware
func NewMiddleware(
	facade ElrondHandler,
	maxConcurrentRequests uint32,
	maxNumRequestsPerAddress uint32,
) (*middleWare, error) {
	if check.IfNil(facade) {
		return nil, ErrNilFacade
	}
	if maxConcurrentRequests == 0 {
		return nil, ErrInvalidMaxNumConcurrentRequests
	}
	if maxNumRequestsPerAddress == 0 {
		return nil, ErrInvalidMaxNumRequests
	}

	return &middleWare{
		queue:          make(chan struct{}, maxConcurrentRequests),
		facade:         facade,
		mutRequests:    sync.Mutex{},
		sourceRequests: make(map[string]uint32),
		maxNumRequests: maxNumRequestsPerAddress,
	}, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (m *middleWare) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", m.facade)

		status, err := m.checkAddressQuota(c)
		if err != nil {
			c.AbortWithStatusJSON(
				status,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: err.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
			return
		}

		select {
		case m.queue <- struct{}{}:
		default:
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: ErrTooManyRequests.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
			return
		}

		c.Next()

		<-m.queue
	}
}

func (m *middleWare) checkAddressQuota(c *gin.Context) (int, error) {
	remoteAddr, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	m.mutRequests.Lock()
	requests := m.sourceRequests[remoteAddr]
	isQuotaReached := requests >= m.maxNumRequests
	m.sourceRequests[remoteAddr]++
	m.mutRequests.Unlock()

	if isQuotaReached {
		return http.StatusTooManyRequests, fmt.Errorf("%w for address %s", ErrTooManyRequests, remoteAddr)
	}

	return http.StatusOK, nil
}

// Reset resets all accumulated counters
func (m *middleWare) Reset() {
	m.mutRequests.Lock()
	m.sourceRequests = make(map[string]uint32)
	m.mutRequests.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *middleWare) IsInterfaceNil() bool {
	return m == nil
}
