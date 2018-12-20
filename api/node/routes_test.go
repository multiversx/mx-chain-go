package node_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
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

func TestStatusFailsWithoutFacade(t *testing.T) {
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

func TestStatusFailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Message, "Invalid app context")
}

func TestStatusReturnsCorrectResponseOnStart(t *testing.T) {
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

func TestStatusReturnsCorrectResponseOnStop(t *testing.T) {
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

func TestStartNodeFailsWithoutFacade(t *testing.T) {
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

func TestStartNodeFailsWithWrongFacadeTypeConversion(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Invalid app context")
}

func TestStartNodeAlreadyRunning(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Node already running")
}

func TestStartNodeFromFacadeErrors(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Bad init of node: error")
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

func TestStopNodeFailsWithoutFacade(t *testing.T) {
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

func TestStopNodeFailsWithWrongFacadeTypeConversion(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Invalid app context")
}

func TestStopNodeAlreadyStopped(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Node already stopped")
}

func TestStopNodeFromFacadeErrors(t *testing.T) {
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
	assert.Equal(t, statusRsp.Message, "Could not stop node: error")
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
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	nodeRoutes := ws.Group("/node")
	if handler != nil {
		nodeRoutes.Use(middleware.WithElrondFacade(handler))
	}
	node.Routes(nodeRoutes)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	nodeRoutes := ws.Group("/node")
	node.Routes(nodeRoutes)
	return ws
}
