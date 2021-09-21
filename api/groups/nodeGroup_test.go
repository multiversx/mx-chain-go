package groups_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNodeGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewNodeGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewNodeGroup(&mock.Facade{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type generalResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type statusResponse struct {
	generalResponse
	Running bool `json:"running"`
}

type queryResponse struct {
	generalResponse
	Result []string `json:"result"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

//------- Heartbeatstatus

func TestHeartbeatstatus_FromFacadeErrors(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	facade := mock.Facade{
		GetHeartbeatsHandler: func() ([]data.PubKeyHeartbeat, error) {
			return nil, errExpected
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := statusResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, errExpected.Error(), statusRsp.Error)
}

func TestHeartbeatstatus(t *testing.T) {
	t.Parallel()

	hbStatus := []data.PubKeyHeartbeat{
		{
			PublicKey:       "pk1",
			TimeStamp:       time.Now(),
			MaxInactiveTime: data.Duration{Duration: 0},
			IsActive:        true,
			ReceivedShardID: uint32(0),
		},
	}
	facade := mock.Facade{
		GetHeartbeatsHandler: func() (heartbeats []data.PubKeyHeartbeat, e error) {
			return hbStatus, nil
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/heartbeatstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := statusResponse{}
	loadResponseAsString(resp.Body, &statusRsp)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotEqual(t, "", statusRsp.Message)
}

func TestStatusMetrics_ShouldDisplayNonP2pMetrics(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-details-key"
	value := "test-details-value"
	statusMetricsProvider.SetStringValue(key, value)

	p2pKey := "a_p2p_specific_key"
	statusMetricsProvider.SetStringValue(p2pKey, "p2p value")
	numCheckpoints := 2

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}
	facade.GetNumCheckpointsFromPeerStateCalled = func() uint32 {
		return uint32(numCheckpoints)
	}
	facade.GetNumCheckpointsFromAccountStateCalled = func() uint32 {
		return uint32(numCheckpoints)
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, value)
	assert.True(t, keyAndValueFoundInResponse)
	assert.False(t, strings.Contains(respStr, p2pKey))

	keyAndValueFoundInResponse = strings.Contains(respStr, groups.AccStateCheckpointsKey) && strings.Contains(respStr, strconv.Itoa(numCheckpoints))
	assert.True(t, keyAndValueFoundInResponse)

	keyAndValueFoundInResponse = strings.Contains(respStr, groups.PeerStateCheckpointsKey) && strings.Contains(respStr, strconv.Itoa(numCheckpoints))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestP2PStatusMetrics_ShouldDisplayNonP2pMetrics(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-details-key"
	value := "test-details-value"
	statusMetricsProvider.SetStringValue(key, value)

	p2pKey := "a_p2p_specific_key"
	p2pValue := "p2p value"
	statusMetricsProvider.SetStringValue(p2pKey, p2pValue)

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/p2pstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, p2pKey) && strings.Contains(respStr, p2pValue)
	assert.True(t, keyAndValueFoundInResponse)

	assert.False(t, strings.Contains(respStr, key))
}

func TestQueryDebug_GetQueryErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetQueryHandlerCalled: func(name string) (handler debug.QueryHandler, err error) {
			return nil, expectedErr
		},
	}

	qdr := &groups.QueryDebugRequest{}
	jsonStr, _ := json.Marshal(qdr)

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("POST", "/node/debug", bytes.NewBuffer(jsonStr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	queryResponse := &generalResponse{}
	loadResponse(resp.Body, queryResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, queryResponse.Error, expectedErr.Error())
}

func TestQueryDebug_GetQueryShouldWork(t *testing.T) {
	t.Parallel()

	str1 := "aaa"
	str2 := "bbb"
	facade := mock.Facade{
		GetQueryHandlerCalled: func(name string) (handler debug.QueryHandler, err error) {
			return &mock.QueryHandlerStub{
					QueryCalled: func(search string) []string {
						return []string{str1, str2}
					},
				},
				nil
		},
	}

	qdr := &groups.QueryDebugRequest{}
	jsonStr, _ := json.Marshal(qdr)

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("POST", "/node/debug", bytes.NewBuffer(jsonStr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	queryResponse := queryResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &queryResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, queryResponse.Result, str1)
	assert.Contains(t, queryResponse.Result, str2)
}

func TestPeerInfo_PeerInfoErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetPeerInfoCalled: func(pid string) ([]core.QueryP2PPeerInfo, error) {
			return nil, expectedErr
		},
	}
	pid := "pid1"

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/peerinfo?pid="+pid, nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestPeerInfo_PeerInfoShouldWork(t *testing.T) {
	t.Parallel()

	pidProvided := "16Uiu2HAmRCVXdXqt8BXfhrzotczHMXXvgHPd7iwGWvS53JT1xdw6"
	val := core.QueryP2PPeerInfo{
		Pid: pidProvided,
	}
	facade := mock.Facade{
		GetPeerInfoCalled: func(pid string) ([]core.QueryP2PPeerInfo, error) {
			if pid == pidProvided {
				return []core.QueryP2PPeerInfo{val}, nil
			}

			assert.Fail(t, "should have received the pid")
			return nil, nil
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/peerinfo?pid="+pidProvided, nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)

	responseInfo, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	assert.NotNil(t, responseInfo["info"])
}

func TestPrometheusMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-key"
	value := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, value)

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/metrics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", value))
	assert.True(t, keyAndValueFoundInResponse)
}

func loadResponseAsString(rsp io.Reader, response *statusResponse) {
	buff, err := ioutil.ReadAll(rsp)
	if err != nil {
		logError(err)
		return
	}

	response.Message = string(buff)
}

func getNodeRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"node": {
				Routes: []config.RouteConfig{
					{Name: "/status", Open: true},
					{Name: "/metrics", Open: true},
					{Name: "/heartbeatstatus", Open: true},
					{Name: "/p2pstatus", Open: true},
					{Name: "/debug", Open: true},
					{Name: "/peerinfo", Open: true},
				},
			},
		},
	}
}

