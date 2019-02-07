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
	errors2 "github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// General response structure
type GeneralResponse struct {
	Error string `json:"error"`
}

//addressResponse structure
type addressResponse struct {
	GeneralResponse
	Balance *big.Int `json:"balance"`
}

func NewAddressResponse() *addressResponse {
	return &addressResponse{
		Balance: big.NewInt(0),
	}
}

type AccountResponse struct {
	GeneralResponse
	Account struct {
		Address  string `json:"address"`
		Nonce    uint64 `json:"nonce"`
		Balance  string `json:"balance"`
		CodeHash []byte `json:"codeHash"`
		RootHash []byte `json:"rootHash"`
	} `json:"account"`
}

func TestAddressRoute_EmptyTrailReturns404(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetBalance_WithCorrectAddressShouldNotReturnError(t *testing.T) {
	t.Parallel()
	amount := big.NewInt(10)
	addr := "testAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return amount, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := NewAddressResponse()
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, amount, addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_WithWrongAddressShouldReturnZero(t *testing.T) {
	t.Parallel()
	otherAddress := "otherAddress"
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", otherAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := NewAddressResponse()
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, big.NewInt(0), addressResponse.Balance)
	assert.Equal(t, "", addressResponse.Error)
}

func TestGetBalance_NodeGetBalanceReturnsError(t *testing.T) {
	t.Parallel()
	addr := "addr"
	balanceError := errors.New("error")
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return nil, balanceError
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := NewAddressResponse()
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, fmt.Sprintf("%s: %s", errors2.ErrGetBalance.Error(), balanceError.Error()), addressResponse.Error)
}

func TestGetBalance_WithEmptyAddressShouldReturnZeroAndError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			panic("aaaa")
			return big.NewInt(0), errors.New("address was empty")
		},
	}

	emptyAddress := ""
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", emptyAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addressResponse := NewAddressResponse()
	loadResponse(resp.Body, &addressResponse)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, big.NewInt(0), addressResponse.Balance)
	assert.NotEmpty(t, addressResponse.Error)
	assert.True(t, strings.Contains(addressResponse.Error,
		fmt.Sprintf("%s: %s", errors2.ErrGetBalance.Error(), errors2.ErrEmptyAddress.Error()),
	))
}

func TestGetBalance_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := NewAddressResponse()
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors2.ErrInvalidAppContext.Error())
}

func TestGetAccount_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := NewAddressResponse()
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors2.ErrInvalidAppContext.Error())
}

func TestGetAccount_FailWhenFacadeGetAccountFails(t *testing.T) {
	t.Parallel()
	returnedError := "i am an error"
	facade := mock.Facade{
		GetAccountHandler: func(address string) (*state.Account, error) {
			return nil, errors.New(returnedError)
		},
	}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/address/test", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	accountResponse := AccountResponse{}
	loadResponse(resp.Body, &accountResponse)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Empty(t, accountResponse.Account)
	assert.NotEmpty(t, accountResponse.Error)
	assert.True(t, strings.Contains(accountResponse.Error, fmt.Sprintf("%s: %s", errors2.ErrCouldNotGetAccount.Error(), returnedError)))
}

func TestGetAccount_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		GetAccountHandler: func(address string) (*state.Account, error) {
			return &state.Account{
				Nonce:   1,
				Balance: big.NewInt(100),
			}, nil
		},
	}
	ws := startNodeServer(&facade)

	reqAddress := "test"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s", reqAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	accountResponse := AccountResponse{}
	loadResponse(resp.Body, &accountResponse)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, accountResponse.Account.Address, reqAddress)
	assert.Equal(t, accountResponse.Account.Nonce, uint64(1))
	assert.Equal(t, accountResponse.Account.Balance, "100")
	assert.Empty(t, accountResponse.Error)
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
