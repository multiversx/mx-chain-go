package hardfork_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-logger"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("api/hardfork_test")

func init() {
	gin.SetMode(gin.TestMode)
}

type generalResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type TriggerResponse struct {
	generalResponse
	Status string `json:"status"`
}

func startNodeServer(handler hardfork.TriggerHardforkHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	hardforkRoute := ws.Group("/hardfork")
	if handler != nil {
		hardforkRoute.Use(middleware.WithElrondFacade(handler))
	}
	hardfork.Routes(hardforkRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	hardforkRoute := ws.Group("/hardfork")
	hardfork.Routes(hardforkRoute)
	return ws
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	log.LogIfError(err)
}

func TestTrigger_WithWrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, triggerResponse.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestTrigger_TriggerCanNotExecuteShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func() error {
			return expectedErr
		},
	})

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Contains(t, triggerResponse.Error, expectedErr.Error())
}

func TestTrigger_ManualShouldWork(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func() error {
			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return false
		},
	})

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, hardfork.ExecManualTrigger, triggerResponse.Status)
}

func TestTrigger_BroadcastShouldWork(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(&mock.HardforkFacade{
		TriggerCalled: func() error {
			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return true
		},
	})

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(nil))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, hardfork.ExecBroadcastTrigger, triggerResponse.Status)
}
