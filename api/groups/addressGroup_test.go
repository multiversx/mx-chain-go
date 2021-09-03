package groups_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNewAddressGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(&mock.Facade{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestAddressRoute_EmptyTrailReturns404(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}

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
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return amount, nil
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
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), nil
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
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return nil, balanceError
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

func TestGetBalance_WithEmptyAddressShoudReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		BalanceHandler: func(s string) (i *big.Int, e error) {
			return big.NewInt(0), errors.New("address was empty")
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
	facade := mock.Facade{
		GetValueForKeyCalled: func(_ string, _ string) (string, error) {
			return "", expectedErr
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
	facade := mock.Facade{
		GetValueForKeyCalled: func(_ string, _ string) (string, error) {
			return testValue, nil
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

func TestGetUsername_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetUsernameCalled: func(_ string) (string, error) {
			return "", expectedErr
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
	facade := mock.Facade{
		GetUsernameCalled: func(_ string) (string, error) {
			return testUsername, nil
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

func TestGetAccount_FailWhenFacadeGetAccountFails(t *testing.T) {
	t.Parallel()

	returnedError := "i am an error"
	facade := mock.Facade{
		GetAccountHandler: func(address string) (api.AccountResponse, error) {
			return api.AccountResponse{}, errors.New(returnedError)
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

	facade := mock.Facade{
		GetAccountHandler: func(address string) (api.AccountResponse, error) {
			return api.AccountResponse{
				Address:         "1234",
				Balance:         big.NewInt(100).String(),
				Nonce:           1,
				DeveloperReward: big.NewInt(120).String(),
			}, nil
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

func TestGetESDTBalance_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetESDTDataCalled: func(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
			return nil, expectedErr
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
	testProperties := "frozen"
	facade := mock.Facade{
		GetESDTDataCalled: func(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
			return &esdt.ESDigitalToken{Value: big.NewInt(100), Properties: []byte(testProperties)}, nil
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
	assert.Equal(t, testProperties, esdtBalanceResponseObj.Data.Properties)
}

func TestGetESDTNFTData_NodeFailsShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetESDTDataCalled: func(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
			return nil, expectedErr
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
	testProperties := "frozen"
	facade := mock.Facade{
		GetESDTDataCalled: func(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
			return &esdt.ESDigitalToken{
				Value:         big.NewInt(100),
				Properties:    []byte(testProperties),
				TokenMetaData: &esdt.MetaData{Nonce: testNonce, Creator: []byte(testAddress)}}, nil
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
	assert.Equal(t, testProperties, esdtResponseObj.Data.Properties)
	assert.Equal(t, testAddress, esdtResponseObj.Data.Creator)
	assert.Equal(t, testNonce, esdtResponseObj.Data.Nonce)
}

func TestGetESDTTokensWithRole_InvalidRoleShouldError(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")
	facade := mock.Facade{
		GetESDTsWithRoleCalled: func(_ string, _ string) ([]string, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetESDTsWithRoleCalled: func(_ string, _ string) ([]string, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetESDTsWithRoleCalled: func(address string, role string) ([]string, error) {
			return expectedTokens, nil
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
	facade := mock.Facade{
		GetNFTTokenIDsRegisteredByAddressCalled: func(_ string) ([]string, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetNFTTokenIDsRegisteredByAddressCalled: func(address string) ([]string, error) {
			return expectedTokens, nil
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
	facade := mock.Facade{
		GetAllESDTTokensCalled: func(_ string) (map[string]*esdt.ESDigitalToken, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetAllESDTTokensCalled: func(address string) (map[string]*esdt.ESDigitalToken, error) {
			tokens := make(map[string]*esdt.ESDigitalToken)
			tokens[testValue1] = &esdt.ESDigitalToken{Value: big.NewInt(10)}
			tokens[testValue2] = &esdt.ESDigitalToken{Value: big.NewInt(100)}
			return tokens, nil
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

func TestGetKeyValuePairs_WithEmptyAddressShoudReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}

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
	facade := mock.Facade{
		GetKeyValuePairsCalled: func(_ string) (map[string]string, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetKeyValuePairsCalled: func(_ string) (map[string]string, error) {
			return pairs, nil
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

func TestGetESDTsRoles_WithEmptyAddressShoudReturnError(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{}

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
	facade := mock.Facade{
		GetESDTsRolesCalled: func(_ string) (map[string][]string, error) {
			return nil, expectedErr
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
	facade := mock.Facade{
		GetESDTsRolesCalled: func(_ string) (map[string][]string, error) {
			return roles, nil
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

func TestAddressGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	roles := map[string][]string{
		"token0": {"role0", "role1"},
		"token1": {"role3", "role1"},
	}
	testAddress := "address"
	facade := mock.Facade{
		GetESDTsRolesCalled: func(_ string) (map[string][]string, error) {
			return roles, nil
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
	newFacade := mock.Facade{
		GetESDTsRolesCalled: func(_ string) (map[string][]string, error) {
			return nil, newErr
		},
	}
	err = addrGroup.UpdateFacade(&newFacade)
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
					{Name: "/:address/balance", Open: true},
					{Name: "/:address/username", Open: true},
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
