package address_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/address"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/address/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

//General response structure
type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

//Address Response structure
type AddressResponse struct {
	GeneralResponse
	Balance big.Int `json:"balance"`
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

func startNodeServer(handler address.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	addressRoutes := ws.Group("/address")
	if handler != nil {
		addressRoutes.Use(middleware.WithElrondFacade(handler))
	}
	address.Routes(addressRoutes)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	addressRoute := ws.Group("/address")
	address.Routes(addressRoute)
	return ws
}

func TestGetBalance_WithCorrectAddressShouldNotReturnError(t *testing.T) {
	t.Parallel()
	amount := big.NewInt(10)
	addr := "testAddress"
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			return amount, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, *amount, addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_WithWrongAddressShouldReturnZero(t *testing.T) {
	t.Parallel()
	otherAddress := "otherAddress"
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", otherAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, *big.NewInt(0), addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_WithEmptyAddressShouldReturnZeroAndError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), errors.New("address was empty")
		},
	}

	emptyAddress := ""
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", emptyAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, *big.NewInt(0), addressResponse.Balance)
	assert.NotEmpty(t, addressResponse.Error)
	assert.True(t, strings.Contains(addressResponse.Error, "Get balance error: Address was empty"))
}

func TestGetAddress_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := AddressResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, "Invalid app context")
}

func TestGetBalance_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := AddressResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, "Invalid app context")
}
