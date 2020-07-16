package middleware_test

import (
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type reseter interface {
	Reset()
}

func startNodeServerSourceThrottler(handler address.FacadeHandler, maxConnections uint32) (*gin.Engine, reseter) {
	ws := gin.New()
	ws.Use(cors.Default())
	sourceThrottler, _ := middleware.NewSourceThrottler(maxConnections)
	ws.Use(sourceThrottler.MiddlewareHandlerFunc())
	ginAddressRoutes := ws.Group("/address")
	if handler != nil {
		ginAddressRoutes.Use(middleware.WithFacade(handler))
	}
	addressRoutes, _ := wrapper.NewRouterWrapper("address", ginAddressRoutes, getRoutesConfig())
	address.Routes(addressRoutes)
	return ws, sourceThrottler
}

func TestNewSourceThrottler_InvalidValueShouldErr(t *testing.T) {
	t.Parallel()

	st, err := middleware.NewSourceThrottler(0)

	assert.True(t, check.IfNil(st))
	assert.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
}

func TestNewSourceThrottler(t *testing.T) {
	t.Parallel()

	st, err := middleware.NewSourceThrottler(1)

	assert.False(t, check.IfNil(st))
	assert.Nil(t, err)
}

func TestSourceThrottler_LimitBadRequestShouldErr(t *testing.T) {
	t.Parallel()
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxConnections := uint32(1000)
	ws, _ := startNodeServerSourceThrottler(&facade, maxConnections)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestSourceThrottler_LimitUnderShouldProcessRequest(t *testing.T) {
	t.Parallel()
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxConnections := uint32(1000)
	ws, _ := startNodeServerSourceThrottler(&facade, maxConnections)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "127.0.0.1:8080"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSourceThrottler_LimitOverShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := 10
	maxConnections := uint32(maxRequests - 1)
	ws, _ := startNodeServerSourceThrottler(&facade, maxConnections)

	mutResponses := sync.Mutex{}
	responses := make(map[int]int)
	for i := 0; i < maxRequests; i++ {
		makeRequestSourceThrottler(ws, &mutResponses, responses)
	}

	mutResponses.Lock()
	assert.Equal(t, maxRequests-1, responses[http.StatusOK])
	assert.Equal(t, 1, responses[http.StatusTooManyRequests])
	mutResponses.Unlock()
}

func TestSourceThrottler_LimitResetShouldWork(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := 10
	maxConnections := uint32(maxRequests - 1)
	ws, resetHandler := startNodeServerSourceThrottler(&facade, maxConnections)

	mutResponses := sync.Mutex{}
	responses := make(map[int]int)
	numBatches := 2
	for j := 0; j < numBatches; j++ {
		for i := 0; i < maxRequests; i++ {
			makeRequestSourceThrottler(ws, &mutResponses, responses)
		}

		resetHandler.Reset()
	}

	mutResponses.Lock()
	assert.Equal(t, numBatches*(maxRequests-1), responses[http.StatusOK])
	assert.Equal(t, numBatches, responses[http.StatusTooManyRequests])
	mutResponses.Unlock()
}

func makeRequestSourceThrottler(ws *gin.Engine, mutResponses *sync.Mutex, responses map[int]int) {
	addr := "testAddress"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "127.0.0.1:8080"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	mutResponses.Lock()
	responses[resp.Code]++
	mutResponses.Unlock()
}
