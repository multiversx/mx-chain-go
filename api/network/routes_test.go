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
	key := core.MetricScDeployEpoch
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
					{Name: "/total-staked", Open: true},
					{Name: "/enable-epochs", Open: true},
				},
			},
		},
	}
}
