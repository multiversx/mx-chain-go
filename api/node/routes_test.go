package node_test

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/heartbeat"
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

type StatisticsResponse struct {
	GeneralResponse
	Statistics struct{
		LiveTPS float32 `json:"liveTPS"`
		PeakTPS float32 `json:"peakTPS"`
		NrOfShards uint32 `json:"nrOfShards"`
		BlockNumber uint64 `json:"blockNumber"`
		RoundTime uint32 `json:"roundTime"`
		AverageBlockTxCount float32 `json:"averageBlockTxCount"`
		LastBlockTxCount uint32 `json:"lastBlockTxCount"`
		TotalProcessedTxCount uint32 `json:"totalProcessedTxCount"`
	} `json:"statistics"`
}

func init() {
	gin.SetMode(gin.TestMode)
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

//------- Heartbeatstatus

func TestHeartbeatStatus_FailsWithoutFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(nil)
	defer func() {
		r := recover()

		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestHeartbeatstatus_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}
	facade.Running = true
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestHeartbeatstatus_FromFacadeErrors(t *testing.T) {
	t.Parallel()

	errExpected := errs.New("expected error")
	facade := mock.Facade{
		GetHeartbeatsHandler: func() ([]heartbeat.PubKeyHeartbeat, error) {
			return nil, errExpected
		},
	}
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, errExpected.Error(), statusRsp.Error)
}

func TestHeartbeatstatus(t *testing.T) {
	t.Parallel()

	hbStatus := []heartbeat.PubKeyHeartbeat{
		{
			HexPublicKey: "pk1",
			PeerHeartBeats: []heartbeat.PeerHeartbeat{
				{
					P2PAddress: "addr",
					IsActive:   true,
				},
			},
		},
	}
	facade := mock.Facade{
		GetHeartbeatsHandler: func() (heartbeats []heartbeat.PubKeyHeartbeat, e error) {
			return hbStatus, nil
		},
	}
	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := StatusResponse{}
	loadResponseAsString(resp.Body, &statusRsp)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotEqual(t, "", statusRsp.Message)
}

func TestStatistics_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/node/statistics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestStatistics_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/node/statistics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statisticsRsp := StatisticsResponse{}
	loadResponse(resp.Body, &statisticsRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statisticsRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestStatistics_ReturnsSuccessfully(t *testing.T) {
	nrOfShards := uint32(10)
	roundTime := uint64(4)
	benchmark, _ := statistics.NewTPSBenchmark(nrOfShards,roundTime)

	facade := mock.Facade{}
	facade.TpsBenchmarkHandler = func() *statistics.TpsBenchmark {
		return benchmark
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/node/statistics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statisticsRsp := StatisticsResponse{}
	loadResponse(resp.Body, &statisticsRsp)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, statisticsRsp.Statistics.NrOfShards, nrOfShards)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func loadResponseAsString(rsp io.Reader, response *StatusResponse) {
	buff, err := ioutil.ReadAll(rsp)
	if err != nil {
		logError(err)
		return
	}

	response.Message = string(buff)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler node.Handler) *gin.Engine {
	server := startNodeServerWithFacade(handler)
	return server
}

func startNodeServerWrongFacade() *gin.Engine {
	return startNodeServerWithFacade(mock.WrongFacade{})
}

func startNodeServerWithFacade(facade interface{}) *gin.Engine {
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
