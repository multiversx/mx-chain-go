package groups_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon"
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
		hg, err := groups.NewNodeGroup(&mock.FacadeStub{})
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

type epochStartResponse struct {
	Data struct {
		common.EpochStartDataAPI `json:"epochStart"`
	} `json:"data"`
	generalResponse
}

type managedKeysCountResponse struct {
	Data struct {
		Count int `json:"count"`
	} `json:"data"`
	generalResponse
}

type managedKeysResponse struct {
	Data struct {
		ManagedKeys []string `json:"managedKeys"`
	} `json:"data"`
	generalResponse
}

type loadedKeysResponse struct {
	Data struct {
		LoadedKeys []string `json:"loadedKeys"`
	} `json:"data"`
	generalResponse
}

type managedEligibleKeysResponse struct {
	Data struct {
		Keys []string `json:"eligibleKeys"`
	} `json:"data"`
	generalResponse
}

type managedWaitingKeysResponse struct {
	Data struct {
		Keys []string `json:"waitingKeys"`
	} `json:"data"`
	generalResponse
}

type waitingEpochsLeftResponse struct {
	Data struct {
		EpochsLeft uint32 `json:"epochsLeft"`
	} `json:"data"`
	generalResponse
}

func init() {
	gin.SetMode(gin.TestMode)
}

//------- Heartbeatstatus

func TestHeartbeatstatus_FromFacadeErrors(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	facade := mock.FacadeStub{
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

func TestHeartbeatStatus(t *testing.T) {
	t.Parallel()

	hbStatus := []data.PubKeyHeartbeat{
		{
			PublicKey:       "pk1",
			TimeStamp:       time.Now(),
			IsActive:        true,
			ReceivedShardID: uint32(0),
		},
	}
	facade := mock.FacadeStub{
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

func TestP2PMetrics_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				StatusP2pMetricsMapCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/p2pstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestNodeStatus_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				StatusMetricsMapWithoutP2PCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestBootstrapStatus_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				BootstrapMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/bootstrapstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestBootstrapStatusMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	statusMetricsProvider.SetUInt64Value(common.MetricTrieSyncNumReceivedBytes, uint64(100))
	statusMetricsProvider.SetUInt64Value(common.MetricTrieSyncNumProcessedNodes, uint64(150))

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/bootstrapstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keysFound := strings.Contains(respStr, common.MetricTrieSyncNumReceivedBytes) && strings.Contains(respStr, common.MetricTrieSyncNumProcessedNodes)
	assert.True(t, keysFound)
	valuesFound := strings.Contains(respStr, "100") && strings.Contains(respStr, "150")
	assert.True(t, valuesFound)
}

func TestNodeGroup_GetConnectedPeersRatings(t *testing.T) {
	t.Parallel()

	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetConnectedPeersRatingsOnMainNetworkCalled: func() (string, error) {
				return "", expectedErr
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/connected-peers-ratings", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedRatings := map[string]string{
			"pid1": "100",
			"pid2": "-50",
			"pid3": "-5",
		}
		buff, _ := json.Marshal(providedRatings)
		facade := mock.FacadeStub{
			GetConnectedPeersRatingsOnMainNetworkCalled: func() (string, error) {
				return string(buff), nil
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/connected-peers-ratings", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)
		respMap, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		ratings, ok := respMap["ratings"].(string)
		assert.True(t, ok)
		assert.Equal(t, string(buff), ratings)
	})
}

func TestStatusMetrics_ShouldDisplayNonP2pMetrics(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-details-key"
	val := "test-details-value"
	statusMetricsProvider.SetStringValue(key, val)

	p2pKey := "a_p2p_specific_key"
	statusMetricsProvider.SetStringValue(p2pKey, "p2p value")

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, val)
	assert.True(t, keyAndValueFoundInResponse)
	assert.False(t, strings.Contains(respStr, p2pKey))
}

func TestP2PStatusMetrics_ShouldDisplayNonP2pMetrics(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-details-key"
	val := "test-details-value"
	statusMetricsProvider.SetStringValue(key, val)

	p2pKey := "a_p2p_specific_key"
	p2pValue := "p2p value"
	statusMetricsProvider.SetStringValue(p2pKey, p2pValue)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/p2pstatus", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, p2pKey) && strings.Contains(respStr, p2pValue)
	assert.True(t, keyAndValueFoundInResponse)

	assert.False(t, strings.Contains(respStr, key))
}

func TestQueryDebug_ShouldBindJSONErrorsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetQueryHandlerCalled: func(name string) (handler debug.QueryHandler, err error) {
			return nil, nil
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("POST", "/node/debug", bytes.NewBuffer([]byte("invalid data")))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	queryResp := &generalResponse{}
	loadResponse(resp.Body, queryResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, queryResp.Error, apiErrors.ErrValidation.Error())
}

func TestQueryDebug_GetQueryErrorsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
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

	queryResp := &generalResponse{}
	loadResponse(resp.Body, queryResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, queryResp.Error, expectedErr.Error())
}

func TestQueryDebug_GetQueryShouldWork(t *testing.T) {
	t.Parallel()

	str1 := "aaa"
	str2 := "bbb"
	facade := mock.FacadeStub{
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

	queryResp := queryResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &queryResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, queryResp.Result, str1)
	assert.Contains(t, queryResp.Result, str2)
}

func TestPeerInfo_PeerInfoErrorsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
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
	facade := mock.FacadeStub{
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

func TestEpochStartData_InvalidEpochShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetEpochStartDataAPICalled: func(epoch uint32) (*common.EpochStartDataAPI, error) {
			return nil, nil
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/epoch-start/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrValidation.Error()))
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrBadUrlParams.Error()))
}

func TestEpochStartData_FacadeErrorsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetEpochStartDataAPICalled: func(epoch uint32) (*common.EpochStartDataAPI, error) {
			return nil, expectedErr
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/epoch-start/4", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestEpochStartData_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedEpochStartData := &common.EpochStartDataAPI{
		Nonce:     1,
		Round:     2,
		Shard:     3,
		Timestamp: 4,
	}

	facade := mock.FacadeStub{
		GetEpochStartDataAPICalled: func(epoch uint32) (*common.EpochStartDataAPI, error) {
			return expectedEpochStartData, nil
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/epoch-start/3", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &epochStartResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)

	require.Equal(t, *expectedEpochStartData, response.Data.EpochStartDataAPI)
}

func TestPrometheusMetrics_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				StatusMetricsWithoutP2PPrometheusStringCalled: func() (string, error) {
					return "", expectedErr
				},
			}
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/metrics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestPrometheusMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := "test-key"
	val := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, val)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/metrics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", val))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestNodeGroup_ManagedKeysCount(t *testing.T) {
	t.Parallel()

	providedCount := 1000
	facade := mock.FacadeStub{
		GetManagedKeysCountCalled: func() int {
			return providedCount
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/managed-keys/count", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &managedKeysCountResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, providedCount, response.Data.Count)
}

func TestNodeGroup_ManagedKeys(t *testing.T) {
	t.Parallel()

	providedKeys := []string{
		"pk1",
		"pk2",
	}
	facade := mock.FacadeStub{
		GetManagedKeysCalled: func() []string {
			return providedKeys
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/managed-keys", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &managedKeysResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, providedKeys, response.Data.ManagedKeys)
}

func TestNodeGroup_LoadedKeys(t *testing.T) {
	t.Parallel()

	providedKeys := []string{
		"pk1",
		"pk2",
	}
	facade := mock.FacadeStub{
		GetLoadedKeysCalled: func() []string {
			return providedKeys
		},
	}

	nodeGroup, err := groups.NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/loaded-keys", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &loadedKeysResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, providedKeys, response.Data.LoadedKeys)
}

func TestNodeGroup_ManagedKeysEligible(t *testing.T) {
	t.Parallel()

	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetEligibleManagedKeysCalled: func() ([]string, error) {
				return nil, expectedErr
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/managed-keys/eligible", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedKeys := []string{
			"key1",
			"key2",
			"key3",
		}
		facade := mock.FacadeStub{
			GetEligibleManagedKeysCalled: func() ([]string, error) {
				return providedKeys, nil
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/managed-keys/eligible", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &managedEligibleKeysResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", response.Error)
		assert.Equal(t, providedKeys, response.Data.Keys)
	})
}

func TestNodeGroup_ManagedKeysWaiting(t *testing.T) {
	t.Parallel()

	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetWaitingManagedKeysCalled: func() ([]string, error) {
				return nil, expectedErr
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/managed-keys/waiting", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedKeys := []string{
			"key1",
			"key2",
			"key3",
		}
		facade := mock.FacadeStub{
			GetWaitingManagedKeysCalled: func() ([]string, error) {
				return providedKeys, nil
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/managed-keys/waiting", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &managedWaitingKeysResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", response.Error)
		assert.Equal(t, providedKeys, response.Data.Keys)
	})
}

func TestNodeGroup_WaitingEpochsLeft(t *testing.T) {
	t.Parallel()

	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetWaitingEpochsLeftForPublicKeyCalled: func(publicKey string) (uint32, error) {
				return 0, expectedErr
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/waiting-epochs-left/key", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEpochsLeft := uint32(10)
		facade := mock.FacadeStub{
			GetWaitingEpochsLeftForPublicKeyCalled: func(publicKey string) (uint32, error) {
				return providedEpochsLeft, nil
			},
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/waiting-epochs-left/key", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &waitingEpochsLeftResponse{}
		loadResponse(resp.Body, response)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", response.Error)
		assert.Equal(t, providedEpochsLeft, response.Data.EpochsLeft)
	})
}

func TestNodeGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		nodeGroup, err := groups.NewNodeGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = nodeGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		nodeGroup, err := groups.NewNodeGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = nodeGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		statusMetricsProvider := statusHandler.NewStatusMetrics()
		key := "test-key"
		val := uint64(37)
		statusMetricsProvider.SetUInt64Value(key, val)

		facade := mock.FacadeStub{}
		facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
			return statusMetricsProvider
		}

		nodeGroup, err := groups.NewNodeGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(nodeGroup, "node", getNodeRoutesConfig())

		req, _ := http.NewRequest("GET", "/node/metrics", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		respBytes, _ := io.ReadAll(resp.Body)
		respStr := string(respBytes)
		assert.Equal(t, resp.Code, http.StatusOK)
		keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", val))
		assert.True(t, keyAndValueFoundInResponse)

		newFacade := mock.FacadeStub{
			StatusMetricsHandler: func() external.StatusMetricsHandler {
				return &testscommon.StatusMetricsStub{
					StatusMetricsWithoutP2PPrometheusStringCalled: func() (string, error) {
						return "", expectedErr
					},
				}
			},
		}

		err = nodeGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/node/metrics", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, response)
		assert.Equal(t, expectedErr.Error(), response.Error)
	})
}

func TestNodeGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	nodeGroup, _ := groups.NewNodeGroup(nil)
	require.True(t, nodeGroup.IsInterfaceNil())

	nodeGroup, _ = groups.NewNodeGroup(&mock.FacadeStub{})
	require.False(t, nodeGroup.IsInterfaceNil())
}

func loadResponseAsString(rsp io.Reader, response *statusResponse) {
	buff, err := io.ReadAll(rsp)
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
					{Name: "/epoch-start/:epoch", Open: true},
					{Name: "/bootstrapstatus", Open: true},
					{Name: "/connected-peers-ratings", Open: true},
					{Name: "/managed-keys/count", Open: true},
					{Name: "/managed-keys", Open: true},
					{Name: "/loaded-keys", Open: true},
					{Name: "/managed-keys/eligible", Open: true},
					{Name: "/managed-keys/waiting", Open: true},
					{Name: "/waiting-epochs-left/:key", Open: true},
				},
			},
		},
	}
}
