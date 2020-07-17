package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/middleware")

// globalThrottler is a middleware global limiter used to limit total number of simultaneous requests
type globalThrottler struct {
	queue            chan struct{}
	mutDebugRequests sync.Mutex
	debugRequests    map[string]int
}

// NewGlobalThrottler creates a new instance of a globalThrottler
func NewGlobalThrottler(maxConnections uint32) (*globalThrottler, error) {
	if maxConnections == 0 {
		return nil, ErrInvalidMaxNumRequests
	}

	return &globalThrottler{
		queue:         make(chan struct{}, maxConnections),
		debugRequests: make(map[string]int),
	}, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (gt *globalThrottler) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		select {
		case gt.queue <- struct{}{}:
			gt.mutDebugRequests.Lock()
			gt.debugRequests[path]++
			gt.mutDebugRequests.Unlock()
		default:
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: ErrTooManyRequests.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)

			gt.printDebugInfo()

			return
		}

		defer gt.finish(path)

		c.Next()
	}
}

func (gt *globalThrottler) finish(path string) {
	gt.mutDebugRequests.Lock()
	gt.debugRequests[path]--
	if gt.debugRequests[path] < 1 {
		delete(gt.debugRequests, path)
	}
	gt.mutDebugRequests.Unlock()

	<-gt.queue
}

func (gt *globalThrottler) printDebugInfo() {
	gt.mutDebugRequests.Lock()
	infoLines := make([]string, 0, len(gt.debugRequests))
	for requestPath, counter := range gt.debugRequests {
		infoLines = append(infoLines, fmt.Sprintf("%s: %d", requestPath, counter))
	}
	gt.mutDebugRequests.Unlock()

	log.Debug(fmt.Sprintf("API engine stuck: \n%s", strings.Join(infoLines, "\n")))
}

// IsInterfaceNil returns true if there is no value under the interface
func (gt *globalThrottler) IsInterfaceNil() bool {
	return gt == nil
}
