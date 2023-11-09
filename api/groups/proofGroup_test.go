package groups_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProofGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewProofGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewProofGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestGetProof_GetProofError(t *testing.T) {
	t.Parallel()

	getProofErr := fmt.Errorf("GetProof error")
	facade := &mock.FacadeStub{
		GetProofCalled: func(rootHash string, address string) (*common.GetProofResponse, error) {
			return nil, getProofErr
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("GET", "/proof/root-hash/roothash/address/addr", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrGetProof.Error()))
}

func TestGetProof(t *testing.T) {
	t.Parallel()

	validProof := [][]byte{[]byte("valid"), []byte("proof")}
	facade := &mock.FacadeStub{
		GetProofCalled: func(rootHash string, address string) (*common.GetProofResponse, error) {
			assert.Equal(t, "roothash", rootHash)
			assert.Equal(t, "addr", address)
			return &common.GetProofResponse{
				Proof: validProof,
			}, nil
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("GET", "/proof/root-hash/roothash/address/addr", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeSuccess, response.Code)

	responseMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	proofs, ok := responseMap["proof"].([]interface{})
	assert.True(t, ok)

	proof1 := proofs[0].(string)
	proof2 := proofs[1].(string)

	assert.Equal(t, hex.EncodeToString([]byte("valid")), proof1)
	assert.Equal(t, hex.EncodeToString([]byte("proof")), proof2)
}

func TestGetProofDataTrie_GetProofError(t *testing.T) {
	t.Parallel()

	getProofDataTrieErr := fmt.Errorf("GetProofDataTrie error")
	facade := &mock.FacadeStub{
		GetProofDataTrieCalled: func(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
			return nil, nil, getProofDataTrieErr
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())
	req, _ := http.NewRequest("GET", "/proof/root-hash/roothash/address/addr/key/key", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrGetProof.Error()))
}

func TestGetProofDataTrie(t *testing.T) {
	t.Parallel()

	mainTrieProof := [][]byte{[]byte("main"), []byte("proof")}
	dataTrieProof := [][]byte{[]byte("data"), []byte("proof")}
	facade := &mock.FacadeStub{
		GetProofDataTrieCalled: func(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
			assert.Equal(t, "roothash", rootHash)
			assert.Equal(t, "addr", address)
			return &common.GetProofResponse{Proof: mainTrieProof},
				&common.GetProofResponse{Proof: dataTrieProof},
				nil
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())
	req, _ := http.NewRequest("GET", "/proof/root-hash/roothash/address/addr/key/key", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeSuccess, response.Code)

	responseMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	proofsResponseMap, ok := responseMap["proofs"].(map[string]interface{})
	assert.True(t, ok)

	proofs, ok := proofsResponseMap["mainProof"].([]interface{})
	assert.True(t, ok)

	proof1 := proofs[0].(string)
	proof2 := proofs[1].(string)

	assert.Equal(t, hex.EncodeToString([]byte("main")), proof1)
	assert.Equal(t, hex.EncodeToString([]byte("proof")), proof2)

	proofs, ok = proofsResponseMap["dataTrieProof"].([]interface{})
	assert.True(t, ok)

	proof1 = proofs[0].(string)
	proof2 = proofs[1].(string)

	assert.Equal(t, hex.EncodeToString([]byte("data")), proof1)
	assert.Equal(t, hex.EncodeToString([]byte("proof")), proof2)
}

func TestGetProofCurrentRootHash_GetProofError(t *testing.T) {
	t.Parallel()

	getProofErr := fmt.Errorf("GetProof error")
	facade := &mock.FacadeStub{
		GetProofCurrentRootHashCalled: func(address string) (*common.GetProofResponse, error) {
			return nil, getProofErr
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("GET", "/proof/address/addr", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrGetProof.Error()))
}

func TestGetProofCurrentRootHash(t *testing.T) {
	t.Parallel()

	validProof := [][]byte{[]byte("valid"), []byte("proof")}
	rootHash := "rootHash"
	facade := &mock.FacadeStub{
		GetProofCurrentRootHashCalled: func(address string) (*common.GetProofResponse, error) {
			assert.Equal(t, "addr", address)
			return &common.GetProofResponse{
				Proof:    validProof,
				RootHash: rootHash,
			}, nil
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("GET", "/proof/address/addr", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeSuccess, response.Code)

	responseMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	proofs, ok := responseMap["proof"].([]interface{})
	assert.True(t, ok)

	proof1 := proofs[0].(string)
	proof2 := proofs[1].(string)

	assert.Equal(t, hex.EncodeToString([]byte("valid")), proof1)
	assert.Equal(t, hex.EncodeToString([]byte("proof")), proof2)

	rh, ok := responseMap["rootHash"].(string)
	assert.True(t, ok)
	assert.Equal(t, rootHash, rh)
}

func TestVerifyProof_BadRequestShouldErr(t *testing.T) {
	t.Parallel()

	proofGroup, err := groups.NewProofGroup(&mock.FacadeStub{})
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer([]byte("invalid bytes")))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrValidation.Error()))
}

func TestVerifyProof_VerifyProofErr(t *testing.T) {
	t.Parallel()

	verifyProfErr := fmt.Errorf("VerifyProof err")
	facade := &mock.FacadeStub{
		VerifyProofCalled: func(rootHash string, address string, proof [][]byte) (bool, error) {
			return false, verifyProfErr
		},
	}

	verifyProofParams := groups.VerifyProofRequest{
		RootHash: "rootHash",
		Address:  "address",
		Proof:    []string{hex.EncodeToString([]byte("valid")), hex.EncodeToString([]byte("proof"))},
	}
	verifyProofBytes, _ := json.Marshal(verifyProofParams)

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrVerifyProof.Error()))
}

func TestVerifyProof_VerifyProofCanNotDecodeProof(t *testing.T) {
	t.Parallel()

	facade := &mock.FacadeStub{}

	verifyProofParams := groups.VerifyProofRequest{
		RootHash: "rootHash",
		Address:  "address",
		Proof:    []string{"invalid", "hex"},
	}
	verifyProofBytes, _ := json.Marshal(verifyProofParams)

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrValidation.Error()))
}

func TestVerifyProof(t *testing.T) {
	t.Parallel()

	rootHash := "rootHash"
	address := "address"
	validProof := []string{hex.EncodeToString([]byte("valid")), hex.EncodeToString([]byte("proof"))}
	verifyProofParams := groups.VerifyProofRequest{
		RootHash: rootHash,
		Address:  address,
		Proof:    validProof,
	}
	verifyProofBytes, _ := json.Marshal(verifyProofParams)

	facade := &mock.FacadeStub{
		VerifyProofCalled: func(rH string, addr string, proof [][]byte) (bool, error) {
			assert.Equal(t, rootHash, rH)
			assert.Equal(t, address, addr)
			for i := range proof {
				assert.Equal(t, validProof[i], hex.EncodeToString(proof[i]))
			}

			return true, nil
		},
	}

	proofGroup, err := groups.NewProofGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

	req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeSuccess, response.Code)

	responseMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	isValid, ok := responseMap["ok"].(bool)
	assert.True(t, ok)
	assert.True(t, isValid)
}

func TestProofGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		proofGroup, err := groups.NewProofGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = proofGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		proofGroup, err := groups.NewProofGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = proofGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		rootHash := "rootHash"
		address := "address"
		validProof := []string{hex.EncodeToString([]byte("valid")), hex.EncodeToString([]byte("proof"))}
		verifyProofParams := groups.VerifyProofRequest{
			RootHash: rootHash,
			Address:  address,
			Proof:    validProof,
		}
		verifyProofBytes, _ := json.Marshal(verifyProofParams)

		facade := &mock.FacadeStub{
			VerifyProofCalled: func(rH string, addr string, proof [][]byte) (bool, error) {
				return true, nil
			},
		}

		proofGroup, err := groups.NewProofGroup(facade)
		require.NoError(t, err)

		ws := startWebServer(proofGroup, "proof", getProofRoutesConfig())

		req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, shared.ReturnCodeSuccess, response.Code)
		responseMap, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		isValid, ok := responseMap["ok"].(bool)
		assert.True(t, ok)
		assert.True(t, isValid)

		verifyProfErr := fmt.Errorf("VerifyProof err")
		newFacade := &mock.FacadeStub{
			VerifyProofCalled: func(rootHash string, address string, proof [][]byte) (bool, error) {
				return false, verifyProfErr
			},
		}

		err = proofGroup.UpdateFacade(newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		loadResponse(resp.Body, &response)
		assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
		assert.True(t, strings.Contains(response.Error, apiErrors.ErrVerifyProof.Error()))
	})
}

func TestProofGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	proofGroup, _ := groups.NewProofGroup(nil)
	require.True(t, proofGroup.IsInterfaceNil())

	proofGroup, _ = groups.NewProofGroup(&mock.FacadeStub{})
	require.False(t, proofGroup.IsInterfaceNil())
}

func getProofRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"proof": {
				Routes: []config.RouteConfig{
					{Name: "/root-hash/:roothash/address/:address", Open: true},
					{Name: "/root-hash/:roothash/address/:address/key/:key", Open: true},
					{Name: "/address/:address", Open: true},
					{Name: "/verify", Open: true},
				},
			},
		},
	}
}
