package proof_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/proof"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetProof_NilContextShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/proof/root-hash/roothash/address/addr", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetProof_GetProofError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("GetProof error")
	facade := &mock.Facade{
		GetProofCalled: func(rootHash []byte, address []byte) ([][]byte, error) {
			return nil, err
		},
	}

	ws := startNodeServer(facade)
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
	facade := &mock.Facade{
		GetProofCalled: func(rootHash []byte, address []byte) ([][]byte, error) {
			assert.Equal(t, []byte("roothash"), rootHash)
			assert.Equal(t, []byte("addr"), address)
			return validProof, nil
		},
	}

	ws := startNodeServer(facade)
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

	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("valid")), proof1)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("proof")), proof2)
}

func TestVerifyProof_NilContextShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(nil)

	req, _ := http.NewRequest("POST", "/proof/verify", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestVerifyProof_BadRequestShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServer(&mock.Facade{})

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

	err := fmt.Errorf("VerifyProof err")
	facade := &mock.Facade{
		VerifyProofCalled: func(rootHash []byte, address []byte, proof [][]byte) (bool, error) {
			return false, err
		},
	}

	varifyProofParams := proof.VerifyProofRequest{
		RootHash: []byte("rootHash"),
		Address:  []byte("address"),
		Proof:    [][]byte{[]byte("valid"), []byte("proof")},
	}
	verifyProofBytes, _ := json.Marshal(varifyProofParams)

	ws := startNodeServer(facade)
	req, _ := http.NewRequest("POST", "/proof/verify", bytes.NewBuffer(verifyProofBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrVerifyProof.Error()))
}

func TestVerifyProof(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	address := []byte("address")
	validProof := [][]byte{[]byte("valid"), []byte("proof")}
	varifyProofParams := proof.VerifyProofRequest{
		RootHash: rootHash,
		Address:  address,
		Proof:    validProof,
	}
	verifyProofBytes, _ := json.Marshal(varifyProofParams)

	facade := &mock.Facade{
		VerifyProofCalled: func(rH []byte, addr []byte, proof [][]byte) (bool, error) {
			assert.Equal(t, rootHash, rH)
			assert.Equal(t, address, addr)
			assert.Equal(t, validProof, proof)
			return true, nil
		},
	}

	ws := startNodeServer(facade)
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

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler transaction.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ginProofRoute := ws.Group("/proof")
	if handler != nil {
		ginProofRoute.Use(middleware.WithFacade(handler))
	}
	proofRoute, _ := wrapper.NewRouterWrapper("proof", ginProofRoute, getRoutesConfig())
	proof.Routes(proofRoute)
	return ws
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"proof": {
				Routes: []config.RouteConfig{
					{Name: "/root-hash/:roothash/address/:address", Open: true},
					{Name: "/verify", Open: true},
				},
			},
		},
	}
}
