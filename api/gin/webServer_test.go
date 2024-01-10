package gin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/testscommon/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsNewWebServer() ArgsNewWebServer {
	return ArgsNewWebServer{
		Facade: &mock.FacadeStub{
			PprofEnabledCalled: func() bool {
				return true // coverage
			},
		},
		ApiConfig: config.ApiRoutesConfig{
			Logging: config.ApiLoggingConfig{
				LoggingEnabled:          true,
				ThresholdInMicroSeconds: 10,
			},
			APIPackages: map[string]config.APIPackageConfig{
				"log": {Routes: []config.RouteConfig{
					{
						Name: "/log",
						Open: true,
					},
				}},
			},
		},
		AntiFloodConfig: config.WebServerAntifloodConfig{
			WebServerAntifloodEnabled:    true,
			SimultaneousRequests:         1,
			SameSourceRequests:           1,
			SameSourceResetIntervalInSec: 1,
		},
	}
}

func TestNewGinWebServerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsNewWebServer()
		args.Facade = nil

		ws, err := NewGinWebServerHandler(args)
		require.True(t, errors.Is(err, apiErrors.ErrCannotCreateGinWebServer))
		require.True(t, strings.Contains(err.Error(), apiErrors.ErrNilFacadeHandler.Error()))
		require.Nil(t, ws)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ws, err := NewGinWebServerHandler(createMockArgsNewWebServer())
		require.Nil(t, err)
		require.NotNil(t, ws)
	})
}

func TestWebServer_StartHttpServer(t *testing.T) {
	t.Run("RestApiInterface returns DefaultRestPortOff", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.Facade = &mock.FacadeStub{
			RestApiInterfaceCalled: func() string {
				return facade.DefaultRestPortOff
			},
		}

		ws, _ := NewGinWebServerHandler(args)
		require.NotNil(t, ws)

		err := ws.StartHttpServer()
		require.Nil(t, err)
	})
	t.Run("createMiddlewareLimiters returns error due to middleware.NewSourceThrottler error", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.AntiFloodConfig.SameSourceRequests = 0
		ws, _ := NewGinWebServerHandler(args)
		require.NotNil(t, ws)

		err := ws.StartHttpServer()
		require.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
	})
	t.Run("createMiddlewareLimiters returns error due to middleware.NewGlobalThrottler error", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.AntiFloodConfig.SimultaneousRequests = 0
		ws, _ := NewGinWebServerHandler(args)
		require.NotNil(t, ws)

		err := ws.StartHttpServer()
		require.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
	})
	t.Run("should work", func(t *testing.T) {
		ws, _ := NewGinWebServerHandler(createMockArgsNewWebServer())
		require.NotNil(t, ws)

		err := ws.StartHttpServer()
		require.Nil(t, err)

		time.Sleep(2 * time.Second)

		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/log", nil)
		require.Nil(t, err)

		req.Header.Set("Sec-Websocket-Version", "13")
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-Websocket-Key", "key")

		resp, err := client.Do(req)
		require.Nil(t, err)

		err = resp.Body.Close()
		require.Nil(t, err)

		time.Sleep(2 * time.Second)
		err = ws.Close()
		require.Nil(t, err)
	})
}

func TestWebServer_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("update with nil facade should error", func(t *testing.T) {
		t.Parallel()

		ws, _ := NewGinWebServerHandler(createMockArgsNewWebServer())
		require.NotNil(t, ws)

		err := ws.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work - one of the groupHandlers returns err", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsNewWebServer()
		args.Facade = &mock.FacadeStub{
			RestApiInterfaceCalled: func() string {
				return "provided interface"
			},
		}

		ws, _ := NewGinWebServerHandler(args)
		require.NotNil(t, ws)

		ws.groups = make(map[string]shared.GroupHandler)
		ws.groups["first"] = &api.GroupHandlerStub{
			UpdateFacadeCalled: func(newFacade interface{}) error {
				return errors.New("error")
			},
		}
		ws.groups["second"] = &api.GroupHandlerStub{
			UpdateFacadeCalled: func(newFacade interface{}) error {
				return nil
			},
		}

		err := ws.UpdateFacade(&mock.FacadeStub{})
		require.Nil(t, err)
	})
}

func TestWebServer_CloseWithDisabledServerShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	args := createMockArgsNewWebServer()
	args.Facade = &mock.FacadeStub{
		RestApiInterfaceCalled: func() string {
			return facade.DefaultRestPortOff
		},
	}

	ws, _ := NewGinWebServerHandler(args)
	require.NotNil(t, ws)

	err := ws.StartHttpServer()
	require.Nil(t, err)

	err = ws.Close()
	assert.Nil(t, err)
}
