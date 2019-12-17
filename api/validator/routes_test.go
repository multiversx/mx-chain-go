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
	"github.com/ElrondNetwork/elrond-go/api/validator"
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
		NrLeaderSuccess:    5,
		NrLeaderFailure:    2,
		NrValidatorSuccess: 7,
		NrValidatorFailure: 3,
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

	response := ValidatorStatisticsResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, response.Result, mapToReturn)
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
	validatorRoute := ws.Group("/validator")
	if handler != nil {
		validatorRoute.Use(middleware.WithElrondFacade(handler))
	}
	validator.Routes(validatorRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	validatorRoute := ws.Group("/validator")
	validator.Routes(validatorRoute)
	return ws
}
