package middleware

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func startNodeServerResponseLogger(handler address.FacadeHandler, respLogMiddleware *responseLoggerMiddleware) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(respLogMiddleware.MiddlewareHandlerFunc())
	ginAddressRoutes := ws.Group("/address")
	if handler != nil {
		ginAddressRoutes.Use(WithFacade(handler))
	}
	addressRoutes, _ := wrapper.NewRouterWrapper("address", ginAddressRoutes, getRoutesConfig())
	address.Routes(addressRoutes)
	return ws
}

type responseLogFields struct {
	title    string
	path     string
	request  string
	duration time.Duration
	status   int
	response string
}

func TestNewResponseLoggerMiddleware(t *testing.T) {
	t.Parallel()

	rlm := NewResponseLoggerMiddleware(10)

	assert.False(t, check.IfNil(rlm))
}

func TestResponseLoggerMiddleware_DurationExceedsTimeout(t *testing.T) {
	t.Parallel()

	thresholdDuration := 10 * time.Millisecond
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			time.Sleep(thresholdDuration + 1*time.Millisecond)
			return big.NewInt(37777), nil
		},
	}

	rlf := responseLogFields{}
	printHandler := func(title string, path string, duration time.Duration, status int, request string, response string) {
		rlf.title = title
		rlf.path = path
		rlf.duration = duration
		rlf.status = status
		rlf.response = response
		rlf.request = request
	}

	rlm := NewResponseLoggerMiddleware(thresholdDuration)
	rlm.printRequestFunc = printHandler

	ws := startNodeServerResponseLogger(&facade, rlm)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, strings.Contains(rlf.title, prefixDurationTooLong))
	assert.True(t, rlf.duration > thresholdDuration)
	assert.Equal(t, http.StatusOK, rlf.status)
	assert.True(t, strings.Contains(rlf.response, "37777"))
}

func TestResponseLoggerMiddleware_BadRequest(t *testing.T) {
	t.Parallel()

	thresholdDuration := 10000 * time.Millisecond
	facade := mock.Facade{}

	rlf := responseLogFields{}
	printHandler := func(title string, path string, duration time.Duration, status int, request string, response string) {
		rlf.title = title
		rlf.path = path
		rlf.duration = duration
		rlf.status = status
		rlf.response = response
		rlf.request = request
	}

	rlm := NewResponseLoggerMiddleware(thresholdDuration)
	rlm.printRequestFunc = printHandler

	ws := startNodeServerResponseLogger(&facade, rlm)

	req, _ := http.NewRequest("GET", "/address//balance", nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(rlf.title, prefixBadRequest))
	assert.True(t, rlf.duration < thresholdDuration)
	assert.Equal(t, http.StatusBadRequest, rlf.status)
	assert.True(t, strings.Contains(rlf.response, "empty"))
}

func TestResponseLoggerMiddleware_InternalError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("internal err")
	thresholdDuration := 10000 * time.Millisecond
	facade := mock.Facade{
		BalanceHandler: func(s string) (*big.Int, error) {
			return nil, expectedErr
		},
	}

	rlf := responseLogFields{}
	printHandler := func(title string, path string, duration time.Duration, status int, request string, response string) {
		rlf.title = title
		rlf.path = path
		rlf.duration = duration
		rlf.status = status
		rlf.response = response
		rlf.request = request
	}

	rlm := NewResponseLoggerMiddleware(thresholdDuration)
	rlm.printRequestFunc = printHandler

	ws := startNodeServerResponseLogger(&facade, rlm)

	req, _ := http.NewRequest("GET", "/address/addr/balance", nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(rlf.title, prefixInternalError))
	assert.True(t, rlf.duration < thresholdDuration)
	assert.Equal(t, http.StatusInternalServerError, rlf.status)
	assert.True(t, strings.Contains(rlf.response, removeWhitespacesFromString(expectedErr.Error())))
}

func TestResponseLoggerMiddleware_ShouldNotCallHandler(t *testing.T) {
	t.Parallel()

	thresholdDuration := 10000 * time.Millisecond
	facade := mock.Facade{
		BalanceHandler: func(s string) (*big.Int, error) {
			return big.NewInt(5555), nil
		},
	}

	handlerWasCalled := false
	printHandler := func(title string, path string, duration time.Duration, status int, request string, response string) {
		handlerWasCalled = true
	}

	rlm := NewResponseLoggerMiddleware(thresholdDuration)
	rlm.printRequestFunc = printHandler

	ws := startNodeServerResponseLogger(&facade, rlm)

	req, _ := http.NewRequest("GET", "/address/addr/balance", nil)
	req.RemoteAddr = "bad address"
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.False(t, handlerWasCalled)
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"address": {
				Routes: []config.RouteConfig{
					{Name: "/:address", Open: true},
					{Name: "/:address/balance", Open: true},
				},
			},
		},
	}
}
