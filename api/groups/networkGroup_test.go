package groups_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
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
	Code  string
}

func TestNetworkConfigMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricMinGasLimit
	value := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, value)

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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", value))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestNetworkStatusMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricEpochNumber
	value := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, value)

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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", value))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestEconomicsMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricTotalSupply
	value := "12345"
	statusMetricsProvider.SetStringValue(key, value)

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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, value)
	assert.True(t, keyAndValueFoundInResponse)
}

func TestEconomicsMetrics_CannotGetStakeValues(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := common.MetricTotalSupply
	value := "12345"
	statusMetricsProvider.SetStringValue(key, value)

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

	respBytes, _ := ioutil.ReadAll(resp.Body)
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

	respBytes, _ := ioutil.ReadAll(resp.Body)
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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	valuesFoundInResponse := delegatorFoundInResponse(respStr, delegator1) && delegatorFoundInResponse(respStr, delegator2)
	assert.True(t, valuesFoundInResponse)
}

func delegatorFoundInResponse(response string, delegator api.Delegator) bool {
	if strings.Contains(response, delegator.TotalAsBigInt.String()) {
		//we should have not encoded the total as big int
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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(respStr, expectedError.Error()))
}

func TestGetEnableEpochs_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetrics := statusHandler.NewStatusMetrics()
	key := common.MetricScDeployEnableEpoch
	value := uint64(4)
	statusMetrics.SetUInt64Value(key, value)

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

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, strconv.FormatUint(value, 10))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestGetESDTTotalSupply_InternalError(t *testing.T) {
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetTokenSupplyCalled: func(token string) (string, error) {
			return "", expectedErr
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdt/supply/token", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)

	keyAndValueInResponse := strings.Contains(respStr, expectedErr.Error())
	require.True(t, keyAndValueInResponse)
}

func TestGetESDTTotalSupply(t *testing.T) {
	facade := mock.FacadeStub{
		GetTokenSupplyCalled: func(token string) (string, error) {
			return "1000", nil
		},
	}

	networkGroup, err := groups.NewNetworkGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(networkGroup, "network", getNetworkRoutesConfig())

	req, _ := http.NewRequest("GET", "/network/esdt/supply/mytoken-aabb", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueInResponse := strings.Contains(respStr, "supply") && strings.Contains(respStr, "1000")
	require.True(t, keyAndValueInResponse)
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
				},
			},
		},
	}
}
