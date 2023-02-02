package groups_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wrappedAcountResponse struct {
	Data  accountResponse `json:"data"`
	Error string          `json:"error"`
	Code  string          `json:"code"`
}

type accountResponse struct {
	Account struct {
		Address         string `json:"address"`
		Nonce           uint64 `json:"nonce"`
		Balance         string `json:"balance"`
		Code            string `json:"code"`
		CodeHash        []byte `json:"codeHash"`
		RootHash        []byte `json:"rootHash"`
		DeveloperReward string `json:"developerReward"`
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

type esdtTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

type esdtNFTTokenData struct {
	TokenIdentifier string   `json:"tokenIdentifier"`
	Balance         string   `json:"balance"`
	Properties      string   `json:"properties"`
	Name            string   `json:"name"`
	Nonce           uint64   `json:"nonce"`
	Creator         string   `json:"creator"`
	Royalties       string   `json:"royalties"`
	Hash            []byte   `json:"hash"`
	URIs            [][]byte `json:"uris"`
	Attributes      []byte   `json:"attributes"`
}

type esdtNFTResponseData struct {
	esdtNFTTokenData `json:"tokenData"`
}

type esdtTokenResponseData struct {
	esdtTokenData `json:"tokenData"`
}

type esdtsWithRoleResponseData struct {
	Tokens []string `json:"tokens"`
}

type esdtsWithRoleResponse struct {
	Data  esdtsWithRoleResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

type esdtTokenResponse struct {
	Data  esdtTokenResponseData `json:"data"`
	Error string                `json:"error"`
	Code  string                `json:"code"`
}

type guardianDataResponseData struct {
	GuardianData api.GuardianData `json:"guardianData"`
}

type guardianDataResponse struct {
	Data  guardianDataResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

type esdtNFTResponse struct {
	Data  esdtNFTResponseData `json:"data"`
	Error string              `json:"error"`
	Code  string              `json:"code"`
}

type esdtTokensCompleteResponseData struct {
	Tokens map[string]esdtNFTTokenData `json:"esdts"`
}

type esdtTokensCompleteResponse struct {
	Data  esdtTokensCompleteResponseData `json:"data"`
	Error string                         `json:"error"`
	Code  string
}

type keyValuePairsResponseData struct {
	Pairs map[string]string `json:"pairs"`
}

type keyValuePairsResponse struct {
	Data  keyValuePairsResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string
}

type esdtRolesResponseData struct {
	Roles map[string][]string `json:"roles"`
}

type esdtRolesResponse struct {
	Data  esdtRolesResponseData `json:"data"`
	Error string                `json:"error"`
	Code  string
}

type usernameResponseData struct {
	Username string `json:"username"`
}

type usernameResponse struct {
	Data  usernameResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

type codeHashResponseData struct {
	CodeHash string `json:"codeHash"`
}

type codeHashResponse struct {
	Data  codeHashResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

func TestNewAddressGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestAddressRoute_EmptyTrailReturns404(t *testing.T) {
	t.Parallel()
	facade := mock.FacadeStub{}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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

func TestGetBalance_WithCorrectAddressShouldNotReturnError(t *testing.T) {
	t.Parallel()

	amount := big.NewInt(10)
	addr := "testAddress"
	facade := mock.FacadeStub{
		GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
			return amount, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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
	facade := mock.FacadeStub{
		GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
			return big.NewInt(0), api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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
	facade := mock.FacadeStub{
		GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
			return nil, api.BlockInfo{}, balanceError
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/balance", addr), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, fmt.Sprintf("%s: %s", apiErrors.ErrGetBalance.Error(), balanceError.Error()), response.Error)
}

func TestGetBalance_WithEmptyAddressShouldReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.FacadeStub{
		GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
			return big.NewInt(0), api.BlockInfo{}, errors.New("address was empty")
		},
	}

	emptyAddress := ""

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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

func TestGetValueForKey_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
			return "", api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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
	facade := mock.FacadeStub{
		GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
			return testValue, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/key/test", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	valueForKeyResponseObj := valueForKeyResponse{}
	loadResponse(resp.Body, &valueForKeyResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
}

func TestGetAccounts(t *testing.T) {
	t.Parallel()

	t.Run("wrong request, should err", func(t *testing.T) {
		t.Parallel()

		addrGroup, _ := groups.NewAddressGroup(&mock.FacadeStub{})

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		invalidRequest := []byte("{invalid json}")
		req, _ := http.NewRequest("POST", "/address/bulk", bytes.NewBuffer(invalidRequest))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		require.NotEmpty(t, response.Error)
		require.Equal(t, shared.ReturnCodeRequestError, response.Code)
	})

	t.Run("facade error, should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		facade := &mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		addrGroup, _ := groups.NewAddressGroup(facade)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("POST", "/address/bulk", bytes.NewBuffer([]byte(`["erd1", "erd1"]`)))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		require.NotEmpty(t, response.Error)
		require.Equal(t, shared.ReturnCodeInternalError, response.Code)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedAccounts := map[string]*api.AccountResponse{
			"erd1alice": {
				Address: "erd1alice",
				Balance: "100000000000000",
				Nonce:   37,
			},
		}
		facade := &mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return expectedAccounts, api.BlockInfo{}, nil
			},
		}
		addrGroup, _ := groups.NewAddressGroup(facade)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("POST", "/address/bulk", bytes.NewBuffer([]byte(`["erd1", "erd1"]`)))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type responseType struct {
			Data struct {
				Accounts map[string]*api.AccountResponse `json:"accounts"`
			} `json:"data"`
			Error string            `json:"error"`
			Code  shared.ReturnCode `json:"code"`
		}
		response := responseType{}
		loadResponse(resp.Body, &response)
		require.Empty(t, response.Error)
		require.Equal(t, shared.ReturnCodeSuccess, response.Code)
		require.Equal(t, expectedAccounts, response.Data.Accounts)
	})
}

func TestGetUsername_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
			return "", api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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
	facade := mock.FacadeStub{
		GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
			return testUsername, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/username", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	usernameResponseObj := usernameResponse{}
	loadResponse(resp.Body, &usernameResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testUsername, usernameResponseObj.Data.Username)
}

func TestGetCodeHash_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/code-hash", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	codeHashResponseObj := codeHashResponse{}
	loadResponse(resp.Body, &codeHashResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(codeHashResponseObj.Error, expectedErr.Error()))
}

func TestGetCodeHash_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testCodeHash := []byte("value")
	expectedResponseCodeHash := base64.StdEncoding.EncodeToString(testCodeHash)
	facade := mock.FacadeStub{
		GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
			return testCodeHash, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/code-hash", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	codeHashResponseObj := codeHashResponse{}
	loadResponse(resp.Body, &codeHashResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedResponseCodeHash, codeHashResponseObj.Data.CodeHash)
}

func TestGetAccount_FailWhenFacadeStubGetAccountFails(t *testing.T) {
	t.Parallel()

	returnedError := "i am an error"
	facade := mock.FacadeStub{
		GetAccountCalled: func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			return api.AccountResponse{}, api.BlockInfo{}, errors.New(returnedError)
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

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

	facade := mock.FacadeStub{
		GetAccountCalled: func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			return api.AccountResponse{
				Address:         "1234",
				Balance:         big.NewInt(100).String(),
				Nonce:           1,
				DeveloperReward: big.NewInt(120).String(),
			}, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	reqAddress := "test"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s", reqAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	mapResponse := response.Data.(map[string]interface{})
	accountResponse := accountResponse{}

	mapResponseBytes, _ := json.Marshal(&mapResponse)
	_ = json.Unmarshal(mapResponseBytes, &accountResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, reqAddress, accountResponse.Account.Address)
	assert.Equal(t, uint64(1), accountResponse.Account.Nonce)
	assert.Equal(t, "100", accountResponse.Account.Balance)
	assert.Equal(t, "120", accountResponse.Account.DeveloperReward)
	assert.Empty(t, response.Error)
}

func TestGetAccount_WithBadQueryOptionsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetAccountCalled: func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			return api.AccountResponse{Nonce: 1}, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	response, code := httpGetAccount(ws, "/address/alice?onFinalBlock=bad")
	require.Equal(t, http.StatusBadRequest, code)
	require.Contains(t, response.Error, apiErrors.ErrBadUrlParams.Error())

	response, code = httpGetAccount(ws, "/address/alice?onStartOfEpoch=bad")
	require.Equal(t, http.StatusBadRequest, code)
	require.Contains(t, response.Error, apiErrors.ErrBadUrlParams.Error())
}

func TestGetAccount_WithQueryOptionsShouldWork(t *testing.T) {
	t.Parallel()

	var calledWithAddress string
	var calledWithOptions api.AccountQueryOptions

	facade := mock.FacadeStub{
		GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			calledWithAddress = address
			calledWithOptions = options
			return api.AccountResponse{Nonce: 1}, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	response, code := httpGetAccount(ws, "/address/alice?onFinalBlock=true")
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, response)
	require.Equal(t, "alice", calledWithAddress)
	require.Equal(t, api.AccountQueryOptions{OnFinalBlock: true}, calledWithOptions)
}

func httpGetAccount(ws *gin.Engine, url string) (wrappedAcountResponse, int) {
	httpRequest, _ := http.NewRequest("GET", url, nil)
	httpResponse := httptest.NewRecorder()
	ws.ServeHTTP(httpResponse, httpRequest)

	accountResponse := wrappedAcountResponse{}
	loadResponse(httpResponse.Body, &accountResponse)
	return accountResponse, httpResponse.Code
}

func TestGetESDTBalance_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt/newToken", testAddress), nil)
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
	testValue := big.NewInt(100).String()
	testProperties := []byte{byte(0), byte(1), byte(0)}
	facade := mock.FacadeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
			return &esdt.ESDigitalToken{Value: big.NewInt(100), Properties: testProperties}, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt/newToken", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtBalanceResponseObj := esdtTokenResponse{}
	loadResponse(resp.Body, &esdtBalanceResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, esdtBalanceResponseObj.Data.Balance)
	assert.Equal(t, "000100", esdtBalanceResponseObj.Data.Properties)
}

func TestGetESDTNFTData_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/nft/newToken/nonce/10", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtNFTResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(esdtResponseObj.Error, expectedErr.Error()))
}

func TestGetESDTNFTData_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testValue := big.NewInt(100).String()
	testNonce := uint64(37)
	testProperties := []byte{byte(1), byte(0), byte(0)}
	facade := mock.FacadeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
			return &esdt.ESDigitalToken{
				Value:         big.NewInt(100),
				Properties:    testProperties,
				TokenMetaData: &esdt.MetaData{Nonce: testNonce, Creator: []byte(testAddress)}}, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/nft/newToken/nonce/10", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtNFTResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, testValue, esdtResponseObj.Data.Balance)
	assert.Equal(t, "010000", esdtResponseObj.Data.Properties)
	assert.Equal(t, testAddress, esdtResponseObj.Data.Creator)
	assert.Equal(t, testNonce, esdtResponseObj.Data.Nonce)
}

func TestGetESDTTokensWithRole_InvalidRoleShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetESDTsWithRoleCalled: func(_ string, _ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts-with-role/invalid", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtsWithRoleResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(esdtResponseObj.Error, "invalid role"))
}

func TestGetESDTTokensWithRole_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetESDTsWithRoleCalled: func(_ string, _ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts-with-role/ESDTRoleNFTCreate", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtsWithRoleResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(esdtResponseObj.Error, expectedErr.Error()))
}

func TestGetESDTTokensWithRole_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
	facade := mock.FacadeStub{
		GetESDTsWithRoleCalled: func(address string, role string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
			return expectedTokens, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts-with-role/ESDTRoleNFTCreate", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtsWithRoleResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
}

func TestGetNFTTokenIDsRegisteredByAddress_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetNFTTokenIDsRegisteredByAddressCalled: func(_ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/registered-nfts", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtsWithRoleResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(esdtResponseObj.Error, expectedErr.Error()))
}

func TestGetNFTTokenIDsRegisteredByAddress_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
	facade := mock.FacadeStub{
		GetNFTTokenIDsRegisteredByAddressCalled: func(address string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
			return expectedTokens, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/registered-nfts", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtResponseObj := esdtsWithRoleResponse{}
	loadResponse(resp.Body, &esdtResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
}

func TestGetFullESDTTokens_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetAllESDTTokensCalled: func(_ string, _ api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtTokenResponseObj := esdtTokensCompleteResponse{}
	loadResponse(resp.Body, &esdtTokenResponseObj)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(esdtTokenResponseObj.Error, expectedErr.Error()))
}

func TestGetFullESDTTokens_ShouldWork(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	testValue1 := "token1"
	testValue2 := "token2"
	facade := mock.FacadeStub{
		GetAllESDTTokensCalled: func(address string, _ api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
			tokens := make(map[string]*esdt.ESDigitalToken)
			tokens[testValue1] = &esdt.ESDigitalToken{Value: big.NewInt(10)}
			tokens[testValue2] = &esdt.ESDigitalToken{Value: big.NewInt(100)}
			return tokens, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdt", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	esdtTokenResponseObj := esdtTokensCompleteResponse{}
	loadResponse(resp.Body, &esdtTokenResponseObj)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, 2, len(esdtTokenResponseObj.Data.Tokens))
}

func TestGetKeyValuePairs_WithEmptyAddressShouldReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.FacadeStub{}

	emptyAddress := ""

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/keys", emptyAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEmpty(t, response)
	assert.True(t, strings.Contains(response.Error,
		fmt.Sprintf("%s: %s", apiErrors.ErrGetKeyValuePairs.Error(), apiErrors.ErrEmptyAddress.Error()),
	))
}

func TestGetKeyValuePairs_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/keys", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetKeyValuePairs_ShouldWork(t *testing.T) {
	t.Parallel()

	pairs := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	testAddress := "address"
	facade := mock.FacadeStub{
		GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
			return pairs, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/keys", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := keyValuePairsResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, pairs, response.Data.Pairs)
}

func TestGetGuardianData(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	t.Run("with empty address should err", func(t *testing.T) {
		facade := mock.FacadeStub{}
		addrGroup, err := groups.NewAddressGroup(&facade)
		require.Nil(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		emptyAddress := ""
		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/guardian-data", emptyAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.NotEmpty(t, response)
		assert.True(t, strings.Contains(response.Error,
			fmt.Sprintf("%s: %s", apiErrors.ErrGetGuardianData.Error(), apiErrors.ErrEmptyAddress.Error()),
		))
	})
	t.Run("with node fail should err", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		facade := mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return api.GuardianData{}, api.BlockInfo{}, expectedErr
			},
		}
		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/guardian-data", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := &shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))

	})
	t.Run("should work", func(t *testing.T) {
		expectedGuardianData := api.GuardianData{
			ActiveGuardian: &api.Guardian{
				Address:         "guardian1",
				ActivationEpoch: 0,
			},
			PendingGuardian: &api.Guardian{
				Address:         "guardian2",
				ActivationEpoch: 10,
			},
			Guarded: true,
		}
		facade := mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return expectedGuardianData, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/guardian-data", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := guardianDataResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedGuardianData, response.Data.GuardianData)
	})
}

func TestGetESDTsRoles_WithEmptyAddressShouldReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.FacadeStub{}

	emptyAddress := ""

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts/roles", emptyAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEmpty(t, response)
	assert.True(t, strings.Contains(response.Error,
		fmt.Sprintf("%s: %s", apiErrors.ErrGetRolesForAccount.Error(), apiErrors.ErrEmptyAddress.Error()),
	))
}

func TestGetESDTsRoles_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, expectedErr
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts/roles", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := &shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetESDTsRoles_ShouldWork(t *testing.T) {
	t.Parallel()

	roles := map[string][]string{
		"token0": {"role0", "role1"},
		"token1": {"role3", "role1"},
	}
	testAddress := "address"
	facade := mock.FacadeStub{
		GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
			return roles, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts/roles", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := esdtRolesResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, roles, response.Data.Roles)
}

func TestAddressGroup_UpdateFacadeStub(t *testing.T) {
	t.Parallel()

	roles := map[string][]string{
		"token0": {"role0", "role1"},
		"token1": {"role3", "role1"},
	}
	testAddress := "address"
	facade := mock.FacadeStub{
		GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
			return roles, api.BlockInfo{}, nil
		},
	}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts/roles", testAddress), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := esdtRolesResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, roles, response.Data.Roles)

	newErr := errors.New("new error")
	newFacadeStub := mock.FacadeStub{
		GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, newErr
		},
	}
	err = addrGroup.UpdateFacade(&newFacadeStub)
	require.NoError(t, err)

	req, _ = http.NewRequest("GET", fmt.Sprintf("/address/%s/esdts/roles", testAddress), nil)
	resp = httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response = esdtRolesResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, newErr.Error()))
}

func getAddressRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"address": {
				Routes: []config.RouteConfig{
					{Name: "/:address", Open: true},
					{Name: "/bulk", Open: true},
					{Name: "/:address/guardian-data", Open: true},
					{Name: "/:address/balance", Open: true},
					{Name: "/:address/username", Open: true},
					{Name: "/:address/code-hash", Open: true},
					{Name: "/:address/keys", Open: true},
					{Name: "/:address/key/:key", Open: true},
					{Name: "/:address/esdt", Open: true},
					{Name: "/:address/esdts/roles", Open: true},
					{Name: "/:address/esdt/:tokenIdentifier", Open: true},
					{Name: "/:address/nft/:tokenIdentifier/nonce/:nonce", Open: true},
					{Name: "/:address/esdts-with-role/:role", Open: true},
					{Name: "/:address/registered-nfts", Open: true},
				},
			},
		},
	}
}
