package groups

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
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
	urlParamOnFinalBlock      = "onFinalBlock"
	urlParamOnStartOfEpoch    = "onStartOfEpoch"
)

// addressFacadeHandler defines the methods to be implemented by a facade for handling address requests
type addressFacadeHandler interface {
	GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)
	GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)
	GetESDTData(address string, key string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error)
	GetESDTsRoles(address string, options api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error)
	GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetESDTsWithRole(address string, role string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetAllESDTTokens(address string, options api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error)
	GetKeyValuePairs(address string, options api.AccountQueryOptions) (map[string]string, api.BlockInfo, error)
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
	if check.IfNil(facade) {
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
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrCouldNotGetAccount, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrCouldNotGetAccount, errors.ErrBadUrlParams)
		return
	}

	accountResponse, blockInfo, err := ag.getFacade().GetAccount(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	accountResponse.Address = addr
	shared.RespondWithSuccess(c, gin.H{"account": accountResponse, "blockInfo": blockInfo})
}

// getBalance returns the balance for the address parameter
func (ag *addressGroup) getBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBalance, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBalance, errors.ErrBadUrlParams)
		return
	}

	balance, blockInfo, err := ag.getFacade().GetBalance(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBalance, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"balance": balance.String(), "blockInfo": blockInfo})
}

// getUsername returns the username for the address parameter
func (ag *addressGroup) getUsername(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetUsername, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetUsername, errors.ErrBadUrlParams)
		return
	}

	userName, blockInfo, err := ag.getFacade().GetUsername(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetUsername, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"username": userName, "blockInfo": blockInfo})
}

// getValueForKey returns the value for the given address and key
func (ag *addressGroup) getValueForKey(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetValueForKey, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetUsername, errors.ErrBadUrlParams)
		return
	}

	key := c.Param("key")
	if key == "" {
		shared.RespondWithValidationError(c, errors.ErrGetValueForKey, errors.ErrEmptyKey)
		return
	}

	value, blockInfo, err := ag.getFacade().GetValueForKey(addr, key, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetValueForKey, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"value": value, "blockInfo": blockInfo})
}

// addressGroup returns all the key-value pairs for the given address
func (ag *addressGroup) getKeyValuePairs(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetKeyValuePairs, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetKeyValuePairs, errors.ErrBadUrlParams)
		return
	}

	value, blockInfo, err := ag.getFacade().GetKeyValuePairs(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetKeyValuePairs, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"pairs": value, "blockInfo": blockInfo})
}

// getESDTBalance returns the balance for the given address and esdt token
func (ag *addressGroup) getESDTBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrBadUrlParams)
		return
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrEmptyTokenIdentifier)
		return
	}

	esdtData, blockInfo, err := ag.getFacade().GetESDTData(addr, tokenIdentifier, 0, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetESDTBalance, err)
		return
	}

	tokenData := esdtTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         esdtData.Value.String(),
		Properties:      hex.EncodeToString(esdtData.Properties),
	}

	shared.RespondWithSuccess(c, gin.H{"tokenData": tokenData, "blockInfo": blockInfo})
}

// getESDTsRoles returns the token identifiers and roles for a given address
func (ag *addressGroup) getESDTsRoles(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetRolesForAccount, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetRolesForAccount, errors.ErrBadUrlParams)
		return
	}

	tokensRoles, blockInfo, err := ag.getFacade().GetESDTsRoles(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetRolesForAccount, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"roles": tokensRoles, "blockInfo": blockInfo})
}

// getESDTTokensWithRole returns the token identifiers where a given address has the given role
func (ag *addressGroup) getESDTTokensWithRole(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrBadUrlParams)
		return
	}

	role := c.Param("role")
	if role == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrEmptyRole)
		return
	}

	if !core.IsValidESDTRole(role) {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, fmt.Errorf("invalid role: %s", role))
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetESDTsWithRole(addr, role, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetESDTBalance, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"tokens": tokens, "blockInfo": blockInfo})
}

// getNFTTokenIDsRegisteredByAddress returns the token identifiers of the tokens where a given address is the owner
func (ag *addressGroup) getNFTTokenIDsRegisteredByAddress(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetESDTBalance, errors.ErrBadUrlParams)
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetNFTTokenIDsRegisteredByAddress(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetESDTBalance, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"tokens": tokens, "blockInfo": blockInfo})
}

// getESDTNFTData returns the nft data for the given token
func (ag *addressGroup) getESDTNFTData(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrBadUrlParams)
		return
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrEmptyTokenIdentifier)
		return
	}

	nonceAsStr := c.Param("nonce")
	if nonceAsStr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrNonceInvalid)
		return
	}

	nonceAsBigInt, okConvert := big.NewInt(0).SetString(nonceAsStr, 10)
	if !okConvert {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrNonceInvalid)
		return
	}

	esdtData, blockInfo, err := ag.getFacade().GetESDTData(addr, tokenIdentifier, nonceAsBigInt.Uint64(), options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetESDTNFTData, err)
		return
	}

	tokenData := buildTokenDataApiResponse(tokenIdentifier, esdtData)
	shared.RespondWithSuccess(c, gin.H{"tokenData": tokenData, "blockInfo": blockInfo})
}

// getAllESDTData returns the tokens list from this account
func (ag *addressGroup) getAllESDTData(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetESDTNFTData, errors.ErrBadUrlParams)
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetAllESDTTokens(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetESDTNFTData, err)
		return
	}

	formattedTokens := make(map[string]*esdtNFTTokenData)
	for tokenID, esdtData := range tokens {
		tokenData := buildTokenDataApiResponse(tokenID, esdtData)

		formattedTokens[tokenID] = tokenData
	}

	shared.RespondWithSuccess(c, gin.H{"esdts": formattedTokens, "blockInfo": blockInfo})
}

func buildTokenDataApiResponse(tokenIdentifier string, esdtData *esdt.ESDigitalToken) *esdtNFTTokenData {
	tokenData := &esdtNFTTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         esdtData.Value.String(),
		Properties:      hex.EncodeToString(esdtData.Properties),
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

func extractAccountQueryOptions(c *gin.Context) (api.AccountQueryOptions, error) {
	options, err := parseAccountQueryOptions(c)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	err = checkAccountQueryOptions(options)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	return options, nil
}

func parseAccountQueryOptions(c *gin.Context) (api.AccountQueryOptions, error) {
	onFinalBlock, err := parseBoolUrlParam(c, urlParamOnFinalBlock)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	onStartOfEpoch, err := parseUintUrlParam(c, urlParamOnStartOfEpoch)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	options := api.AccountQueryOptions{OnFinalBlock: onFinalBlock, OnStartOfEpoch: uint32(onStartOfEpoch)}
	return options, nil
}

// Question for review: should we move this validation logic in elrond-go-core, within the definition of AccountQueryOptions?
func checkAccountQueryOptions(options api.AccountQueryOptions) error {
	numSpecifiedBlockCoordinates := 0

	if options.BlockNonce > 0 {
		numSpecifiedBlockCoordinates++
	}
	if len(options.BlockHash) > 0 {
		numSpecifiedBlockCoordinates++
	}
	if len(options.BlockRootHash) > 0 {
		numSpecifiedBlockCoordinates++
	}

	if numSpecifiedBlockCoordinates > 1 {
		return fmt.Errorf("%w: only one block coordinate (blockNonce vs. blockHash vs. blockRootHash) can be specified at a time", errors.ErrBadAccountQueryOptions)
	}
	if options.OnFinalBlock && numSpecifiedBlockCoordinates > 0 {
		return fmt.Errorf("%w: onFinalBlock is not compatible with any other block coordinates", errors.ErrBadAccountQueryOptions)
	}
	if options.OnStartOfEpoch > 0 && numSpecifiedBlockCoordinates > 0 {
		return fmt.Errorf("%w: onStartOfEpoch is not compatible with any other block coordinates", errors.ErrBadAccountQueryOptions)
	}

	return nil
}
