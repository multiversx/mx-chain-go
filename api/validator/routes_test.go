package validator_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type ValidatorStatisticsResponse struct {
	Result map[string]*state.ValidatorApiResponse `json:"statistics"`
	Error  string                                 `json:"error"`
}

func TestValidatorStatistics_ErrorWithWrongFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/validator/statistics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestValidatorStatistics_ErrorWhenFacadeFails(t *testing.T) {
	t.Parallel()

	errStr := "error in facade"

	facade := mock.Facade{
		ValidatorStatisticsHandler: func() (map[string]*state.ValidatorApiResponse, error) {
			return nil, errors.New(errStr)
		},
	}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/validator/statistics", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := ValidatorStatisticsResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, response.Error, errStr)
}

func TestValidatorStatistics_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	mapToReturn := make(map[string]*state.ValidatorApiResponse)
	mapToReturn["test"] = &state.ValidatorApiResponse{
		NumLeaderSuccess:    5,
		NumLeaderFailure:    2,
		NumValidatorSuccess: 7,
		NumValidatorFailure: 3,
	}

	facade := mock.Facade{
		ValidatorStatisticsHandler: func() (map[string]*state.ValidatorApiResponse, error) {
			return mapToReturn, nil
		},
	}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/validator/statistics", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	validatorStatistics := ValidatorStatisticsResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &validatorStatistics)

	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, validatorStatistics.Result, mapToReturn)
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

func startNodeServer(handler validator.ValidatorsStatisticsApiHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ginValidatorRoute := ws.Group("/validator")
	if handler != nil {
		ginValidatorRoute.Use(middleware.WithTestingElrondFacade(handler))
	}
	validatorRoute, _ := wrapper.NewRouterWrapper("validator", ginValidatorRoute, getRoutesConfig())
	validator.Routes(validatorRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	ginValidatorRoute := ws.Group("/validator")
	validatorRoute, _ := wrapper.NewRouterWrapper("validator", ginValidatorRoute, getRoutesConfig())
	validator.Routes(validatorRoute)
	return ws
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"validator": {
				[]config.RouteConfig{
					{Name: "/statistics", Open: true},
				},
			},
		},
	}
}
