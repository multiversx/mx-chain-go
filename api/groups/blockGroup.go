package groups

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/api/shared/logging"
)

const (
	getBlockByNoncePath       = "/by-nonce/:nonce"
	getBlockByHashPath        = "/by-hash/:hash"
	getBlockByRoundPath       = "/by-round/:round"
	getAlteredAccountsByNonce = "/altered-accounts/by-nonce/:nonce"
	getAlteredAccountsByHash  = "/altered-accounts/by-hash/:hash"
	urlParamTokensFilter      = "tokens"
	urlParamWithTxs           = "withTxs"
	urlParamWithLogs          = "withLogs"
)

// blockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type blockFacadeHandler interface {
	GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error)
	IsInterfaceNil() bool
}

type blockGroup struct {
	*baseGroup
	facade    blockFacadeHandler
	mutFacade sync.RWMutex
}

// NewBlockGroup returns a new instance of blockGroup
func NewBlockGroup(facade blockFacadeHandler) (*blockGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for block group", errors.ErrNilFacadeHandler)
	}

	bg := &blockGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: bg.getBlockByNonce,
		},
		{
			Path:    getBlockByHashPath,
			Method:  http.MethodGet,
			Handler: bg.getBlockByHash,
		},
		{
			Path:    getBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: bg.getBlockByRound,
		},
		{
			Path:    getAlteredAccountsByNonce,
			Method:  http.MethodGet,
			Handler: bg.getAlteredAccountsByNonce,
		},
		{
			Path:    getAlteredAccountsByHash,
			Method:  http.MethodGet,
			Handler: bg.getAlteredAccountsByHash,
		},
	}
	bg.endpoints = endpoints

	return bg, nil
}

func (bg *blockGroup) getBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockNonce)
		return
	}

	options, err := parseBlockQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrBadUrlParams)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetBlockByNonce(nonce, options)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetBlockByNonce")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (bg *blockGroup) getBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	options, err := parseBlockQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrBadUrlParams)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetBlockByHash(hash, options)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetBlockByHash")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (bg *blockGroup) getBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockRound)
		return
	}

	options, err := parseBlockQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrBadUrlParams)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetBlockByRound(round, options)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetBlockByRound")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (bg *blockGroup) getAlteredAccountsByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetAlteredAccountsForBlock, errors.ErrInvalidBlockRound)
		return
	}

	options := parseAlteredAccountsForBlockQueryOptionsWithoutRequestType(c)

	options.GetBlockParameters = api.GetBlockParameters{
		RequestType: api.BlockFetchTypeByNonce,
		Nonce:       nonce,
	}

	start := time.Now()
	alteredAccountsResponse, err := bg.getFacade().GetAlteredAccountsForBlock(options)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetAlteredAccountsForBlock by nonce")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetAlteredAccountsForBlock, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"accounts": alteredAccountsResponse})
}

func (bg *blockGroup) getAlteredAccountsByHash(c *gin.Context) {
	hash, err := getQueryParamHash(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetAlteredAccountsForBlock, err)
		return
	}

	options := parseAlteredAccountsForBlockQueryOptionsWithoutRequestType(c)

	options.GetBlockParameters = api.GetBlockParameters{
		RequestType: api.BlockFetchTypeByHash,
		Hash:        hash,
	}

	start := time.Now()
	alteredAccountsResponse, err := bg.getFacade().GetAlteredAccountsForBlock(options)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetAlteredAccountsForBlock by hash")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetAlteredAccountsForBlock, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"accounts": alteredAccountsResponse})
}

func parseBlockQueryOptions(c *gin.Context) (api.BlockQueryOptions, error) {
	withTxs, err := parseBoolUrlParam(c, urlParamWithTxs)
	if err != nil {
		return api.BlockQueryOptions{}, err
	}

	withLogs, err := parseBoolUrlParam(c, urlParamWithLogs)
	if err != nil {
		return api.BlockQueryOptions{}, err
	}

	options := api.BlockQueryOptions{WithTransactions: withTxs, WithLogs: withLogs}
	return options, nil
}

func parseAlteredAccountsForBlockQueryOptionsWithoutRequestType(c *gin.Context) api.GetAlteredAccountsForBlockOptions {
	tokensFilter := c.Request.URL.Query().Get(urlParamTokensFilter)

	return api.GetAlteredAccountsForBlockOptions{
		TokensFilter: tokensFilter,
	}
}

func getQueryParamNonce(c *gin.Context) (uint64, error) {
	nonceStr := c.Param("nonce")
	return strconv.ParseUint(nonceStr, 10, 64)
}

func getQueryParamHash(c *gin.Context) ([]byte, error) {
	hash := c.Param("hash")

	return hex.DecodeString(hash)
}

func getQueryParamRound(c *gin.Context) (uint64, error) {
	roundStr := c.Param("round")
	return strconv.ParseUint(roundStr, 10, 64)
}

func (bg *blockGroup) getFacade() blockFacadeHandler {
	bg.mutFacade.RLock()
	defer bg.mutFacade.RUnlock()

	return bg.facade
}

// UpdateFacade will update the facade
func (bg *blockGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(blockFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	bg.mutFacade.Lock()
	bg.facade = castFacade
	bg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bg *blockGroup) IsInterfaceNil() bool {
	return bg == nil
}
