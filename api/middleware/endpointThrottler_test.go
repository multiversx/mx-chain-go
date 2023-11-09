package middleware_test

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/stretchr/testify/assert"
)

func startNodeServerEndpointThrottler(handler func(c *gin.Context), facade interface{}, throttlerName string) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())

	ws.Use(middleware.CreateEndpointThrottlerFromFacade(throttlerName, facade))

	ginAddressRoutes := ws.Group("/address")

	ginAddressRoutes.Handle(http.MethodGet, "/:address/balance", handler)

	return ws
}

func TestCreateEndpointThrottler_NoThrottlerShouldExecute(t *testing.T) {
	t.Parallel()

	numCalls := uint32(0)
	handlerFunc := func(c *gin.Context) {
		atomic.AddUint32(&numCalls, 1)
		c.JSON(200, "ok")
	}

	ws := startNodeServerEndpointThrottler(handlerFunc, &mock.FacadeStub{}, "test")
	mutResponses := sync.Mutex{}
	responses := make(map[int]int)

	makeRequestGlobalThrottler(ws, &mutResponses, responses)

	mutResponses.Lock()
	defer mutResponses.Unlock()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
	assert.Equal(t, 1, responses[http.StatusOK])
}

func TestCreateEndpointThrottler_ThrottlerCanNotProcessShouldNotExecute(t *testing.T) {
	t.Parallel()

	numCalls := uint32(0)
	facade := mock.FacadeStub{
		GetThrottlerForEndpointCalled: func(endpoint string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool {
					return false
				},
			}, true
		},
	}

	handlerFunc := func(c *gin.Context) {
		atomic.AddUint32(&numCalls, 1)
	}
	ws := startNodeServerEndpointThrottler(handlerFunc, &facade, "test")
	mutResponses := sync.Mutex{}
	responses := make(map[int]int)

	makeRequestGlobalThrottler(ws, &mutResponses, responses)

	mutResponses.Lock()
	defer mutResponses.Unlock()

	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalls))
	assert.Equal(t, 1, responses[http.StatusTooManyRequests])
}

func TestCreateEndpointThrottler_ThrottlerStartShouldExecute(t *testing.T) {
	t.Parallel()

	numCalls := uint32(0)
	numStart := uint32(0)
	numEnd := uint32(0)
	facade := mock.FacadeStub{
		GetThrottlerForEndpointCalled: func(endpoint string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool {
					return true
				},
				StartProcessingCalled: func() {
					atomic.AddUint32(&numStart, 1)
				},
				EndProcessingCalled: func() {
					atomic.AddUint32(&numEnd, 1)
				},
			}, true
		},
	}
	handlerFunc := func(c *gin.Context) {
		atomic.AddUint32(&numCalls, 1)
	}
	ws := startNodeServerEndpointThrottler(handlerFunc, &facade, "test")
	mutResponses := sync.Mutex{}
	responses := make(map[int]int)

	makeRequestGlobalThrottler(ws, &mutResponses, responses)

	mutResponses.Lock()
	defer mutResponses.Unlock()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numStart))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numEnd))
	assert.Equal(t, 1, responses[http.StatusOK])
}
