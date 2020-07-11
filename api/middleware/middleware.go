package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/middleware")
var maxSecondsBetweenReset = int64(300)

// ElrondHandler interface defines methods that can be used from `elrondFacade` context variable
//TODO rename ElrondHandler (take out Elrond part)
type ElrondHandler interface {
	IsInterfaceNil() bool
}

// middleWare can decide if a message can be processed or not.
// It is very important that we have a maximum of one instance of the middleware implementations as the
// call c.Next() should not be done multiple times
type middleWare struct {
	facade              ElrondHandler
	queue               chan struct{}
	mutRequests         sync.Mutex
	sourceRequests      map[string]uint32
	maxNumRequests      uint32
	lastResetTimestamp  int64
	resetFailedActionFn func(lastResetTimestamp int64, maxSecondsBetweenReset int64)

	//TODO remove this debug data before merging the PR
	mutDebug  sync.RWMutex
	debugInfo map[string]int
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

	mw := &middleWare{
		queue:              make(chan struct{}, maxConcurrentRequests),
		facade:             facade,
		mutRequests:        sync.Mutex{},
		sourceRequests:     make(map[string]uint32),
		maxNumRequests:     maxNumRequestsPerAddress,
		debugInfo:          make(map[string]int),
		lastResetTimestamp: time.Now().Unix(),
	}
	mw.resetFailedActionFn = mw.resetFailedAction

	return mw, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (m *middleWare) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", m.facade)

		lastResetTimestamp := atomic.LoadInt64(&m.lastResetTimestamp)
		maxSeconds := atomic.LoadInt64(&maxSecondsBetweenReset)
		if lastResetTimestamp+maxSeconds < time.Now().Unix() {
			m.resetFailedActionFn(lastResetTimestamp, maxSeconds)
		}

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
			m.mutDebug.Lock()
			m.debugInfo[c.Request.URL.Path]++
			m.mutDebug.Unlock()
		default:
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: ErrTooManyRequests.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)

			output := make([]string, 0)
			m.mutDebug.Lock()
			for route, num := range m.debugInfo {
				output = append(output, fmt.Sprintf("%s: %d", route, num))
			}
			m.mutDebug.Unlock()

			log.Warn("system busy\n" + strings.Join(output, "\n"))

			return
		}

		c.Next()

		m.mutDebug.Lock()
		m.debugInfo[c.Request.URL.Path]--
		if m.debugInfo[c.Request.URL.Path] == 0 {
			delete(m.debugInfo, c.Request.URL.Path)
		}
		m.mutDebug.Unlock()

		<-m.queue
	}
}

func (m *middleWare) resetFailedAction(lastResetTimestamp int64, maxSeconds int64) {
	log.Warn("api middleware, source throttler reset failed",
		"max seconds", maxSeconds,
		"last reset timestamp", lastResetTimestamp,
		"current timestamp", time.Now().Unix(),
	)
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
	atomic.StoreInt64(&m.lastResetTimestamp, time.Now().Unix())

	m.mutRequests.Lock()
	m.sourceRequests = make(map[string]uint32)
	m.mutRequests.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *middleWare) IsInterfaceNil() bool {
	return m == nil
}
