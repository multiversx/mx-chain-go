package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/stretchr/testify/assert"
)

type reseter interface {
	Reset()
}

func startNodeServerSourceThrottler(handler func(c *gin.Context), maxConnections uint32) (*gin.Engine, reseter) {
	ws := gin.New()
	ws.Use(cors.Default())
	sourceThrottler, _ := middleware.NewSourceThrottler(maxConnections)
	ws.Use(sourceThrottler.MiddlewareHandlerFunc())
	ginAddressRoutes := ws.Group("/address")

	ginAddressRoutes.Handle(http.MethodGet, "/:address/balance", handler)

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

	maxConnections := uint32(1000)
	ws, _ := startNodeServerSourceThrottler(func(c *gin.Context) {}, maxConnections)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestSourceThrottler_LimitUnderShouldProcessRequest(t *testing.T) {
	t.Parallel()
	addr := "testAddress"

	maxConnections := uint32(1000)
	ws, _ := startNodeServerSourceThrottler(func(c *gin.Context) {}, maxConnections)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "127.0.0.1:8080"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSourceThrottler_LimitOverShouldErr(t *testing.T) {
	t.Parallel()

	maxRequests := 10
	maxConnections := uint32(maxRequests - 1)
	ws, _ := startNodeServerSourceThrottler(func(c *gin.Context) {}, maxConnections)

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

	maxRequests := 10
	maxConnections := uint32(maxRequests - 1)
	ws, resetHandler := startNodeServerSourceThrottler(func(c *gin.Context) {}, maxConnections)

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
