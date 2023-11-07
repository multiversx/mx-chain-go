package groups_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/api"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewNetworkGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewNetworkGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type esdtTokensResponseData struct {
	Tokens []string `json:"tokens"`
}

type esdtTokensResponse struct {
	Data  esdtTokensResponseData `json:"data"`
	Error string                 `json:"error"`
	Code  string                 `json:"code"`
}

type genesisNodesConfigResponse struct {
	Data  genesisNodesConfigData `json:"data"`
	Error string                 `json:"error"`
	Code  string                 `json:"code"`
}

type genesisNodesConfigData struct {
	Nodes groups.GenesisNodesConfig `json:"nodes"`
}

type genesisBalancesResponse struct {
	Data  genesisBalancesResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

type genesisBalancesResponseData struct {
	Balances []*common.InitialAccountAPI `json:"balances"`
}

type ratingsConfigResponse struct {
	Data struct {
		Config map[string]interface{} `json:"config"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type gasConfigsResponse struct {
	Data  gasConfigsData `json:"data"`
	Error string         `json:"error"`
	Code  string         `json:"code"`
}

type gasConfigsData struct {
	Configs groups.GasConfig `json:"gasConfigs"`
}

func TestNetworkConfigMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricMinGasLimit
	val := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, val)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", val))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestGetNetworkConfig_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				ConfigMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestGetNetworkStatus_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				NetworkMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestNetworkConfigMetrics_GasLimitGuardedTxShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricExtraGasLimitGuardedTx
	val := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, val)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", val))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestNetworkStatusMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricEpochNumber
	val := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, val)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", val))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestEconomicsMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricTotalSupply
	val := "12345"
	statusMetricsProvider.SetStringValue(key, val)

	facade := mock.FacadeStub{
		GetTotalStakedValueHandler: func() (*api.StakeValues, error) {
			return &api.StakeValues{
				BaseStaked: big.NewInt(100),
				TopUp:      big.NewInt(20),
			}, nil
		},
	}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/economics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, val)
	assert.True(t, keyAndValueFoundInResponse)
}

func TestGetEconomicValues_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				EconomicsMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
		GetTotalStakedValueHandler: func() (*api.StakeValues, error) {
			return nil, nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/economics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestEconomicsMetrics_CannotGetStakeValues(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricTotalSupply
	val := "12345"
	statusMetricsProvider.SetStringValue(key, val)

	localErr := fmt.Errorf("%s", "local error")
	facade := mock.FacadeStub{
		GetTotalStakedValueHandler: func() (*api.StakeValues, error) {
			return nil, localErr
		},
	}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/economics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestGetAllIssuedESDTs_ShouldWork(t *testing.T) {
	tokens := []string{"tokenA", "tokenB"}
	facade := mock.FacadeStub{
		GetAllIssuedESDTsCalled: func(_ string) ([]string, error) {
			return tokens, nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdts", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := esdtTokensResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, resp.Code, http.StatusOK)

	assert.Equal(t, tokens, response.Data.Tokens)
}

func TestGetAllIssuedESDTs_Error(t *testing.T) {
	localErr := fmt.Errorf("%s", "local error")
	facade := mock.FacadeStub{
		GetAllIssuedESDTsCalled: func(_ string) ([]string, error) {
			return nil, localErr
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdts", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestDirectStakedInfo_ShouldWork(t *testing.T) {
	stakedData1 := api.DirectStakedValue{
		Address:    "addr1",
		BaseStaked: "1111",
		TopUp:      "2222",
		Total:      "3333",
	}
	stakedData2 := api.DirectStakedValue{
		Address:    "addr2",
		BaseStaked: "4444",
		TopUp:      "5555",
		Total:      "9999",
	}

	facade := mock.FacadeStub{
		GetDirectStakedListHandler: func() ([]*api.DirectStakedValue, error) {
			return []*api.DirectStakedValue{&stakedData1, &stakedData2}, nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/direct-staked-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	valuesFoundInResponse := directStakedFoundInResponse(respStr, stakedData1) && directStakedFoundInResponse(respStr, stakedData2)
	assert.True(t, valuesFoundInResponse)
}

func directStakedFoundInResponse(response string, stakedData api.DirectStakedValue) bool {
	return strings.Contains(response, stakedData.Address) && strings.Contains(response, stakedData.Total) &&
		strings.Contains(response, stakedData.BaseStaked) && strings.Contains(response, stakedData.TopUp)
}

func TestDirectStakedInfo_CannotGetDirectStakedList(t *testing.T) {
	expectedError := fmt.Errorf("%s", "expected error")
	facade := mock.FacadeStub{
		GetDirectStakedListHandler: func() ([]*api.DirectStakedValue, error) {
			return nil, expectedError
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/direct-staked-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(respStr, expectedError.Error()))
}

func TestDelegatedInfo_ShouldWork(t *testing.T) {
	delegator1 := api.Delegator{
		DelegatorAddress: "addr1",
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: "addr2",
				Value:               "1111",
			},
			{
				DelegationScAddress: "addr3",
				Value:               "2222",
			},
		},
		Total:         "3333",
		TotalAsBigInt: big.NewInt(4444),
	}

	delegator2 := api.Delegator{
		DelegatorAddress: "addr4",
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: "addr5",
				Value:               "5555",
			},
			{
				DelegationScAddress: "addr6",
				Value:               "6666",
			},
		},
		Total:         "12221",
		TotalAsBigInt: big.NewInt(13331),
	}

	facade := mock.FacadeStub{
		GetDelegatorsListHandler: func() ([]*api.Delegator, error) {
			return []*api.Delegator{&delegator1, &delegator2}, nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/delegated-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	valuesFoundInResponse := delegatorFoundInResponse(respStr, delegator1) && delegatorFoundInResponse(respStr, delegator2)
	assert.True(t, valuesFoundInResponse)
}

func delegatorFoundInResponse(response string, delegator api.Delegator) bool {
	if strings.Contains(response, delegator.TotalAsBigInt.String()) {
		// we should have not encoded the total as big int
		return false
	}

	found := strings.Contains(response, delegator.DelegatorAddress) && strings.Contains(response, delegator.Total)
	for _, delegatedTo := range delegator.DelegatedTo {
		found = found && strings.Contains(response, delegatedTo.Value) && strings.Contains(response, delegatedTo.DelegationScAddress)
	}

	return found
}

func TestDelegatedInfo_CannotGetDelegatedList(t *testing.T) {
	expectedError := fmt.Errorf("%s", "expected error")
	facade := mock.FacadeStub{
		GetDelegatorsListHandler: func() ([]*api.Delegator, error) {
			return nil, expectedError
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/delegated-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(respStr, expectedError.Error()))
}

func TestGetEnableEpochs_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				EnableEpochsMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/enable-epochs", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestGetEnableEpochs_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetrics := statusHandler.NewStatusMetrics()
	key := common.MetricScDeployEnableEpoch
	val := uint64(4)
	statusMetrics.SetUInt64Value(key, val)

	facade := mock.FacadeStub{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetrics
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/enable-epochs", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, strconv.FormatUint(val, 10))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestGetESDTTotalSupply_InternalError(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetTokenSupplyCalled: func(token string) (*api.ESDTSupply, error) {
			return nil, expectedErr
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdt/supply/token", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)

	keyAndValueInResponse := strings.Contains(respStr, expectedErr.Error())
	require.True(t, keyAndValueInResponse)
}

func TestGetNetworkRatings_ShouldReturnErrorIfFacadeReturnsError(t *testing.T) {
	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				RatingsMetricsCalled: func() (map[string]interface{}, error) {
					return nil, expectedErr
				},
			}
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/ratings", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedErr.Error(), response.Error)
}

func TestGetNetworkRatings_ShouldWork(t *testing.T) {
	expectedMap := map[string]interface{}{
		"key0": "val0",
	}

	facade := mock.FacadeStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			return &testscommon.StatusMetricsStub{
				RatingsMetricsCalled: func() (map[string]interface{}, error) {
					return expectedMap, nil
				},
			}
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/ratings", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	response := &ratingsConfigResponse{}
	loadResponse(resp.Body, response)

	assert.Equal(t, expectedMap, response.Data.Config)
}

func TestGetESDTTotalSupply(t *testing.T) {
	t.Parallel()

	type supplyResponse struct {
		Data *api.ESDTSupply `json:"data"`
	}

	facade := mock.FacadeStub{
		GetTokenSupplyCalled: func(token string) (*api.ESDTSupply, error) {
			return &api.ESDTSupply{
				Supply: "1000",
				Burned: "500",
				Minted: "1500",
			}, nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdt/supply/mytoken-aabb", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := io.ReadAll(resp.Body)
	assert.Equal(t, resp.Code, http.StatusOK)

	respSupply := &supplyResponse{}
	err = json.Unmarshal(respBytes, respSupply)
	require.Nil(t, err)

	require.Equal(t, &supplyResponse{Data: &api.ESDTSupply{
		Supply: "1000",
		Burned: "500",
		Minted: "1500",
	}}, respSupply)
}

func TestGetGenesisNodes(t *testing.T) {
	t.Parallel()

	t.Run("facade error, should fail", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetGenesisNodesPubKeysCalled: func() (map[uint32][]string, map[uint32][]string, error) {
				return nil, nil, expectedErr
			},
		}

		networkGroup, err := groups.NewNetworkGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

		req, _ := http.NewRequest("GET", "/network/genesis-nodes", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := genesisNodesConfigResponse{}
		loadResponse(resp.Body, &response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		eligible := map[uint32][]string{
			1: {"pubkey1"},
		}
		waiting := map[uint32][]string{
			1: {"pubkey2"},
		}

		expectedOutput := groups.GenesisNodesConfig{
			Eligible: eligible,
			Waiting:  waiting,
		}

		facade := &mock.FacadeStub{
			GetGenesisNodesPubKeysCalled: func() (map[uint32][]string, map[uint32][]string, error) {
				return eligible, waiting, nil
			},
		}

		response := &genesisNodesConfigResponse{}
		loadNetworkGroupResponse(
			t,
			facade,
			"/network/genesis-nodes",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedOutput, response.Data.Nodes)
	})
}

func TestGetGenesisBalances(t *testing.T) {
	t.Parallel()

	t.Run("facade error, should fail", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetGenesisBalancesCalled: func() ([]*common.InitialAccountAPI, error) {
				return nil, expectedErr
			},
		}

		networkGroup, err := groups.NewNetworkGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

		req, _ := http.NewRequest("GET", "/network/genesis-balances", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := genesisBalancesResponse{}
		loadResponse(resp.Body, &response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		initialAccounts := []*common.InitialAccountAPI{
			{
				Address:      "addr0",
				StakingValue: "3700000",
			},
			{
				Address:      "addr1",
				StakingValue: "5700000",
			},
		}

		facade := &mock.FacadeStub{
			GetGenesisBalancesCalled: func() ([]*common.InitialAccountAPI, error) {
				return initialAccounts, nil
			},
		}

		response := &genesisBalancesResponse{}
		loadNetworkGroupResponse(
			t,
			facade,
			"/network/genesis-balances",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, initialAccounts, response.Data.Balances)
	})
}

func TestGetGasConfigs(t *testing.T) {
	t.Parallel()

	t.Run("facade error, should fail", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetGasConfigsCalled: func() (map[string]map[string]uint64, error) {
				return nil, expectedErr
			},
		}

		networkGroup, err := groups.NewNetworkGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

		req, _ := http.NewRequest("GET", "/network/gas-configs", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := genesisNodesConfigResponse{}
		loadResponse(resp.Body, &response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		builtInCost := map[string]uint64{
			"val1": 1,
		}
		metaChainSystemSCsCost := map[string]uint64{
			"val2": 2,
		}
		expectedMap := map[string]map[string]uint64{
			common.BuiltInCost:            builtInCost,
			common.MetaChainSystemSCsCost: metaChainSystemSCsCost,
		}

		facade := &mock.FacadeStub{
			GetGasConfigsCalled: func() (map[string]map[string]uint64, error) {
				return expectedMap, nil
			},
		}

		response := &gasConfigsResponse{}
		loadNetworkGroupResponse(
			t,
			facade,
			"/network/gas-configs",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, builtInCost, response.Data.Configs.BuiltInCost)
		assert.Equal(t, metaChainSystemSCsCost, response.Data.Configs.MetaChainSystemSCsCost)
	})
}

func TestNetworkGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		networkGroup, err := groups.NewNetworkGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = networkGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		networkGroup, err := groups.NewNetworkGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = networkGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		builtInCost := map[string]uint64{
			"val1": 1,
		}
		expectedMap := map[string]map[string]uint64{
			common.BuiltInCost: builtInCost,
		}
		facade := mock.FacadeStub{
			GetGasConfigsCalled: func() (map[string]map[string]uint64, error) {
				return expectedMap, nil
			},
		}
		networkGroup, err := groups.NewNetworkGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

		req, _ := http.NewRequest("GET", "/network/gas-configs", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		assert.Equal(t, resp.Code, http.StatusOK)
		response := gasConfigsResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, builtInCost, response.Data.Configs.BuiltInCost)

		newFacade := mock.FacadeStub{
			GetGasConfigsCalled: func() (map[string]map[string]uint64, error) {
				return nil, expectedErr
			},
		}
		err = networkGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/network/gas-configs", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
}

func TestNetworkGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	networkGroup, _ := groups.NewNetworkGroup(nil)
	require.True(t, networkGroup.IsInterfaceNil())

	networkGroup, _ = groups.NewNetworkGroup(&mock.FacadeStub{})
	require.False(t, networkGroup.IsInterfaceNil())
}

func loadNetworkGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	networkGroup, err := groups.NewNetworkGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func getNetworkRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"network": {
				Routes: []config.RouteConfig{
					{Name: "/config", Open: true},
					{Name: "/status", Open: true},
					{Name: "/economics", Open: true},
					{Name: "/esdts", Open: true},
					{Name: "/total-staked", Open: true},
					{Name: "/enable-epochs", Open: true},
					{Name: "/direct-staked-info", Open: true},
					{Name: "/delegated-info", Open: true},
					{Name: "/esdt/supply/:token", Open: true},
					{Name: "/genesis-nodes", Open: true},
					{Name: "/genesis-balances", Open: true},
					{Name: "/ratings", Open: true},
					{Name: "/gas-configs", Open: true},
				},
			},
		},
	}
}
