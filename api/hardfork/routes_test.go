package hardfork_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go-logger"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("api/hardfork_test")

func init() {
	gin.SetMode(gin.TestMode)
}

type TriggerResponse struct {
	Status string `json:"status"`
}

func startNodeServer(handler hardfork.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ginHardforkRoute := ws.Group("/hardfork")
	if handler != nil {
		ginHardforkRoute.Use(middleware.WithFacade(handler))
	}
	hardForkRoute, _ := wrapper.NewRouterWrapper("hardfork", ginHardforkRoute, getRoutesConfig())
	hardfork.Routes(hardForkRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("facade", mock.WrongFacade{})
	})
	ginHardforkRoute := ws.Group("/hardfork")
	hardForkRoute, _ := wrapper.NewRouterWrapper("hardfork", ginHardforkRoute, getRoutesConfig())
	hardfork.Routes(hardForkRoute)
	return ws
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	log.LogIfError(err)
}

func TestTrigger_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestTrigger_WithWrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, response.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestTrigger_TriggerCanNotExecuteShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func(epoch uint32, forced bool) error {
			return expectedErr
		},
	})

	hr := &hardfork.HarforkRequest{
		Epoch: 4,
	}

	buffHr, _ := json.Marshal(hr)
	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Contains(t, response.Error, expectedErr.Error())
}

func TestTrigger_TriggerWrongRequestTypeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(&mock.HardforkFacade{})

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer([]byte("wrong buffer")))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusBadRequest)
}

func TestTrigger_ManualShouldWork(t *testing.T) {
	t.Parallel()

	recoveredEpoch := uint32(0)
	hr := &hardfork.HarforkRequest{
		Epoch: 4,
	}
	buffHr, _ := json.Marshal(hr)
	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func(epoch uint32, forced bool) error {
			atomic.StoreUint32(&recoveredEpoch, epoch)

			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return false
		},
	})

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	triggerResponse := TriggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, hardfork.ExecManualTrigger, triggerResponse.Status)
	assert.Equal(t, hr.Epoch, atomic.LoadUint32(&recoveredEpoch))
}

func TestTrigger_BroadcastShouldWork(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func(_ uint32, _ bool) error {
			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return true
		},
	})

	hr := &hardfork.HarforkRequest{
		Epoch: 4,
	}
	buffHr, _ := json.Marshal(hr)
	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	triggerResponse := TriggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, hardfork.ExecBroadcastTrigger, triggerResponse.Status)
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"hardfork": {
				[]config.RouteConfig{
					{Name: "/trigger", Open: true},
				},
			},
		},
	}
}
