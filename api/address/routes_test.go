package address_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/address"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/address/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type AddressResponse struct {
	GeneralResponse
	Balance big.Int `json:"balance"`
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

func TestGetBalance_WithCorrectAddress_ShouldNotReturnError(t *testing.T) {
	t.Parallel()
	amount := big.NewInt(10)
	addr := "testAddress"
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			if s == addr {
				return amount, nil
			}
			return nil, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/addr/"+addr+"/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, *amount, addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_WithWrongAddress_ShouldReturnZero(t *testing.T) {
	t.Parallel()
	amount := big.NewInt(10)
	addr := "testAddress"
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			if s == addr {
				return amount, nil
			}
			return big.NewInt(0), nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/addr/"+"otherAddress"+"/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, *big.NewInt(0), addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_WithEmptyAddress_ShouldReturnZeroAndError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		GetBalanceCalled: func(s string) (i *big.Int, e error) {
			if s == "" {
				return big.NewInt(0), errors.New("address was empty")
			}
			return big.NewInt(0), nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/address/"+""+"/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := AddressResponse{}
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, *big.NewInt(0), addressResponse.Balance)
	assert.NotEmpty(t, addressResponse.Message)
	assert.True(t, strings.Contains(addressResponse.Message, "Get balance error:"))
}
