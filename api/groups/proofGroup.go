package groups

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getProofCurrentRootHashEndpoint = "/proof/address/:address"
	getProofEndpoint                = "/proof/root-hash/:roothash/address/:address"
	verifyProofEndpoint             = "/proof/verify"

	getProofCurrentRootHashPath = "/address/:address"
	getProofPath                = "/root-hash/:roothash/address/:address"
	verifyProofPath             = "/verify"
)

// proofFacadeHandler defines the methods to be implemented by a facade for proof requests
type proofFacadeHandler interface {
	GetProof(rootHash string, address string) ([][]byte, error)
	GetProofCurrentRootHash(address string) ([][]byte, []byte, error)
	VerifyProof(rootHash string, address string, proof [][]byte) (bool, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
}

type proofGroup struct {
	facade    proofFacadeHandler
	mutFacade sync.RWMutex
	*baseGroup
}

// NewProofGroup returns a new instance of proofGroup
func NewProofGroup(facadeHandler interface{}) (*proofGroup, error) {
	if facadeHandler == nil {
		return nil, errors.ErrNilFacadeHandler
	}

	facade, ok := facadeHandler.(proofFacadeHandler)
	if !ok {
		return nil, fmt.Errorf("%w for proof group", errors.ErrFacadeWrongTypeAssertion)
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
					Before:     true,
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
					Before:     true,
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
					Before:     true,
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

	proof, err := pg.getFacade().GetProof(rootHash, address)
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

	hexProof := make([]string, 0)
	for _, byteProof := range proof {
		hexProof = append(hexProof, hex.EncodeToString(byteProof))
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"proof": hexProof},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
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

	proof, rootHash, err := pg.getFacade().GetProofCurrentRootHash(address)
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

	hexProof := make([]string, 0)
	for _, byteProof := range proof {
		hexProof = append(hexProof, hex.EncodeToString(byteProof))
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"proof":    hexProof,
				"rootHash": hex.EncodeToString(rootHash),
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
	castedFacade, ok := newFacade.(proofFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	pg.mutFacade.Lock()
	pg.facade = castedFacade
	pg.mutFacade.Unlock()

	return nil
}
