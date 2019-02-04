package node_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/node/mock"
)

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type StatusResponse struct {
	GeneralResponse
	Running bool `json:"running"`
}

type AddressResponse struct {
	GeneralResponse
	Address string `json:"address"`
}

func TestStatus_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestStatus_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestStatus_ReturnsCorrectResponseOnStart(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.True(t, statusRsp.Running)
}

func TestStatus_ReturnsCorrectResponseOnStop(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	facade.Running = false
	req2, _ := http.NewRequest("GET", "/node/status", nil)
	resp2 := httptest.NewRecorder()
	ws.ServeHTTP(resp2, req2)

	statusRsp2 := StatusResponse{}
	loadResponse(resp2.Body, &statusRsp2)
	assert.False(t, statusRsp2.Running)
}

func TestStartNode_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/start", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestStartNode_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/start", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestStartNode_AlreadyRunning(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/start", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, statusRsp.Message, errors.ErrNodeAlreadyRunning.Error())
}

func TestStartNode_FromFacadeErrors(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.ShouldErrorStart = true
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/start", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, fmt.Sprintf("%s: error", errors.ErrBadInitOfNode.Error()))
}

func TestStartNode(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/start", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, statusRsp.Message, "ok")
}

func TestStopNode_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestStopNode_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestStopNode_AlreadyStopped(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = false
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, statusRsp.Message, errors.ErrNodeAlreadyStopped.Error())
}

func TestStopNode_FromFacadeErrors(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	facade.ShouldErrorStop = true
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, fmt.Sprintf("%s: error", errors.ErrCouldNotStopNode.Error()))
}

func TestStopNode(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	facade.ShouldErrorStop = false
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, statusRsp.Message, "ok")
}

func TestAddress_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestAddress_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressRsp := AddressResponse{}
	loadResponse(resp.Body, &addressRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, addressRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestAddress_FailsWithInvalidUrlString(t *testing.T) {
	facade := mock.Facade{}
	facade.GetCurrentPublicKeyHandler = func() string {
		// we return a malformed scheme so that url.Parse will error
		return "cache_object:foo/bar"
	}
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressRsp := AddressResponse{}
	loadResponse(resp.Body, &addressRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, addressRsp.Error, errors.ErrCouldNotParsePubKey.Error())
}

func TestAddress_ReturnsSuccessfully(t *testing.T) {
	facade := mock.Facade{}
	address := "abcdefghijklmnopqrstuvwxyz"
	facade.GetCurrentPublicKeyHandler = func() string {
		// we return a malformed scheme so that url.Parse will error
		return address
	}
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressRsp := AddressResponse{}
	loadResponse(resp.Body, &addressRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, addressRsp.Address, address)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler node.Handler) *gin.Engine {
	return startNodeServerWithFacade(handler)
}

func startNodeServerWrongFacade() *gin.Engine {
	return startNodeServerWithFacade(mock.WrongFacade{})
}

func startNodeServerWithFacade(facade interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	if facade != nil {
		ws.Use(func(c *gin.Context) {
			c.Set("elrondFacade", facade)
		})
	}

	nodeRoutes := ws.Group("/node")
	node.Routes(nodeRoutes)
	return ws
}
