package groups

import (
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getAccountPath            = "/:address"
	getBalancePath            = "/:address/balance"
	getUsernamePath           = "/:address/username"
	getKeysPath               = "/:address/keys"
	getKeyPath                = "/:address/key/:key"
	getESDTTokensPath         = "/:address/esdt"
	getESDTBalancePath        = "/:address/esdt/:tokenIdentifier"
	getESDTTokensWithRolePath = "/:address/esdts-with-role/:role"
	getESDTsRolesPath         = "/:address/esdts/roles"
	getRegisteredNFTsPath     = "/:address/registered-nfts"
	getESDTNFTDataPath        = "/:address/nft/:tokenIdentifier/nonce/:nonce"
)

type addressFacadeHandler interface {
	GetBalance(address string) (*big.Int, error)
	GetUsername(address string) (string, error)
	GetValueForKey(address string, key string) (string, error)
	GetAccount(address string) (api.AccountResponse, error)
	GetESDTData(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error)
	GetESDTsRoles(address string) (map[string][]string, error)
	GetNFTTokenIDsRegisteredByAddress(address string) ([]string, error)
	GetESDTsWithRole(address string, role string) ([]string, error)
	GetAllESDTTokens(address string) (map[string]*esdt.ESDigitalToken, error)
	GetKeyValuePairs(address string) (map[string]string, error)
	IsInterfaceNil() bool
}

type addressGroup struct {
	*baseGroup
	facade    addressFacadeHandler
	mutFacade sync.RWMutex
}

type esdtTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

type esdtNFTTokenData struct {
	TokenIdentifier string   `json:"tokenIdentifier"`
	Balance         string   `json:"balance"`
	Properties      string   `json:"properties,omitempty"`
	Name            string   `json:"name,omitempty"`
	Nonce           uint64   `json:"nonce,omitempty"`
	Creator         string   `json:"creator,omitempty"`
	Royalties       string   `json:"royalties,omitempty"`
	Hash            []byte   `json:"hash,omitempty"`
	URIs            [][]byte `json:"uris,omitempty"`
	Attributes      []byte   `json:"attributes,omitempty"`
}

// NewAddressGroup returns a new instance of addressGroup
func NewAddressGroup(facade addressFacadeHandler) (*addressGroup, error) {
	if facade == nil {
		return nil, fmt.Errorf("%w for address group", errors.ErrNilFacadeHandler)
	}

	ag := &addressGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getAccountPath,
			Method:  http.MethodGet,
			Handler: ag.getAccount,
		},
		{
			Path:    getBalancePath,
			Method:  http.MethodGet,
			Handler: ag.getBalance,
		},
		{
			Path:    getUsernamePath,
			Method:  http.MethodGet,
			Handler: ag.getUsername,
		},
		{
			Path:    getKeyPath,
			Method:  http.MethodGet,
			Handler: ag.getValueForKey,
		},
		{
			Path:    getKeysPath,
			Method:  http.MethodGet,
			Handler: ag.getKeyValuePairs,
		},
		{
			Path:    getESDTBalancePath,
			Method:  http.MethodGet,
			Handler: ag.getESDTBalance,
		},
		{
			Path:    getESDTNFTDataPath,
			Method:  http.MethodGet,
			Handler: ag.getESDTNFTData,
		},
		{
			Path:    getESDTTokensPath,
			Method:  http.MethodGet,
			Handler: ag.getAllESDTData,
		},
		{
			Path:    getRegisteredNFTsPath,
			Method:  http.MethodGet,
			Handler: ag.getNFTTokenIDsRegisteredByAddress,
		},
		{
			Path:    getESDTTokensWithRolePath,
			Method:  http.MethodGet,
			Handler: ag.getESDTTokensWithRole,
		},
		{
			Path:    getESDTsRolesPath,
			Method:  http.MethodGet,
			Handler: ag.getESDTsRoles,
		},
	}
	ag.endpoints = endpoints

	return ag, nil
}

// addressGroup returns a response containing information about the account correlated with provided address
func (ag *addressGroup) getAccount(c *gin.Context) {
	addr := c.Param("address")
	accountResponse, err := ag.getFacade().GetAccount(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrCouldNotGetAccount.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	accountResponse.Address = addr

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"account": accountResponse},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getBalance returns the balance for the address parameter
func (ag *addressGroup) getBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	balance, err := ag.getFacade().GetBalance(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"balance": balance.String()},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getUsername returns the username for the address parameter
func (ag *addressGroup) getUsername(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetUsername.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	userName, err := ag.getFacade().GetUsername(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetUsername.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"username": userName},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getValueForKey returns the value for the given address and key
func (ag *addressGroup) getValueForKey(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyAddress.Error()),
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
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyKey.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	value, err := ag.getFacade().GetValueForKey(addr, key)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"value": value},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// addressGroup returns all the key-value pairs for the given address
func (ag *addressGroup) getKeyValuePairs(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetKeyValuePairs.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	value, err := ag.getFacade().GetKeyValuePairs(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetKeyValuePairs.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"pairs": value},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getESDTBalance returns the balance for the given address and esdt token
func (ag *addressGroup) getESDTBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyTokenIdentifier.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	esdtData, err := ag.getFacade().GetESDTData(addr, tokenIdentifier, 0)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	tokenData := esdtTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         esdtData.Value.String(),
		Properties:      string(esdtData.Properties),
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokenData": tokenData},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getESDTsRoles returns the token identifiers and roles for a given address
func (ag *addressGroup) getESDTsRoles(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetRolesForAccount.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokensRoles, err := ag.getFacade().GetESDTsRoles(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetRolesForAccount.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"roles": tokensRoles},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getESDTTokensWithRole returns the token identifiers where a given address has the given role
func (ag *addressGroup) getESDTTokensWithRole(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	role := c.Param("role")
	if role == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyRole.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	if !core.IsValidESDTRole(role) {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("invalid role: %s", role),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokens, err := ag.getFacade().GetESDTsWithRole(addr, role)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokens": tokens},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getNFTTokenIDsRegisteredByAddress returns the token identifiers of the tokens where a given address is the owner
func (ag *addressGroup) getNFTTokenIDsRegisteredByAddress(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokens, err := ag.getFacade().GetNFTTokenIDsRegisteredByAddress(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokens": tokens},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getESDTNFTData returns the nft data for the given token
func (ag *addressGroup) getESDTNFTData(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTNFTData.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTNFTData.Error(), errors.ErrEmptyTokenIdentifier.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	nonceAsStr := c.Param("nonce")
	if nonceAsStr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNonceInvalid.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	nonceAsBigInt, okConvert := big.NewInt(0).SetString(nonceAsStr, 10)
	if !okConvert {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNonceInvalid.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	esdtData, err := ag.getFacade().GetESDTData(addr, tokenIdentifier, nonceAsBigInt.Uint64())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	tokenData := buildTokenDataApiResponse(tokenIdentifier, esdtData)

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokenData": tokenData},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getAllESDTData returns the tokens list from this account
func (ag *addressGroup) getAllESDTData(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTTokens.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokens, err := ag.getFacade().GetAllESDTTokens(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTTokens.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	formattedTokens := make(map[string]*esdtNFTTokenData)
	for tokenID, esdtData := range tokens {
		tokenData := buildTokenDataApiResponse(tokenID, esdtData)

		formattedTokens[tokenID] = tokenData
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"esdts": formattedTokens},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func buildTokenDataApiResponse(tokenIdentifier string, esdtData *esdt.ESDigitalToken) *esdtNFTTokenData {
	tokenData := &esdtNFTTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         esdtData.Value.String(),
		Properties:      string(esdtData.Properties),
	}
	if esdtData.TokenMetaData != nil {
		tokenData.Name = string(esdtData.TokenMetaData.Name)
		tokenData.Nonce = esdtData.TokenMetaData.Nonce
		tokenData.Creator = string(esdtData.TokenMetaData.Creator)
		tokenData.Royalties = big.NewInt(int64(esdtData.TokenMetaData.Royalties)).String()
		tokenData.Hash = esdtData.TokenMetaData.Hash
		tokenData.URIs = esdtData.TokenMetaData.URIs
		tokenData.Attributes = esdtData.TokenMetaData.Attributes
	}

	return tokenData
}

func (ag *addressGroup) getFacade() addressFacadeHandler {
	ag.mutFacade.RLock()
	defer ag.mutFacade.RUnlock()

	return ag.facade
}

// UpdateFacade will update the facade
func (ag *addressGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(addressFacadeHandler)
	if !ok {
		return fmt.Errorf("%w for address group", errors.ErrFacadeWrongTypeAssertion)
	}

	ag.mutFacade.Lock()
	ag.facade = castFacade
	ag.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ag *addressGroup) IsInterfaceNil() bool {
	return ag == nil
}
