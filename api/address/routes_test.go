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

	"github.com/ElrondNetwork/elrond-go/api/address"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type AccountResponse struct {
	Account struct {
		Address  string `json:"address"`
		Nonce    uint64 `json:"nonce"`
		Balance  string `json:"balance"`
		Code     string `json:"code"`
		CodeHash []byte `json:"codeHash"`
		RootHash []byte `json:"rootHash"`
	} `json:"account"`
}

type valueForKeyResponseData struct {
	Value string `json:"value"`
}

type valueForKeyResponse struct {
	Data  valueForKeyResponseData `json:"data"`
	Error string                  `json:"error"`
	Code  string                  `json:"code"`
}

type usernameResponseData struct {
	Username string `json:"username"`
}

type usernameResponse struct {
	Data  usernameResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
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

func getValueForKey(dataFromResponse interface{}, key string) string {
	dataMap, ok := dataFromResponse.(map[string]interface{})
	if !ok {
		return ""
	}

	valueI, okCast := dataMap[key]
	if okCast {
		return fmt.Sprintf("%v", valueI)
	}
	return ""
}

func TestGetBalance_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/testAddress/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
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

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	balanceStr := getValueForKey(response.Data, "balance")
	balanceResponse, ok := big.NewInt(0).SetString(balanceStr, 10)
	assert.True(t, ok)
	assert.Equal(t, amount, balanceResponse)
	assert.Equal(t, "", response.Error)
}

func TestGetBalance_WithWrongAddressShouldError(t *testing.T) {
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

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)
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

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, fmt.Sprintf("%s: %s", apiErrors.ErrGetBalance.Error(), balanceError.Error()), response.Error)
}

func TestGetBalance_WithEmptyAddressShoudReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), errors.New("address was empty")
		},
	}

	emptyAddress := ""
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", emptyAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEmpty(t, response)
	assert.True(t, strings.Contains(response.Error,
		fmt.Sprintf("%s: %s", apiErrors.ErrGetBalance.Error(), apiErrors.ErrEmptyAddress.Error()),
	))
}

func TestGetBalance_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty/balance", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, response.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestGetValueForKey_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/testAddress/key/test", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetValueForKey_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetValueForKeyCalled: func(_ string, _ string) (string, error) {
			return "", expectedErr
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/key/test", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	valueForKeyResponseObj := valueForKeyResponse{}
	loadResponse(resp.Body, &valueForKeyResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(valueForKeyResponseObj.Error, expectedErr.Error()))
}

func TestGetValueForKey_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testValue := "value"
	facade := mock.Facade{
		GetValueForKeyCalled: func(_ string, _ string) (string, error) {
			return testValue, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/key/test", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	valueForKeyResponseObj := valueForKeyResponse{}
	loadResponse(resp.Body, &valueForKeyResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
}

func TestGetUsername_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/testAddress/username", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetUsername_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetUsernameCalled: func(_ string) (string, error) {
			return "", expectedErr
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/username", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	usernameResponseObj := usernameResponse{}
	loadResponse(resp.Body, &usernameResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(usernameResponseObj.Error, expectedErr.Error()))
}

func TestGetUsername_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testUsername := "value"
	facade := mock.Facade{
		GetUsernameCalled: func(_ string) (string, error) {
			return testUsername, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/username", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	usernameResponseObj := usernameResponse{}
	loadResponse(resp.Body, &usernameResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testUsername, usernameResponseObj.Data.Username)
}

func TestGetAccount_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetAccount_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/address/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, response.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestGetAccount_FailWhenFacadeGetAccountFails(t *testing.T) {
	t.Parallel()
	returnedError := "i am an error"
	facade := mock.Facade{
		GetAccountHandler: func(address string) (state.UserAccountHandler, error) {
			return nil, errors.New(returnedError)
		},
	}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/address/test", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Empty(t, response.Data)
	assert.NotEmpty(t, response.Error)
	assert.True(t, strings.Contains(response.Error, fmt.Sprintf("%s: %s", apiErrors.ErrCouldNotGetAccount.Error(), returnedError)))
}

func TestGetAccount_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		GetAccountHandler: func(address string) (state.UserAccountHandler, error) {
			acc, _ := state.NewUserAccount([]byte("1234"))
			_ = acc.AddToBalance(big.NewInt(100))
			acc.IncreaseNonce(1)

			return acc, nil
		},
	}
	ws := startNodeServer(&facade)

	reqAddress := "test"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s", reqAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	mapResponse := response.Data.(map[string]interface{})
	accountResponse := AccountResponse{}

	mapResponseBytes, _ := json.Marshal(&mapResponse)
	_ = json.Unmarshal(mapResponseBytes, &accountResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, accountResponse.Account.Address, reqAddress)
	assert.Equal(t, accountResponse.Account.Nonce, uint64(1))
	assert.Equal(t, accountResponse.Account.Balance, "100")
	assert.Empty(t, response.Error)
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

func startNodeServer(handler address.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	addressRoutes := ws.Group("/address")
	if handler != nil {
		addressRoutes.Use(middleware.WithFacade(handler))
	}
	addressRoute, _ := wrapper.NewRouterWrapper("address", addressRoutes, getRoutesConfig())
	address.Routes(addressRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("facade", mock.WrongFacade{})
	})
	ginAddressRoute := ws.Group("/address")
	addressRoute, _ := wrapper.NewRouterWrapper("address", ginAddressRoute, getRoutesConfig())
	address.Routes(addressRoute)
	return ws
}

func TestGetESDTBalance_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/esdt/tokenName/newToken", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetESDTBalance_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetESDTBalanceCalled: func(_ string, _ string) (string, string, error) {
			return "", "", expectedErr
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt/tokenName/newToken", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	usernameResponseObj := usernameResponse{}
	loadResponse(resp.Body, &usernameResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(usernameResponseObj.Error, expectedErr.Error()))
}

func TestGetESDTBalance_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testValue := "value"
	facade := mock.Facade{
		GetESDTBalanceCalled: func(_ string, _ string) (string, string, error) {
			return testValue, testValue, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt/tokenName/newToken", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	valueForKeyResponseObj := valueForKeyResponse{}
	loadResponse(resp.Body, &valueForKeyResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
}

func TestGetESDTTokens_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/address/some/allesdttokens", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetESDTTokens_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetAllESDTTokensCalled: func(_ string) ([]string, error) {
			return nil, expectedErr
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/allesdttokens", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	usernameResponseObj := usernameResponse{}
	loadResponse(resp.Body, &usernameResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(usernameResponseObj.Error, expectedErr.Error()))
}

func TestGetESDTTokens_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testValue := "value"
	facade := mock.Facade{
		GetAllESDTTokensCalled: func(address string) ([]string, error) {
			return []string{testValue}, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/allesdttokens", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	valueForKeyResponseObj := valueForKeyResponse{}
	loadResponse(resp.Body, &valueForKeyResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"address": {
				[]config.RouteConfig{
					{Name: "/:address", Open: true},
					{Name: "/:address/balance", Open: true},
					{Name: "/:address/username", Open: true},
					{Name: "/:address/key/:key", Open: true},
					{Name: "/:address/allesdttokens", Open: true},
					{Name: "/:address/esdt/:tokenName", Open: true},
				},
			},
		},
	}
}
