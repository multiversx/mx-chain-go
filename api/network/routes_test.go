package network_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/network"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type esdtTokensResponseData struct {
	Tokens []string `json:"tokens"`
}

type esdtTokensResponse struct {
	Data  esdtTokensResponseData `json:"data"`
	Error string                 `json:"error"`
	Code  string
}

func TestNetworkConfigMetrics_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))
}

func TestNetworkStatusMetrics_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/network/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))
}

func TestNetworkConfigMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := core.MetricMinGasLimit
	value := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, value)

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", value))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestNetwork_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/network/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := GeneralResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestNetworkStatusMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := core.MetricEpochNumber
	value := uint64(37)
	statusMetricsProvider.SetUInt64Value(key, value)

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, fmt.Sprintf("%d", value))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestNetworkStatus_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/network/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := GeneralResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestEconomicsMetrics_NilContextShouldErr(t *testing.T) {
	ws := startNodeServer(nil)
	req, _ := http.NewRequest("GET", "/network/economics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))

}

func TestEconomicsMetrics_ShouldWork(t *testing.T) {
	statusMetricsProvider := statusHandler.NewStatusMetrics()
	key := core.MetricTotalSupply
	value := "12345"
	statusMetricsProvider.SetStringValue(key, value)

	facade := mock.Facade{
		GetTotalStakedValueHandler: func() (*api.StakeValues, error) {
			return &api.StakeValues{
				TotalStaked: big.NewInt(100),
				TopUp:       big.NewInt(20),
			}, nil
		},
	}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	ws := startNodeServer(&facade)
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
	key := core.MetricTotalSupply
	value := "12345"
	statusMetricsProvider.SetStringValue(key, value)

	localErr := fmt.Errorf("%s", "local error")
	facade := mock.Facade{
		GetTotalStakedValueHandler: func() (*api.StakeValues, error) {
			return nil, localErr
		},
	}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetricsProvider
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/economics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestGetAllIssuedESDTs_ShouldWork(t *testing.T) {
	tokens := []string{"tokenA", "tokenB"}
	facade := mock.Facade{
		GetAllIssuedESDTsCalled: func() ([]string, error) {
			return tokens, nil
		},
	}

	ws := startNodeServer(&facade)
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
	facade := mock.Facade{
		GetAllIssuedESDTsCalled: func() ([]string, error) {
			return nil, localErr
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/esdts", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestDirectStakedInfo_NilContextShouldErr(t *testing.T) {
	ws := startNodeServer(nil)
	req, _ := http.NewRequest("GET", "/network/direct-staked-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))
}

func TestDirectStakedInfo_ShouldWork(t *testing.T) {
	stakedData1 := api.DirectStakedValue{
		Address: "addr1",
		Staked:  "1111",
		TopUp:   "2222",
		Total:   "3333",
	}
	stakedData2 := api.DirectStakedValue{
		Address: "addr2",
		Staked:  "4444",
		TopUp:   "5555",
		Total:   "9999",
	}

	facade := mock.Facade{
		GetDirectStakedListHandler: func() ([]*api.DirectStakedValue, error) {
			return []*api.DirectStakedValue{&stakedData1, &stakedData2}, nil
		},
	}

	ws := startNodeServer(&facade)
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
		strings.Contains(response, stakedData.Staked) && strings.Contains(response, stakedData.TopUp)
}

func TestDirectStakedInfo_CannotGetDirectStakedList(t *testing.T) {
	expectedError := fmt.Errorf("%s", "expected error")
	facade := mock.Facade{
		GetDirectStakedListHandler: func() ([]*api.DirectStakedValue, error) {
			return nil, expectedError
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/direct-staked-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(respStr, expectedError.Error()))
}

func TestDelegatedInfo_NilContextShouldErr(t *testing.T) {
	ws := startNodeServer(nil)
	req, _ := http.NewRequest("GET", "/network/delegated-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))
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

	facade := mock.Facade{
		GetDelegatorsListHandler: func() ([]*api.Delegator, error) {
			return []*api.Delegator{&delegator1, &delegator2}, nil
		},
	}

	ws := startNodeServer(&facade)
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
	facade := mock.Facade{
		GetDelegatorsListHandler: func() ([]*api.Delegator, error) {
			return nil, expectedError
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/delegated-info", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(respStr, expectedError.Error()))
}

func TestGetEnableEpochs_NilContextShouldErr(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/network/enable-epochs", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, errors.ErrNilAppContext.Error()))
}

func TestGetEnableEpochs_ShouldWork(t *testing.T) {
	t.Parallel()

	statusMetrics := statusHandler.NewStatusMetrics()
	key := core.MetricScDeployEnableEpoch
	value := uint64(4)
	statusMetrics.SetUInt64Value(key, value)

	facade := mock.Facade{}
	facade.StatusMetricsHandler = func() external.StatusMetricsHandler {
		return statusMetrics
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/network/enable-epochs", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	respBytes, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respBytes)
	assert.Equal(t, resp.Code, http.StatusOK)

	keyAndValueFoundInResponse := strings.Contains(respStr, key) && strings.Contains(respStr, strconv.FormatUint(value, 10))
	assert.True(t, keyAndValueFoundInResponse)
}

func TestGetEnableEpochs_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/network/enable-epochs", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := GeneralResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	logError(err)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler network.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	networkRoutes := ws.Group("/network")
	if handler != nil {
		networkRoutes.Use(middleware.WithFacade(handler))
	}
	networkRouteWrapper, _ := wrapper.NewRouterWrapper("network", networkRoutes, getRoutesConfig())
	network.Routes(networkRouteWrapper)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("facade", mock.WrongFacade{})
	})
	networkRoute := ws.Group("/network")
	networkRouteWrapper, _ := wrapper.NewRouterWrapper("network", networkRoute, getRoutesConfig())
	network.Routes(networkRouteWrapper)
	return ws
}

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func getRoutesConfig() config.ApiRoutesConfig {
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
				},
			},
		},
	}
}
