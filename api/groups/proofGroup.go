package groups

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/gin-gonic/gin"
)

const (
	getProofCurrentRootHashEndpoint = "/proof/address/:address"
	getProofEndpoint                = "/proof/root-hash/:roothash/address/:address"
	getProofDataTrieEndpoint        = "/proof/root-hash/:roothash/address/:address/key/:key"
	verifyProofEndpoint             = "/proof/verify"
	getProofCurrentRootHashPath     = "/address/:address"
	getProofPath                    = "/root-hash/:roothash/address/:address"
	getProofDataTriePath            = "/root-hash/:roothash/address/:address/key/:key"
	verifyProofPath                 = "/verify"
)

// proofFacadeHandler defines the methods to be implemented by a facade for proof requests
type proofFacadeHandler interface {
	GetProof(rootHash string, address string) (*common.GetProofResponse, error)
	GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error)
	GetProofCurrentRootHash(address string) (*common.GetProofResponse, error)
	VerifyProof(rootHash string, address string, proof [][]byte) (bool, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
	IsInterfaceNil() bool
}

type proofGroup struct {
	*baseGroup
	facade    proofFacadeHandler
	mutFacade sync.RWMutex
}

// NewProofGroup returns a new instance of proofGroup
func NewProofGroup(facade proofFacadeHandler) (*proofGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for proof group", errors.ErrNilFacadeHandler)
	}

	pg := &proofGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getProofPath,
			Method:  http.MethodGet,
			Handler: pg.getProof,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(getProofEndpoint, facade),
					Position:   shared.Before,
				},
			},
		},
		{
			Path:    getProofDataTriePath,
			Method:  http.MethodGet,
			Handler: pg.getProofDataTrie,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(getProofDataTrieEndpoint, facade),
					Position:   shared.Before,
				},
			},
		},
		{
			Path:    getProofCurrentRootHashPath,
			Method:  http.MethodGet,
			Handler: pg.getProofCurrentRootHash,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(getProofCurrentRootHashEndpoint, facade),
					Position:   shared.Before,
				},
			},
		},
		{
			Path:    verifyProofPath,
			Method:  http.MethodPost,
			Handler: pg.verifyProof,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(verifyProofEndpoint, facade),
					Position:   shared.Before,
				},
			},
		},
	}
	pg.endpoints = endpoints

	return pg, nil
}

// VerifyProofRequest represents the parameters needed to verify a Merkle proof
type VerifyProofRequest struct {
	RootHash string   `json:"roothash"`
	Address  string   `json:"address"`
	Proof    []string `json:"proof"`
}

// getProof will receive a rootHash and an address from the client, and it will return the Merkle proof
func (pg *proofGroup) getProof(c *gin.Context) {
	rootHash := c.Param("roothash")
	if rootHash == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyRootHash.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	address := c.Param("address")
	if address == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	response, err := pg.getFacade().GetProof(rootHash, address)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetProof.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	hexProof := bytesToHex(response.Proof)

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"proof": hexProof,
				"value": hex.EncodeToString(response.Value),
			},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getProofDataTrie will receive a rootHash, a key and an address from the client, and it will return the Merkle proofs
// for the address and key
func (pg *proofGroup) getProofDataTrie(c *gin.Context) {
	rootHash := c.Param("roothash")
	if rootHash == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyRootHash.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	address := c.Param("address")
	if address == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyKey.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	mainTrieResponse, dataTrieResponse, err := pg.getFacade().GetProofDataTrie(rootHash, address, key)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetProof.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	proofs := make(map[string]interface{})
	proofs["mainProof"] = bytesToHex(mainTrieResponse.Proof)
	proofs["dataTrieProof"] = bytesToHex(dataTrieResponse.Proof)

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"proofs":           proofs,
				"value":            hex.EncodeToString(dataTrieResponse.Value),
				"dataTrieRootHash": dataTrieResponse.RootHash,
			},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func bytesToHex(bytesValue [][]byte) []string {
	hexValue := make([]string, 0)
	for _, byteValue := range bytesValue {
		hexValue = append(hexValue, hex.EncodeToString(byteValue))
	}

	return hexValue
}

// getProofCurrentRootHash will receive an address from the client, and it will return the
// Merkle proof for the current root hash
func (pg *proofGroup) getProofCurrentRootHash(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	response, err := pg.getFacade().GetProofCurrentRootHash(address)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetProof.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	hexProof := bytesToHex(response.Proof)

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"proof":    hexProof,
				"value":    hex.EncodeToString(response.Value),
				"rootHash": response.RootHash,
			},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// verifyProof will receive a rootHash, an address and a Merkle proof from the client,
// and it will verify the proof
func (pg *proofGroup) verifyProof(c *gin.Context) {
	var verifyProofParams = &VerifyProofRequest{}
	err := c.ShouldBindJSON(&verifyProofParams)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	proof := make([][]byte, 0)
	for _, hexProof := range verifyProofParams.Proof {
		bytesProof, err := hex.DecodeString(hexProof)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		proof = append(proof, bytesProof)
	}

	var proofOk bool
	proofOk, err = pg.getFacade().VerifyProof(verifyProofParams.RootHash, verifyProofParams.Address, proof)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrVerifyProof.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"ok": proofOk},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (pg *proofGroup) getFacade() proofFacadeHandler {
	pg.mutFacade.RLock()
	defer pg.mutFacade.RUnlock()

	return pg.facade
}

// UpdateFacade will update the facade
func (pg *proofGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(proofFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	pg.mutFacade.Lock()
	pg.facade = castFacade
	pg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pg *proofGroup) IsInterfaceNil() bool {
	return pg == nil
}
