package proof

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/gin-gonic/gin"
)

const (
	getProofCurrentRootHashEndpoint = "/proof/address/:address"
	getProofEndpoint                = "/proof/root-hash/:roothash/address/:address"
	getProofDataTrieEndpoint        = "/proof/root-hash/:roothash/address/:address/key/:key"
	verifyProofEndpoint             = "/proof/verify"

	getProofCurrentRootHashPath = "/address/:address"
	getProofPath                = "/root-hash/:roothash/address/:address"
	getProofDataTriePath        = "/root-hash/:roothash/address/:address/key/:key"
	verifyProofPath             = "/verify"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetProof(rootHash string, address string) (*shared.GetProofResponse, error)
	GetProofDataTrie(rootHash string, address string, key string) (*shared.GetProofResponse, *shared.GetProofResponse, error)
	GetProofCurrentRootHash(address string) (*shared.GetProofResponse, error)
	VerifyProof(rootHash string, address string, proof [][]byte) (bool, error)
}

// Routes defines Merkle proof related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(
		http.MethodGet,
		getProofPath,
		middleware.CreateEndpointThrottler(getProofEndpoint),
		GetProof,
	)
	router.RegisterHandler(
		http.MethodGet,
		getProofDataTriePath,
		middleware.CreateEndpointThrottler(getProofDataTrieEndpoint),
		GetProofDataTrie,
	)
	router.RegisterHandler(
		http.MethodGet,
		getProofCurrentRootHashPath,
		middleware.CreateEndpointThrottler(getProofCurrentRootHashEndpoint),
		GetProofCurrentRootHash,
	)
	router.RegisterHandler(
		http.MethodPost,
		verifyProofPath,
		middleware.CreateEndpointThrottler(verifyProofEndpoint),
		VerifyProof,
	)
}

// VerifyProofRequest represents the parameters needed to verify a Merkle proof
type VerifyProofRequest struct {
	RootHash string   `json:"roothash"`
	Address  string   `json:"address"`
	Proof    []string `json:"proof"`
}

// GetProof will receive a rootHash and an address from the client, and it will return the Merkle proof
func GetProof(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

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

	response, err := facade.GetProof(rootHash, address)
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

// GetProofDataTrie will receive a rootHash, a key and an address from the client, and it will return the Merkle proofs
// for the address and key
func GetProofDataTrie(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

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

	mainTrieResponse, dataTrieResponse, err := facade.GetProofDataTrie(rootHash, address, key)
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

	var mainProofHex, dataTrieProofHex []string
	var dataTrieRootHash, dataTrieValue string
	if mainTrieResponse != nil {
		mainProofHex = bytesToHex(mainTrieResponse.Proof)
	}
	if dataTrieResponse != nil {
		dataTrieProofHex = bytesToHex(dataTrieResponse.Proof)
		dataTrieRootHash = dataTrieResponse.RootHash
		dataTrieValue = hex.EncodeToString(dataTrieResponse.Value)
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"mainProof":        mainProofHex,
				"dataTrieProof":    dataTrieProofHex,
				"value":            dataTrieValue,
				"dataTrieRootHash": dataTrieRootHash,
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

// GetProofCurrentRootHash will receive an address from the client, and it will return the
// Merkle proof for the current root hash
func GetProofCurrentRootHash(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
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

	response, err := facade.GetProofCurrentRootHash(address)
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

// VerifyProof will receive a rootHash, an address and a Merkle proof from the client,
// and it will verify the proof
func VerifyProof(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

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

	ok, err = facade.VerifyProof(verifyProofParams.RootHash, verifyProofParams.Address, proof)
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
			Data:  gin.H{"ok": ok},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func getFacade(c *gin.Context) (FacadeHandler, bool) {
	facadeObj, ok := c.Get("facade")
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNilAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	facade, ok := facadeObj.(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	return facade, true
}
