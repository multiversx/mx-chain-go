package groups_test

import (
	"bytes"
	"context"
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
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
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
	Tokens        []string               `json:"tokens"`
	IteratorState *checkpoint.Checkpoint `json:"iteratorState"`
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
	Tokens        map[string]esdtNFTTokenData `json:"esdts"`
	IteratorState *checkpoint.Checkpoint      `json:"iteratorState"`
}

type esdtTokensCompleteResponse struct {
	Data  esdtTokensCompleteResponseData `json:"data"`
	Error string                         `json:"error"`
	Code  string                         `json:"code"`
}

type keyValuePairsResponseData struct {
	Pairs         map[string]string      `json:"pairs"`
	IteratorState *checkpoint.Checkpoint `json:"iteratorState"`
}

type accountWithKeysResponseData struct {
	Account struct {
		Address string            `json:"address"`
		Pairs   map[string]string `json:"pairs"`
	} `json:"account"`
	IteratorState *checkpoint.Checkpoint `json:"iteratorState"`
}

type accountWithKeysResponse struct {
	Data  accountWithKeysResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

type keyValuePairsResponse struct {
	Data  keyValuePairsResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

type iterateKeysResponseData struct {
	Pairs            map[string]string `json:"pairs"`
	NewIteratorState [][]byte          `json:"newIteratorState"`
}
type iterateKeysResponse struct {
	Data  iterateKeysResponseData `json:"data"`
	Error string                  `json:"error"`
	Code  string                  `json:"code"`
}

type esdtRolesResponseData struct {
	Roles         map[string][]string    `json:"roles"`
	IteratorState *checkpoint.Checkpoint `json:"iteratorState"`
}

type esdtRolesResponse struct {
	Data  esdtRolesResponseData `json:"data"`
	Error string                `json:"error"`
	Code  string                `json:"code"`
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

func TestAddressGroup_getAccount(t *testing.T) {
	t.Parallel()

	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(_ context.Context, address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/addr",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(_ context.Context, address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{
					Address:         "addr",
					Balance:         big.NewInt(100).String(),
					Nonce:           1,
					DeveloperReward: big.NewInt(120).String(),
				}, api.BlockInfo{}, nil
			},
		}

		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr", "GET", nil, response)

		mapResponse := response.Data.(map[string]interface{})
		accResp := accountResponse{}

		mapResponseBytes, _ := json.Marshal(&mapResponse)
		_ = json.Unmarshal(mapResponseBytes, &accResp)

		assert.Equal(t, "addr", accResp.Account.Address)
		assert.Equal(t, uint64(1), accResp.Account.Nonce)
		assert.Equal(t, "100", accResp.Account.Balance)
		assert.Equal(t, "120", accResp.Account.DeveloperReward)
		assert.Empty(t, response.Error)
	})
	t.Run("POST without withKeys pages the keys", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountWithKeysCalled: func(_ context.Context, address string, options api.AccountQueryOptions, numKeys uint, _ *checkpoint.Checkpoint) (api.AccountResponse, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.WithKeys)
				assert.Equal(t, uint(3), numKeys)
				return api.AccountResponse{Address: address}, nil, api.BlockInfo{}, nil
			},
		}
		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr", "POST", bytes.NewBufferString(`{"numKeys":3}`), response)
		assert.Empty(t, response.Error)
	})
	t.Run("paging parameters are ignored without withKeys, as before", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(_ context.Context, address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{Address: address}, api.BlockInfo{}, nil
			},
		}
		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr?numKeys=abc&iteratorState=!!!", "GET", nil, response)
		assert.Empty(t, response.Error)
	})
	t.Run("invalid withKeys should error",
		testErrorScenario("/address/erd1alice?withKeys=not-bool", "GET", nil,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, errors.New("strconv.ParseBool: parsing \"not-bool\": invalid syntax"))))
	t.Run("invalid iterator state should error",
		testErrorScenario("/address/erd1alice?withKeys=true&iteratorState=!!!", "GET", nil,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("with keys, unusable iterator state should be a bad request", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountWithKeysCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (api.AccountResponse, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return api.AccountResponse{}, nil, api.BlockInfo{}, fmt.Errorf("%w: pruned", common.ErrIteratorNodeUnavailable)
			},
		}
		testAddressGroup(t, facade, "/address/addr?withKeys=true", "GET", nil, http.StatusBadRequest,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, fmt.Errorf("%w: pruned", common.ErrIteratorNodeUnavailable)))
	})
	t.Run("with keys should work through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		pairs := map[string]string{"k1": "v1"}
		facade := &mock.FacadeStub{
			GetAccountWithKeysCalled: func(_ context.Context, address string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (api.AccountResponse, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.WithKeys)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return api.AccountResponse{Address: address, Pairs: pairs}, nextCheckpoint, api.BlockInfo{}, nil
			},
			GetAccountCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				assert.Fail(t, "the unpaginated account loader must not be used with keys")
				return api.AccountResponse{}, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &accountWithKeysResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr?withKeys=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Equal(t, "addr", response.Data.Account.Address)
		assert.Equal(t, pairs, response.Data.Account.Pairs)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		body, _ := json.Marshal(groups.PagedRequestBody{IteratorState: givenCheckpoint})
		response = &accountWithKeysResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr?withKeys=true&numKeys=7", "POST", bytes.NewBuffer(body), response)
		assert.Equal(t, pairs, response.Data.Account.Pairs)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})
	t.Run("without keys, no iterator state is returned", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(_ context.Context, address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				assert.False(t, options.WithKeys)
				return api.AccountResponse{Address: address}, api.BlockInfo{}, nil
			},
		}
		response := &accountWithKeysResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr?withKeys=false", "GET", nil, response)
		assert.Equal(t, "addr", response.Data.Account.Address)
		assert.Nil(t, response.Data.IteratorState)
	})
}

func TestAddressGroup_getBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//balance", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/balance?blockNonce=not-uint64", "GET", nil,
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
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		amount := big.NewInt(10)
		addr := "testAddress"
		facade := &mock.FacadeStub{
			GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
				return amount, api.BlockInfo{}, nil
			},
		}

		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			fmt.Sprintf("/address/%s/balance", addr),
			"GET",
			nil,
			response,
		)

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
		testErrorScenario("/address/bulk?blockNonce=not-uint64", "POST", bytes.NewBuffer([]byte(`["erd1", "erd1"]`)),
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
		facade := &mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return expectedAccounts, api.BlockInfo{}, nil
			},
		}

		type responseType struct {
			Data struct {
				Accounts map[string]*api.AccountResponse `json:"accounts"`
			} `json:"data"`
			Error string            `json:"error"`
			Code  shared.ReturnCode `json:"code"`
		}
		response := &responseType{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/bulk",
			"POST",
			bytes.NewBuffer([]byte(`["erd1", "erd1"]`)),
			response,
		)

		require.Empty(t, response.Error)
		require.Equal(t, shared.ReturnCodeSuccess, response.Code)
		require.Equal(t, expectedAccounts, response.Data.Accounts)
	})
}

func TestAddressGroup_getUsername(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//username", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetUsername, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/username?blockNonce=not-uint64", "GET", nil,
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
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetUsername, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testUsername := "provided username"
		facade := &mock.FacadeStub{
			GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testUsername, api.BlockInfo{}, nil
			},
		}

		usernameResponseObj := &usernameResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/username",
			"GET",
			nil,
			usernameResponseObj,
		)
		assert.Equal(t, testUsername, usernameResponseObj.Data.Username)
	})
}

func TestAddressGroup_getCodeHash(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//code-hash", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetCodeHash, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/code-hash?blockNonce=not-uint64", "GET", nil,
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
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetCodeHash, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testCodeHash := []byte("value")
		expectedResponseCodeHash := base64.StdEncoding.EncodeToString(testCodeHash)
		facade := &mock.FacadeStub{
			GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
				return testCodeHash, api.BlockInfo{}, nil
			},
		}

		codeHashResponseObj := &codeHashResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/code-hash",
			"GET",
			nil,
			codeHashResponseObj,
		)
		assert.Equal(t, expectedResponseCodeHash, codeHashResponseObj.Data.CodeHash)
	})
}

func TestAddressGroup_getValueForKey(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//key/test", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetValueForKey, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/key/test?blockNonce=not-uint64", "GET", nil,
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
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetValueForKey, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue := "value"
		facade := &mock.FacadeStub{
			GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testValue, api.BlockInfo{}, nil
			},
		}

		valueForKeyResponseObj := &valueForKeyResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/key/test",
			"GET",
			nil,
			valueForKeyResponseObj,
		)
		assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
	})
}

func TestAddressGroup_getGuardianData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//guardian-data", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetGuardianData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/guardian-data?blockNonce=not-uint64", "GET", nil,
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
			"/address/erd1alice/guardian-data",
			"GET",
			nil,
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
		facade := &mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return expectedGuardianData, api.BlockInfo{}, nil
			},
		}

		response := &guardianDataResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/guardian-data",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedGuardianData, response.Data.GuardianData)
	})
}

func TestAddressGroup_getKeyValuePairs(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//keys", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/keys?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrBadUrlParams)))
	t.Run("invalid iterator state should error",
		testErrorScenario("/address/erd1alice/keys?iteratorState=!!!", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrBadUrlParams)))
	t.Run("invalid body should error",
		testErrorScenario("/address/erd1alice/keys", "POST", strings.NewReader("not json"),
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrValidation)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/keys",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, expectedErr),
		)
	})
	t.Run("unusable iterator state should be a bad request", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, fmt.Errorf("%w: conflict", common.ErrInvalidIteratorState)
			},
		}
		testAddressGroup(t, facade, "/address/erd1alice/keys", "GET", nil, http.StatusBadRequest,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, fmt.Errorf("%w: conflict", common.ErrInvalidIteratorState)))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pairs := map[string]string{
			"k1": "v1",
			"k2": "v2",
		}
		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (map[string]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.Equal(t, api.AccountQueryOptions{}, options)
				assert.Zero(t, numKeys)
				assert.Nil(t, iteratorState)
				return pairs, nil, api.BlockInfo{}, nil
			},
		}

		response := &keyValuePairsResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/keys",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, pairs, response.Data.Pairs)
		assert.Nil(t, response.Data.IteratorState)
	})
	t.Run("should page through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		pairs := map[string]string{"k1": "v1"}
		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (map[string]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.OnFinalBlock)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return pairs, nextCheckpoint, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &keyValuePairsResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/keys?onFinalBlock=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Equal(t, pairs, response.Data.Pairs)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		numKeys := uint(7)
		body, _ := json.Marshal(groups.PagedRequestBody{NumKeys: &numKeys, IteratorState: givenCheckpoint})
		response = &keyValuePairsResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/keys?onFinalBlock=true", "POST", bytes.NewBuffer(body), response)
		assert.Equal(t, pairs, response.Data.Pairs)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})
}

func TestAddressGroup_iterateKeys(t *testing.T) {
	t.Parallel()

	t.Run("invalid body should error",
		testErrorScenario("/address/iterate-keys", "POST", bytes.NewBuffer([]byte("invalid body")),
			formatExpectedErr(apiErrors.ErrValidation, errors.New("invalid character 'i' looking for beginning of value"))))
	t.Run("empty address should error", func(t *testing.T) {
		t.Parallel()

		body := &groups.IterateKeysRequest{
			Address: "",
		}
		bodyBytes, _ := json.Marshal(body)
		testAddressGroup(
			t,
			&mock.FacadeStub{},
			"/address/iterate-keys",
			"POST",
			bytes.NewBuffer(bodyBytes),
			http.StatusBadRequest,
			formatExpectedErr(apiErrors.ErrValidation, apiErrors.ErrEmptyAddress),
		)
	})
	t.Run("invalid query options should error", func(t *testing.T) {
		t.Parallel()

		body := &groups.IterateKeysRequest{
			Address: "erd1",
		}
		bodyBytes, _ := json.Marshal(body)
		testAddressGroup(
			t,
			&mock.FacadeStub{},
			"/address/iterate-keys?blockNonce=not-uint64",
			"POST",
			bytes.NewBuffer(bodyBytes),
			http.StatusBadRequest,
			formatExpectedErr(apiErrors.ErrIterateKeys, apiErrors.ErrBadUrlParams),
		)
	})
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		body := &groups.IterateKeysRequest{
			Address: "erd1",
		}
		bodyBytes, _ := json.Marshal(body)
		facade := &mock.FacadeStub{
			IterateKeysCalled: func(_ context.Context, address string, numKeys uint, iteratorState [][]byte, options api.AccountQueryOptions) (map[string]string, [][]byte, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/iterate-keys",
			"POST",
			bytes.NewBuffer(bodyBytes),
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrIterateKeys, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pairs := map[string]string{
			"k1": "v1",
			"k2": "v2",
		}

		body := &groups.IterateKeysRequest{
			Address:       "erd1",
			NumKeys:       10,
			IteratorState: [][]byte{[]byte("starting"), []byte("state")},
		}
		newIteratorState := [][]byte{[]byte("new"), []byte("state")}
		bodyBytes, _ := json.Marshal(body)
		facade := &mock.FacadeStub{
			IterateKeysCalled: func(_ context.Context, address string, numKeys uint, iteratorState [][]byte, options api.AccountQueryOptions) (map[string]string, [][]byte, api.BlockInfo, error) {
				assert.Equal(t, body.Address, address)
				assert.Equal(t, body.NumKeys, numKeys)
				assert.Equal(t, body.IteratorState, iteratorState)
				return pairs, newIteratorState, api.BlockInfo{}, nil
			},
		}

		response := &iterateKeysResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/iterate-keys",
			"POST",
			bytes.NewBuffer(bodyBytes),
			response,
		)
		assert.Equal(t, pairs, response.Data.Pairs)
		assert.Equal(t, newIteratorState, response.Data.NewIteratorState)
	})
}

func TestAddressGroup_getESDTBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdt/newToken", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdt/newToken?blockNonce=not-uint64", "GET", nil,
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
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue := big.NewInt(100).String()
		testProperties := []byte{byte(0), byte(1), byte(0)}
		facade := &mock.FacadeStub{
			GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
				return &esdt.ESDigitalToken{Value: big.NewInt(100), Properties: testProperties}, api.BlockInfo{}, nil
			},
		}

		esdtBalanceResponseObj := &esdtTokenResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/esdt/newToken",
			"GET",
			nil,
			esdtBalanceResponseObj,
		)
		assert.Equal(t, testValue, esdtBalanceResponseObj.Data.Balance)
		assert.Equal(t, "000100", esdtBalanceResponseObj.Data.Properties)
	})
}

func TestAddressGroup_getESDTsRoles(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdts/roles", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdts/roles?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTsRolesCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string][]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdts/roles",
			"GET",
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
		facade := &mock.FacadeStub{
			GetESDTsRolesCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string][]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return roles, nil, api.BlockInfo{}, nil
			},
		}

		response := &esdtRolesResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/esdts/roles",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, roles, response.Data.Roles)
	})
	t.Run("should page through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		facade := &mock.FacadeStub{
			GetESDTsRolesCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (map[string][]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.OnFinalBlock)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return map[string][]string{"token0": {"role0"}}, nextCheckpoint, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &esdtRolesResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdts/roles?onFinalBlock=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Equal(t, map[string][]string{"token0": {"role0"}}, response.Data.Roles)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		numKeys := uint(7)
		body, _ := json.Marshal(groups.PagedRequestBody{NumKeys: &numKeys, IteratorState: givenCheckpoint})
		response = &esdtRolesResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdts/roles?onFinalBlock=true", "POST", bytes.NewBuffer(body), response)
		assert.Equal(t, map[string][]string{"token0": {"role0"}}, response.Data.Roles)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})

}

func TestAddressGroup_getESDTTokensWithRole(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdts-with-role/ESDTRoleNFTCreate", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, apiErrors.ErrBadUrlParams)))
	t.Run("invalid role should error",
		testErrorScenario("/address/erd1alice/esdts-with-role/invalid", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, fmt.Errorf("%w: invalid role: %s", apiErrors.ErrValidation, "invalid"))))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetESDTsWithRoleCalled: func(_ context.Context, _ string, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTTokensWithRole, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := &mock.FacadeStub{
			GetESDTsWithRoleCalled: func(_ context.Context, address string, role string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return expectedTokens, nil, api.BlockInfo{}, nil
			},
		}

		esdtResponseObj := &esdtsWithRoleResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate",
			"GET",
			nil,
			esdtResponseObj,
		)
		assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
	})
	t.Run("should page through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		facade := &mock.FacadeStub{
			GetESDTsWithRoleCalled: func(_ context.Context, _ string, role string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.Equal(t, "ESDTRoleNFTCreate", role)
				assert.True(t, options.OnFinalBlock)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return []string{"ABC-0o9i8u"}, nextCheckpoint, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &esdtsWithRoleResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate?onFinalBlock=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Equal(t, []string{"ABC-0o9i8u"}, response.Data.Tokens)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		numKeys := uint(7)
		body, _ := json.Marshal(groups.PagedRequestBody{NumKeys: &numKeys, IteratorState: givenCheckpoint})
		response = &esdtsWithRoleResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdts-with-role/ESDTRoleNFTCreate?onFinalBlock=true", "POST", bytes.NewBuffer(body), response)
		assert.Equal(t, []string{"ABC-0o9i8u"}, response.Data.Tokens)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})

}

func TestAddressGroup_getNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//registered-nfts", "GET", nil,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/registered-nfts?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/registered-nfts",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(_ context.Context, address string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return expectedTokens, nil, api.BlockInfo{}, nil
			},
		}

		esdtResponseObj := &esdtsWithRoleResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/registered-nfts",
			"GET",
			nil,
			esdtResponseObj,
		)
		assert.Equal(t, expectedTokens, esdtResponseObj.Data.Tokens)
	})
	t.Run("should page through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) ([]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.OnFinalBlock)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return []string{"ABC-0o9i8u"}, nextCheckpoint, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &esdtsWithRoleResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/registered-nfts?onFinalBlock=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Equal(t, []string{"ABC-0o9i8u"}, response.Data.Tokens)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		numKeys := uint(7)
		body, _ := json.Marshal(groups.PagedRequestBody{NumKeys: &numKeys, IteratorState: givenCheckpoint})
		response = &esdtsWithRoleResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/registered-nfts?onFinalBlock=true", "POST", bytes.NewBuffer(body), response)
		assert.Equal(t, []string{"ABC-0o9i8u"}, response.Data.Tokens)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})

}

func TestAddressGroup_getESDTNFTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//nft/newToken/nonce/10", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/nft/newToken/nonce/10?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("invalid nonce should error",
		testErrorScenario("/address/erd1alice/nft/newToken/nonce/not-int", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrNonceInvalid)))
	t.Run("nonce too long should error",
		testErrorScenario("/address/erd1alice/nft/newToken/nonce/1234567489123456789123", "GET", nil,
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
			"GET",
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
		facade := &mock.FacadeStub{
			GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
				return &esdt.ESDigitalToken{
					Value:         big.NewInt(100),
					Properties:    testProperties,
					TokenMetaData: &esdt.MetaData{Nonce: testNonce, Creator: []byte(testAddress)}}, api.BlockInfo{}, nil
			},
		}

		esdtResponseObj := &esdtNFTResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/nft/newToken/nonce/10",
			"GET",
			nil,
			esdtResponseObj,
		)
		assert.Equal(t, testValue, esdtResponseObj.Data.Balance)
		assert.Equal(t, "010000", esdtResponseObj.Data.Properties)
		assert.Equal(t, testAddress, esdtResponseObj.Data.Creator)
		assert.Equal(t, testNonce, esdtResponseObj.Data.Nonce)
	})
}

func TestAddressGroup_getAllESDTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//esdt", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/erd1alice/esdt?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("invalid iterator state should error",
		testErrorScenario("/address/erd1alice/esdt?iteratorState=!!!", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("invalid body should error",
		testErrorScenario("/address/erd1alice/esdt", "POST", strings.NewReader("not json"),
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, apiErrors.ErrValidation)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAllESDTTokensCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string]*esdt.ESDigitalToken, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/erd1alice/esdt",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, expectedErr),
		)
	})
	t.Run("unusable iterator state should be a bad request", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAllESDTTokensCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string]*esdt.ESDigitalToken, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, fmt.Errorf("%w: pruned", common.ErrIteratorNodeUnavailable)
			},
		}
		testAddressGroup(t, facade, "/address/erd1alice/esdt", "GET", nil, http.StatusBadRequest,
			formatExpectedErr(apiErrors.ErrGetESDTNFTData, fmt.Errorf("%w: pruned", common.ErrIteratorNodeUnavailable)))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue1 := "token1"
		testValue2 := "token2"
		facade := &mock.FacadeStub{
			GetAllESDTTokensCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (map[string]*esdt.ESDigitalToken, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.Equal(t, api.AccountQueryOptions{}, options)
				assert.Zero(t, numKeys)
				assert.Nil(t, iteratorState)
				tokens := make(map[string]*esdt.ESDigitalToken)
				tokens[testValue1] = &esdt.ESDigitalToken{Value: big.NewInt(10)}
				tokens[testValue2] = &esdt.ESDigitalToken{Value: big.NewInt(100)}
				return tokens, nil, api.BlockInfo{}, nil
			},
		}

		esdtTokenResponseObj := &esdtTokensCompleteResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/erd1alice/esdt",
			"GET",
			nil,
			esdtTokenResponseObj,
		)
		assert.Equal(t, 2, len(esdtTokenResponseObj.Data.Tokens))
		assert.Nil(t, esdtTokenResponseObj.Data.IteratorState)
	})
	t.Run("should page through GET and POST, returning the iterator state", func(t *testing.T) {
		t.Parallel()

		givenCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("s")}}
		nextCheckpoint := &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", IteratorState: [][]byte{[]byte("t")}}
		facade := &mock.FacadeStub{
			GetAllESDTTokensCalled: func(_ context.Context, _ string, options api.AccountQueryOptions, numKeys uint, iteratorState *checkpoint.Checkpoint) (map[string]*esdt.ESDigitalToken, *checkpoint.Checkpoint, api.BlockInfo, error) {
				assert.True(t, options.OnFinalBlock)
				assert.Equal(t, uint(7), numKeys)
				assert.Equal(t, givenCheckpoint, iteratorState)
				return map[string]*esdt.ESDigitalToken{"token1": {Value: big.NewInt(10)}}, nextCheckpoint, api.BlockInfo{}, nil
			},
		}

		encodedCheckpoint, err := checkpoint.Encode(givenCheckpoint)
		require.Nil(t, err)
		response := &esdtTokensCompleteResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdt?onFinalBlock=true&numKeys=7&iteratorState="+encodedCheckpoint, "GET", nil, response)
		assert.Len(t, response.Data.Tokens, 1)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)

		numKeys := uint(7)
		body, _ := json.Marshal(groups.PagedRequestBody{NumKeys: &numKeys, IteratorState: givenCheckpoint})
		response = &esdtTokensCompleteResponse{}
		loadAddressGroupResponse(t, facade, "/address/erd1alice/esdt?onFinalBlock=true", "POST", bytes.NewBuffer(body), response)
		assert.Len(t, response.Data.Tokens, 1)
		assert.Equal(t, nextCheckpoint, response.Data.IteratorState)
	})
}

func TestAddressGroup_UpdateFacade(t *testing.T) {
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
			GetESDTsRolesCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string][]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return roles, nil, api.BlockInfo{}, nil
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
		newFacade := mock.FacadeStub{
			GetESDTsRolesCalled: func(_ context.Context, _ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint) (map[string][]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
				return nil, nil, api.BlockInfo{}, newErr
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
	})
}

func TestAddressGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	addrGroup, _ := groups.NewAddressGroup(nil)
	require.True(t, addrGroup.IsInterfaceNil())

	addrGroup, _ = groups.NewAddressGroup(&mock.FacadeStub{})
	require.False(t, addrGroup.IsInterfaceNil())
}

func testErrorScenario(url string, method string, body io.Reader, expectedErr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testAddressGroup(
			t,
			&mock.FacadeStub{},
			url,
			method,
			body,
			http.StatusBadRequest,
			expectedErr,
		)
	}
}

func loadAddressGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	addrGroup, err := groups.NewAddressGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func testAddressGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	expectedRespCode int,
	expectedRespError string,
) {
	addrGroup, err := groups.NewAddressGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
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
					{Name: "/iterate-keys", Open: true},
					{Name: "/:address/key/:key", Open: true},
					{Name: "/:address/esdt", Open: true},
					{Name: "/:address/esdts/roles", Open: true},
					{Name: "/:address/esdt/:tokenIdentifier", Open: true},
					{Name: "/:address/nft/:tokenIdentifier/nonce/:nonce", Open: true},
					{Name: "/:address/esdts-with-role/:role", Open: true},
					{Name: "/:address/registered-nfts", Open: true},
					{Name: "/:address/is-data-trie-migrated", Open: true},
				},
			},
		},
	}
}

func TestIsDataTrieMigrated(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")

	t.Run("should return error if IsDataTrieMigrated returns error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return false, expectedErr
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should return true if IsDataTrieMigrated returns true", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return true, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.True(t, response.Error == "")

		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.True(t, respData["isMigrated"].(bool))
	})

	t.Run("should return false if IsDataTrieMigrated returns false", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return false, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.True(t, response.Error == "")

		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.False(t, respData["isMigrated"].(bool))
	})
}
