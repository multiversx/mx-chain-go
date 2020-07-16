package middleware_test

import (
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func startNodeServerEndpointThrottler(handler interface{}, throttlerName string) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	if handler == nil {
		ws.Use(middleware.CreateEndpointThrottler(throttlerName))
	} else {
		ws.Use(middleware.WithFacade(handler), middleware.CreateEndpointThrottler(throttlerName))
	}
	ginAddressRoutes := ws.Group("/address")
	addressRoutes, _ := wrapper.NewRouterWrapper("address", ginAddressRoutes, getRoutesConfig())
	address.Routes(addressRoutes)
	return ws
}

func TestCreateEndpointThrottler_NilContextShouldError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerEndpointThrottler(nil, "test")
	mutResponses := sync.Mutex{}
	responses := make(map[int]int)

	makeRequestGlobalThrottler(ws, &mutResponses, responses)

	mutResponses.Lock()
	defer mutResponses.Unlock()

	assert.Equal(t, 1, responses[http.StatusInternalServerError])
}

func TestCreateEndpointThrottler_WrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServerEndpointThrottler(&struct{}{}, "test")
	mutResponses := sync.Mutex{}
	responses := make(map[int]int)

	makeRequestGlobalThrottler(ws, &mutResponses, responses)

	mutResponses.Lock()
	defer mutResponses.Unlock()

	assert.Equal(t, 1, responses[http.StatusInternalServerError])
}

func TestCreateEndpointThrottler_NoThrottlerShouldExecute(t *testing.T) {
	t.Parallel()

	numCalls := uint32(0)
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			atomic.AddUint32(&numCalls, 1)

			return big.NewInt(10), nil
		},
	}
	ws := startNodeServerEndpointThrottler(&facade, "test")
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
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			atomic.AddUint32(&numCalls, 1)

			return big.NewInt(10), nil
		},
		GetThrottlerForEndpointCalled: func(endpoint string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool {
					return false
				},
			}, true
		},
	}
	ws := startNodeServerEndpointThrottler(&facade, "test")
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
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			atomic.AddUint32(&numCalls, 1)

			return big.NewInt(10), nil
		},
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
	ws := startNodeServerEndpointThrottler(&facade, "test")
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
