package middleware_test

import (
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type reseter interface {
	Reset()
}

func startNodeServerMiddleware(
	handler address.FacadeHandler,
	maxConcurrentConnections uint32,
	maxNumRequestsPerAddress uint32,
) (*gin.Engine, reseter) {

	ws := gin.New()
	ws.Use(cors.Default())
	mw, _ := middleware.NewMiddleware(handler, maxConcurrentConnections, maxNumRequestsPerAddress)
	ws.Use(mw.MiddlewareHandlerFunc())
	ginAddressRoutes := ws.Group("/address")
	addressRoutes, _ := wrapper.NewRouterWrapper("address", ginAddressRoutes, getRoutesConfig())
	address.Routes(addressRoutes)
	return ws, mw
}

func makeRequestOnMiddleware(ws *gin.Engine, mutResponses *sync.Mutex, responses map[int]int) {
	addr := "testAddress"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "127.0.0.1:8080"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	mutResponses.Lock()
	responses[resp.Code]++
	mutResponses.Unlock()
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"address": {
				[]config.RouteConfig{
					{Name: "/:address", Open: true},
					{Name: "/:address/balance", Open: true},
				},
			},
		},
	}
}

func TestNewMiddleware_NilFacadeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mw, err := middleware.NewMiddleware(nil, 1, 1)

	assert.True(t, check.IfNil(mw))
	assert.Equal(t, middleware.ErrNilFacade, err)
}

func TestNewMiddleware_InvalidMaxConcurrentConnectionsShouldErr(t *testing.T) {
	t.Parallel()

	mw, err := middleware.NewMiddleware(&mock.Facade{}, 0, 1)

	assert.True(t, check.IfNil(mw))
	assert.Equal(t, middleware.ErrInvalidMaxNumConcurrentRequests, err)
}

func TestNewMiddleware_InvalidMaxNumRequestsPerAddressShouldErr(t *testing.T) {
	t.Parallel()

	mw, err := middleware.NewMiddleware(&mock.Facade{}, 1, 0)

	assert.True(t, check.IfNil(mw))
	assert.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
}

func TestNewMiddleware_ShouldWork(t *testing.T) {
	t.Parallel()

	mw, err := middleware.NewMiddleware(&mock.Facade{}, 1, 1)

	assert.False(t, check.IfNil(mw))
	assert.Nil(t, err)
}

// Limit testing

func TestNewMiddleware_LimitBadRequestShouldErr(t *testing.T) {
	t.Parallel()
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := uint32(1000)
	maxConcurrentConnections := uint32(1)
	ws, _ := startNodeServerMiddleware(&facade, maxConcurrentConnections, maxRequests)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestNewMiddleware_LimitPerAddressIsUnderThresholdShouldProcessRequest(t *testing.T) {
	t.Parallel()
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := uint32(1000)
	maxConcurrentConnections := uint32(1)
	ws, _ := startNodeServerMiddleware(&facade, maxConcurrentConnections, maxRequests)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "127.0.0.1:8080"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestNewMiddleware_LimitPerAddressIsOverThresholdShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := uint32(1000)
	maxConcurrentConnections := uint32(1)
	executedRequests := int(maxRequests) + 1 //one more to reach the limit
	ws, _ := startNodeServerMiddleware(&facade, maxConcurrentConnections, maxRequests)

	mutResponses := sync.Mutex{}
	responses := make(map[int]int)
	for i := 0; i < executedRequests; i++ {
		makeRequestOnMiddleware(ws, &mutResponses, responses)
	}

	mutResponses.Lock()
	assert.Equal(t, int(maxRequests), responses[http.StatusOK])
	assert.Equal(t, 1, responses[http.StatusTooManyRequests])
	mutResponses.Unlock()
}

func TestNewMiddleware_LimitResetShouldWork(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(10), nil
		},
	}

	maxRequests := uint32(1000)
	maxConcurrentConnections := uint32(1)
	executedRequests := int(maxRequests) + 1 //one more to reach the limit
	ws, mw := startNodeServerMiddleware(&facade, maxConcurrentConnections, maxRequests)

	mutResponses := sync.Mutex{}
	responses := make(map[int]int)
	numBatches := 2
	for j := 0; j < numBatches; j++ {
		for i := 0; i < executedRequests; i++ {
			makeRequestOnMiddleware(ws, &mutResponses, responses)
		}

		mw.Reset()
	}

	mutResponses.Lock()
	assert.Equal(t, numBatches*int(maxRequests), responses[http.StatusOK])
	assert.Equal(t, numBatches, responses[http.StatusTooManyRequests])
	mutResponses.Unlock()
}

func TestNewMiddleware_LimitOfConcurrentRequestsOverShouldError(t *testing.T) {
	t.Parallel()

	numCalls := uint32(0)
	responseDelay := time.Second
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			time.Sleep(responseDelay)
			atomic.AddUint32(&numCalls, 1)

			return big.NewInt(10), nil
		},
	}

	maxRequests := uint32(1000)
	maxConcurrentConnections := uint32(1)
	ws, _ := startNodeServerMiddleware(&facade, maxConcurrentConnections, maxRequests)

	mutResponses := sync.Mutex{}
	responses := make(map[int]int)
	numRequests := 10
	numBatches := 2

	for j := 0; j < numBatches; j++ {
		fmt.Printf("Starting batch requests %d, making %d simultaneous requests...\n", j, numRequests)

		for i := 0; i < numRequests; i++ {
			go makeRequestOnMiddleware(ws, &mutResponses, responses)
		}

		time.Sleep(responseDelay + time.Second)
	}

	assert.Equal(t, uint32(numBatches), atomic.LoadUint32(&numCalls))
	mutResponses.Lock()
	assert.Equal(t, numBatches, responses[http.StatusOK])
	assert.Equal(t, numBatches*(numRequests-1), responses[http.StatusTooManyRequests])
	mutResponses.Unlock()
}
