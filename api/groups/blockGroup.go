package groups

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getBlockByNoncePath = "/by-nonce/:nonce"
	getBlockByHashPath  = "/by-hash/:hash"
)

// blockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type blockFacadeHandler interface {
	GetBlockByHash(hash string, withTxs bool) (*api.Block, error)
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
}

type blockGroup struct {
	facade    blockFacadeHandler
	mutFacade sync.RWMutex
	*baseGroup
}

// NewBlockGroup returns a new instance of blockGroup
func NewBlockGroup(facadeHandler interface{}) (*blockGroup, error) {
	if facadeHandler == nil {
		return nil, errors.ErrNilFacadeHandler
	}

	facade, ok := facadeHandler.(blockFacadeHandler)
	if !ok {
		return nil, fmt.Errorf("%w for block group", errors.ErrFacadeWrongTypeAssertion)
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
	}
	bg.endpoints = endpoints

	return bg, nil
}

func (bg *blockGroup) getBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetBlockByNonce(nonce, withTxs)
	log.Debug(fmt.Sprintf("GetBlockByNonce took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

}

func (bg *blockGroup) getBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetBlockByHash(hash, withTxs)
	log.Debug(fmt.Sprintf("GetBlockByHash took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func getQueryParamWithTxs(c *gin.Context) (bool, error) {
	withTxsStr := c.Request.URL.Query().Get("withTxs")
	if withTxsStr == "" {
		return false, nil
	}

	return strconv.ParseBool(withTxsStr)
}

func getQueryParamNonce(c *gin.Context) (uint64, error) {
	nonceStr := c.Param("nonce")
	if nonceStr == "" {
		return 0, errors.ErrInvalidBlockNonce
	}

	return strconv.ParseUint(nonceStr, 10, 64)
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
	castedFacade, ok := newFacade.(blockFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	bg.mutFacade.Lock()
	bg.facade = castedFacade
	bg.mutFacade.Unlock()

	return nil
}
