package groups

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
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
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyRootHash)
		return
	}

	address := c.Param("address")
	if address == "" {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyAddress)
		return
	}

	response, err := pg.getFacade().GetProof(rootHash, address)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetProof, err)
		return
	}

	hexProof := bytesToHex(response.Proof)

	shared.RespondWithSuccess(c, gin.H{"proof": hexProof, "value": hex.EncodeToString(response.Value)})
}

// getProofDataTrie will receive a rootHash, a key and an address from the client, and it will return the Merkle proofs
// for the address and key
func (pg *proofGroup) getProofDataTrie(c *gin.Context) {
	rootHash := c.Param("roothash")
	if rootHash == "" {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyRootHash)
		return
	}

	address := c.Param("address")
	if address == "" {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyAddress)
		return
	}

	key := c.Param("key")
	if key == "" {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyKey)
		return
	}

	mainTrieResponse, dataTrieResponse, err := pg.getFacade().GetProofDataTrie(rootHash, address, key)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetProof, err)
		return
	}

	proofs := make(map[string]interface{})
	proofs["mainProof"] = bytesToHex(mainTrieResponse.Proof)
	proofs["dataTrieProof"] = bytesToHex(dataTrieResponse.Proof)

	shared.RespondWithSuccess(c, gin.H{
		"proofs":           proofs,
		"value":            hex.EncodeToString(dataTrieResponse.Value),
		"dataTrieRootHash": dataTrieResponse.RootHash,
	})
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
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrValidationEmptyAddress)
		return
	}

	response, err := pg.getFacade().GetProofCurrentRootHash(address)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetProof, err)
		return
	}

	hexProof := bytesToHex(response.Proof)

	shared.RespondWithSuccess(c, gin.H{
		"proof":    hexProof,
		"value":    hex.EncodeToString(response.Value),
		"rootHash": response.RootHash,
	})
}

// verifyProof will receive a rootHash, an address and a Merkle proof from the client,
// and it will verify the proof
func (pg *proofGroup) verifyProof(c *gin.Context) {
	var verifyProofParams = &VerifyProofRequest{}
	err := c.ShouldBindJSON(&verifyProofParams)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrValidation, err)
		return
	}

	proof := make([][]byte, 0)
	for _, hexProof := range verifyProofParams.Proof {
		bytesProof, err := hex.DecodeString(hexProof)
		if err != nil {
			shared.RespondWithValidationError(c, errors.ErrValidation, err)
			return
		}

		proof = append(proof, bytesProof)
	}

	var proofOk bool
	proofOk, err = pg.getFacade().VerifyProof(verifyProofParams.RootHash, verifyProofParams.Address, proof)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrVerifyProof, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"ok": proofOk})
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
