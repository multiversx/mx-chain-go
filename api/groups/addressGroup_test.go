package groups_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

var expectedErr = errors.New("expected err")

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

func TestAddressGroup_getAccount(t *testing.T) {
	t.Parallel()

	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/addr",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{
					Address:         "addr",
					Balance:         big.NewInt(100).String(),
					Nonce:           1,
					DeveloperReward: big.NewInt(120).String(),
				}, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/addr"), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		mapResponse := response.Data.(map[string]interface{})
		accResp := accountResponse{}

		mapResponseBytes, _ := json.Marshal(&mapResponse)
		_ = json.Unmarshal(mapResponseBytes, &accResp)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "addr", accResp.Account.Address)
		assert.Equal(t, uint64(1), accResp.Account.Nonce)
		assert.Equal(t, "100", accResp.Account.Balance)
		assert.Equal(t, "120", accResp.Account.DeveloperReward)
		assert.Empty(t, response.Error)
	})
}

func TestAddressGroup_GetBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//balance", nil,
			formatExpectedErr(apiErrors.ErrGetBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/balance?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetBalance, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/balance",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
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
	})
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

func TestAddressGroup_getAccounts(t *testing.T) {
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
	t.Run("invalid query options should error",
		testErrorScenario("/address/bulk?blockNonce=not-uint64", bytes.NewBuffer([]byte(`["erd1", "erd1"]`)),
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("facade error, should err", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		addrGroup, _ := groups.NewAddressGroup(&facade)

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
		facade := mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return expectedAccounts, api.BlockInfo{}, nil
			},
		}
		addrGroup, _ := groups.NewAddressGroup(&facade)

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

func TestAddressGroup_getUsername(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//username", nil,
			formatExpectedErr(apiErrors.ErrGetUsername, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/username?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetUsername, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return "", api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/username",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetUsername, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testUsername := "provided username"
		facade := mock.FacadeStub{
			GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testUsername, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/username", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		usernameResponseObj := usernameResponse{}
		loadResponse(resp.Body, &usernameResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testUsername, usernameResponseObj.Data.Username)
	})
}

func TestAddressGroup_getCodeHash(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//code-hash", nil,
			formatExpectedErr(apiErrors.ErrGetCodeHash, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/code-hash?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetCodeHash, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/code-hash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetCodeHash, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

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

		req, _ := http.NewRequest("GET", "/address/erd1alice/code-hash", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		codeHashResponseObj := codeHashResponse{}
		loadResponse(resp.Body, &codeHashResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponseCodeHash, codeHashResponseObj.Data.CodeHash)
	})
}

func TestAddressGroup_getValueForKey(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//key/test", nil,
			formatExpectedErr(apiErrors.ErrGetValueForKey, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/key/test?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetValueForKey, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return "", api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/key/test",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetValueForKey, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue := "value"
		facade := mock.FacadeStub{
			GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testValue, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/key/test", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		valueForKeyResponseObj := valueForKeyResponse{}
		loadResponse(resp.Body, &valueForKeyResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
	})
}

func TestAddressGroup_getGuardianData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//guardian-data", nil,
			formatExpectedErr(apiErrors.ErrGetGuardianData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/guardian-data?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetGuardianData, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return api.GuardianData{}, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/guardian-data", nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetGuardianData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

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

		req, _ := http.NewRequest("GET", "/address/erd1alice/guardian-data", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := guardianDataResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedGuardianData, response.Data.GuardianData)
	})
}

func TestAddressGroup_getKeyValuePairs(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//keys", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/keys?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/keys",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pairs := map[string]string{
			"k1": "v1",
			"k2": "v2",
		}
		facade := mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
				return pairs, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/keys", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := keyValuePairsResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, pairs, response.Data.Pairs)
	})
}

func TestAddressGroup_getESDTBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdt/newToken", nil,
			formatExpectedErr(apiErrors.ErrGetESDTBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdt/newToken?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetESDTBalance, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
				return &esdt.ESDigitalToken{}, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdt/newToken",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

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

		req, _ := http.NewRequest("GET", "/address/erd1alice/esdt/newToken", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		esdtBalanceResponseObj := esdtTokenResponse{}
		loadResponse(resp.Body, &esdtBalanceResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testValue, esdtBalanceResponseObj.Data.Balance)
		assert.Equal(t, "000100", esdtBalanceResponseObj.Data.Properties)
	})
}

func TestAddressGroup_getESDTsRoles(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdts/roles", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdts/roles?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdts/roles",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		roles := map[string][]string{
			"token0": {"role0", "role1"},
			"token1": {"role3", "role1"},
		}
		facade := mock.FacadeStub{
			GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return roles, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/esdts/roles", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := esdtRolesResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, roles, response.Data.Roles)
	})
}

func TestAddressGroup_getESDTTokensWithRole(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdts-with-role/ESDTRoleNFTCreate", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, apiErrors.ErrBadUrlParams)))
	t.Run("invalid role should error",
		testErrorScenario("/address/erd1alice/esdts-with-role/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, fmt.Errorf("invalid role: %s", "invalid"))))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTsWithRoleCalled: func(_ string, _ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := mock.FacadeStub{
			GetESDTsWithRoleCalled: func(address string, role string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return expectedTokens, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		esdtResponseObj := esdtsWithRoleResponse{}
		loadResponse(resp.Body, &esdtResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
	})
}

func TestAddressGroup_getNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//registered-nfts", nil,
			formatExpectedErr(apiErrors.ErrNFTTokenIDsRegistered, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/registered-nfts?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrNFTTokenIDsRegistered, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(_ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/registered-nfts",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrNFTTokenIDsRegistered, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(address string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return expectedTokens, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", "/address/erd1alice/registered-nfts", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		esdtResponseObj := esdtsWithRoleResponse{}
		loadResponse(resp.Body, &esdtResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
	})
}

func TestAddressGroup_getESDTNFTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//nft/newToken/nonce/10", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/nft/newToken/nonce/10?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("invalid nonce should error",
		testErrorScenario("/address/erd1alice/nft/newToken/nonce/not-int", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrNonceInvalid)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/nft/newToken/nonce/10",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
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

		req, _ := http.NewRequest("GET", "/address/erd1alice/nft/newToken/nonce/10", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		esdtResponseObj := esdtNFTResponse{}
		loadResponse(resp.Body, &esdtResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testValue, esdtResponseObj.Data.Balance)
		assert.Equal(t, "010000", esdtResponseObj.Data.Properties)
		assert.Equal(t, testAddress, esdtResponseObj.Data.Creator)
		assert.Equal(t, testNonce, esdtResponseObj.Data.Nonce)
	})
}

func TestAddressGroup_getAllESDTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdt", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdt?blockNonce=not-uint64", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAllESDTTokensCalled: func(address string, options api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdt",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

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

		req, _ := http.NewRequest("GET", "/address/erd1alice/esdt", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		esdtTokenResponseObj := esdtTokensCompleteResponse{}
		loadResponse(resp.Body, &esdtTokenResponseObj)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 2, len(esdtTokenResponseObj.Data.Tokens))
	})
}

func TestAddressGroup_UpdateFacadeStub(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		addrGroup, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = addrGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		addrGroup, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = addrGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
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
	})
}

func TestAddressGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	addrGroup, _ := groups.NewAddressGroup(nil)
	require.True(t, addrGroup.IsInterfaceNil())

	addrGroup, _ = groups.NewAddressGroup(&mock.FacadeStub{})
	require.False(t, addrGroup.IsInterfaceNil())
}

func testErrorScenario(url string, body io.Reader, expectedErr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testAddressGroup(
			t,
			&mock.FacadeStub{},
			url,
			body,
			http.StatusBadRequest,
			expectedErr,
		)
	}
}

func testAddressGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	body io.Reader,
	expectedRespCode int,
	expectedRespError string,
) {
	addrGroup, err := groups.NewAddressGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, expectedRespCode, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedRespError))
}

func formatExpectedErr(err, innerErr error) string {
	return fmt.Sprintf("%s: %s", err.Error(), innerErr.Error())
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
